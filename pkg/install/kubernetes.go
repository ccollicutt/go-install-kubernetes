package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-install-kubernetes/pkg/config"
	"go-install-kubernetes/pkg/exec"
	"io/fs"
)

func kubeadmInit(cfg *config.Config) error {
	output, err := exec.Command("ip route get 1", cfg)
	if err != nil {
		return err
	}

	// Parse the IP address from the output
	fields := strings.Fields(output)
	var mainIP string
	for i, field := range fields {
		if field == "src" && i+1 < len(fields) {
			mainIP = fields[i+1]
			break
		}
	}

	if mainIP == "" {
		return fmt.Errorf("could not determine main IP address")
	}

	// Create secure temporary directory
	tmpDir, err := os.MkdirTemp("", "kubeadm-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set secure permissions
	if err := os.Chmod(tmpDir, 0700); err != nil {
		return fmt.Errorf("failed to set permissions on temp directory: %v", err)
	}

	configContent := fmt.Sprintf(`apiVersion: kubeadm.k8s.io/v1beta3
kind: ClusterConfiguration
kubernetesVersion: v%s
networking:
  podSubnet: 192.168.0.0/16
controlPlaneEndpoint: "%s:6443"`, config.KubeVersion, mainIP)

	configPath := filepath.Join(tmpDir, "kubeadm-config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		return fmt.Errorf("failed to write kubeadm config: %v", err)
	}

	_, err = exec.Command(fmt.Sprintf("kubeadm init --config %s", configPath), cfg)
	return err
}

func configureKubeconfig(cfg *config.Config) error {
	cmds := []string{
		"mkdir -p /root/.kube",
		"cp -i /etc/kubernetes/admin.conf /root/.kube/config",
		"mkdir -p /home/ubuntu/.kube",
		"cp -i /etc/kubernetes/admin.conf /home/ubuntu/.kube/config",
		"chown ubuntu:ubuntu /home/ubuntu/.kube/config",
	}
	for _, cmd := range cmds {
		if _, err := exec.Command(cmd, cfg); err != nil {
			// Ignore errors for ubuntu user operations
			continue
		}
	}
	return nil
}

func installCalicoCNI(cfg *config.Config, manifestFiles fs.FS) error {
	// Create temporary directory for manifest files
	tmpDir, err := os.MkdirTemp("", "calico-manifests-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Extract and apply tigera-operator
	operatorContent, err := fs.ReadFile(manifestFiles, "manifests/calico/tigera-operator.yaml")
	if err != nil {
		return fmt.Errorf("failed to read tigera-operator manifest: %v", err)
	}

	operatorFile := filepath.Join(tmpDir, "tigera-operator.yaml")
	if err := os.WriteFile(operatorFile, operatorContent, 0644); err != nil {
		return fmt.Errorf("failed to write tigera-operator manifest: %v", err)
	}

	// Apply the operator manifest
	if _, err := exec.Command(fmt.Sprintf("kubectl create -f %s", operatorFile), cfg); err != nil {
		return fmt.Errorf("failed to apply tigera-operator: %v", err)
	}

	// Add a delay to allow CRDs to be established
	fmt.Println("Waiting for Calico CRDs to be established...")
	time.Sleep(20 * time.Second)

	// Wait for specific CRDs to be established
	crds := []string{
		"installations.operator.tigera.io",
		"tigerastatuses.operator.tigera.io",
		"ippools.crd.projectcalico.org",
	}

	for _, crd := range crds {
		fmt.Printf("Waiting for CRD %s...\n", crd)
		if _, err := exec.Command(fmt.Sprintf("kubectl wait --for=condition=established --timeout=60s crd/%s", crd), cfg); err != nil {
			return fmt.Errorf("timeout waiting for CRD %s: %v", crd, err)
		}
	}

	// Extract and apply custom-resources
	customResContent, err := fs.ReadFile(manifestFiles, "manifests/calico/custom-resources.yaml")
	if err != nil {
		return fmt.Errorf("failed to read custom-resources manifest: %v", err)
	}

	customResFile := filepath.Join(tmpDir, "custom-resources.yaml")
	if err := os.WriteFile(customResFile, customResContent, 0644); err != nil {
		return fmt.Errorf("failed to write custom-resources manifest: %v", err)
	}

	// Apply the custom resources manifest
	if _, err := exec.Command(fmt.Sprintf("kubectl create -f %s", customResFile), cfg); err != nil {
		return fmt.Errorf("failed to apply custom-resources: %v", err)
	}

	// Wait for tigera-operator pod to be running
	fmt.Println("Waiting for tigera-operator pod to be ready...")
	if _, err := exec.Command(fmt.Sprintf("kubectl wait --for=condition=Ready pod -l k8s-app=tigera-operator -n tigera-operator --timeout=%s", config.KubectlTimeout), cfg); err != nil {
		return fmt.Errorf("timeout waiting for tigera-operator: %v", err)
	}

	// Wait for Calico installation to be ready
	fmt.Println("Waiting for Calico installation to be ready...")
	if _, err := exec.Command("kubectl wait --for=condition=Ready installation.operator.tigera.io/default --timeout=300s", cfg); err != nil {
		return fmt.Errorf("timeout waiting for Calico installation: %v", err)
	}

	// Wait for calico-node pods
	fmt.Println("Waiting for calico-node pods to be ready...")
	if _, err := exec.Command("kubectl wait --for=condition=Ready pod -l k8s-app=calico-node -n calico-system --timeout=300s", cfg); err != nil {
		// If the first attempt fails, check if the namespace exists
		if _, err := exec.Command("kubectl get ns calico-system", cfg); err != nil {
			return fmt.Errorf("calico-system namespace not found: %v", err)
		}

		// Show pod status for debugging
		if _, err := exec.Command("kubectl get pods -n calico-system", cfg); err != nil {
			return fmt.Errorf("failed to get calico pods status: %v", err)
		}

		// Try waiting one more time with a longer timeout
		time.Sleep(30 * time.Second)
		if _, err := exec.Command("kubectl wait --for=condition=Ready pod -l k8s-app=calico-node -n calico-system --timeout=300s", cfg); err != nil {
			return fmt.Errorf("timeout waiting for calico-node pods: %v", err)
		}
	}

	return nil
}

func waitForNodes(cfg *config.Config) error {
	_, err := exec.Command(fmt.Sprintf("kubectl wait --for=condition=Ready --all nodes --timeout=%s", config.KubectlTimeout), cfg)
	return err
}

func testKubernetesVersion(cfg *config.Config) error {
	out, err := exec.Command("kubectl version -o json", cfg)
	if err != nil {
		return err
	}

	if !strings.Contains(out, fmt.Sprintf("v%s", config.KubeVersion)) {
		return fmt.Errorf("kubernetes version mismatch")
	}
	return nil
}

func configureAsSingleNode(cfg *config.Config) error {
	if _, err := exec.Command("kubectl taint nodes --all node-role.kubernetes.io/control-plane:NoSchedule-", cfg); err != nil {
		return err
	}
	time.Sleep(10 * time.Second) // Wait for taint to take effect
	return nil
}

func testNginxPod(cfg *config.Config) error {
	cmds := []string{
		"kubectl run --image nginx --namespace default nginx",
		fmt.Sprintf("kubectl wait --for=condition=Ready --all pods --namespace default --timeout=%s", config.KubectlTimeout),
		"kubectl delete pod nginx --namespace default",
	}
	for _, cmd := range cmds {
		if _, err := exec.Command(cmd, cfg); err != nil {
			return err
		}
	}
	return nil
}

func waitForPodsRunning(cfg *config.Config) error {
	timeout := time.After(5 * time.Minute)
	tick := time.Tick(10 * time.Second)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for pods to be running")
		case <-tick:
			out, err := exec.Command("kubectl get pods --all-namespaces --no-headers", cfg)
			if err != nil {
				return err
			}

			nonRunningCount := 0
			for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
				if line != "" && !strings.Contains(line, "Running") {
					nonRunningCount++
				}
			}

			if nonRunningCount == 0 {
				return nil
			}
		}
	}
}

func checkWorkerServices(cfg *config.Config) error {
	_, err := exec.Command("systemctl is-active containerd", cfg)
	return err
}

func installMetricsServer(cfg *config.Config, manifestFiles fs.FS) error {
	fmt.Println("Installing metrics server...")
	metricsContent, err := fs.ReadFile(manifestFiles, "manifests/metrics-server.yaml")
	if err != nil {
		return fmt.Errorf("failed to read metrics-server manifest: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "metrics-server-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	metricsFile := filepath.Join(tmpDir, "metrics-server.yaml")
	if err := os.WriteFile(metricsFile, metricsContent, 0644); err != nil {
		return fmt.Errorf("failed to write metrics-server manifest: %v", err)
	}

	if _, err := exec.Command(fmt.Sprintf("kubectl apply -f %s", metricsFile), cfg); err != nil {
		return fmt.Errorf("failed to apply metrics-server: %v", err)
	}

	return nil
}
