package install

import (
	"fmt"
	"os"
	"strings"

	"go-install-kubernetes/pkg/config"
	"go-install-kubernetes/pkg/exec"

	"github.com/bitfield/script"
)

func checkUbuntuVersion(cfg *config.Config) error {
	version, err := script.File("/etc/lsb-release").Match("DISTRIB_RELEASE").String()
	if err != nil {
		return err
	}
	if !strings.Contains(version, config.UbuntuVersion) {
		return fmt.Errorf("this script only works on Ubuntu %s", config.UbuntuVersion)
	}
	return nil
}

func disableSwap(cfg *config.Config) error {
	if _, err := exec.Command("swapoff -a", cfg); err != nil {
		return err
	}
	_, err := script.File("/etc/fstab").
		FilterLine(func(line string) string {
			if strings.Contains(line, " swap ") {
				return "#" + line
			}
			return line
		}).
		WriteFile("/etc/fstab")
	return err
}

func removePackages(cfg *config.Config) error {
	cmds := []string{
		"apt-mark unhold kubelet kubeadm kubectl kubernetes-cni",
		"apt-get remove -y moby-buildx moby-cli moby-compose moby-containerd moby-engine moby-runc",
		"apt-get autoremove -y",
		"apt-get remove -y docker.io containerd kubelet kubeadm kubectl",
		"systemctl daemon-reload",
	}
	for _, cmd := range cmds {
		if _, err := exec.Command(cmd, cfg); err != nil {
			// Ignore errors as some packages might not exist
			continue
		}
	}
	return nil
}

func installPackages(cfg *config.Config) error {
	if _, err := exec.Command("apt-get update", cfg); err != nil {
		return err
	}

	packages := []string{
		"apt-transport-https",
		"ca-certificates",
		"curl",
		"gnupg",
		"lsb-release",
		"software-properties-common",
		"wget",
		"jq",
	}
	installCmd := fmt.Sprintf("apt-get install -y %s", strings.Join(packages, " "))
	_, err := exec.Command(installCmd, cfg)
	return err
}

func installContainerd(cfg *config.Config) error {
	cmds := []string{
		"apt-get update",
		"apt-get install -y containerd",
	}
	for _, cmd := range cmds {
		if _, err := exec.Command(cmd, cfg); err != nil {
			return err
		}
	}
	return nil
}

func installKubernetesPackages(cfg *config.Config) error {
	// Extract major version (1.29 from 1.29.0)
	kubeRepoVersion := strings.Join(strings.Split(config.KubeVersion, ".")[:2], ".")

	// Remove old repo file and GPG key if they exist
	os.Remove("/etc/apt/sources.list.d/kubernetes.list")
	os.Remove("/etc/apt/keyrings/kubernetes-apt-keyring.gpg")

	// Create keyrings directory if it doesn't exist
	if err := os.MkdirAll("/etc/apt/keyrings", 0755); err != nil {
		return err
	}

	// Download and install GPG key
	gpgKeyURL := fmt.Sprintf("https://pkgs.k8s.io/core:/stable:/v%s/deb/Release.key", kubeRepoVersion)
	if _, err := exec.Command(fmt.Sprintf("curl -fsSLo /tmp/k8s-key.gpg %s", gpgKeyURL), cfg); err != nil {
		return err
	}

	if _, err := exec.Command("gpg --dearmor --yes -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg /tmp/k8s-key.gpg", cfg); err != nil {
		return err
	}

	// Clean up temp file
	os.Remove("/tmp/k8s-key.gpg")

	// Add new repo
	repoContent := fmt.Sprintf("deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v%s/deb/ /", kubeRepoVersion)
	if err := os.WriteFile("/etc/apt/sources.list.d/kubernetes.list", []byte(repoContent), 0644); err != nil {
		return err
	}

	cmds := []string{
		"apt-get update",
		fmt.Sprintf("apt-get install -y --allow-downgrades kubelet=%s-* kubeadm=%s-* kubectl=%s-*", config.KubeVersion, config.KubeVersion, config.KubeVersion),
		"apt-mark hold kubelet kubeadm kubectl",
	}
	for _, cmd := range cmds {
		if _, err := exec.Command(cmd, cfg); err != nil {
			return err
		}
	}
	return nil
}

func configureSystem(cfg *config.Config) error {
	modulesContent := "overlay\nbr_netfilter\n"
	if err := os.WriteFile("/etc/modules-load.d/containerd.conf", []byte(modulesContent), 0644); err != nil {
		return err
	}

	sysctlContent := `net.bridge.bridge-nf-call-iptables  = 1
net.ipv4.ip_forward                 = 1
net.bridge.bridge-nf-call-ip6tables = 1`
	if err := os.WriteFile("/etc/sysctl.d/99-kubernetes-cri.conf", []byte(sysctlContent), 0644); err != nil {
		return err
	}

	cmds := []string{
		"modprobe overlay",
		"modprobe br_netfilter",
		"sysctl --system",
	}
	for _, cmd := range cmds {
		if _, err := exec.Command(cmd, cfg); err != nil {
			return err
		}
	}
	return nil
}

func configureCrictl(cfg *config.Config) error {
	content := "runtime-endpoint: unix:///run/containerd/containerd.sock\n"
	return os.WriteFile("/etc/crictl.yaml", []byte(content), 0644)
}

func configureKubelet(cfg *config.Config) error {
	content := "KUBELET_EXTRA_ARGS=\"--container-runtime-endpoint unix:///run/containerd/containerd.sock\"\n"
	return os.WriteFile("/etc/default/kubelet", []byte(content), 0644)
}

func configureContainerd(cfg *config.Config) error {
	if err := os.MkdirAll("/etc/containerd", 0755); err != nil {
		return err
	}

	configContent := `disabled_plugins = []
imports = []
oom_score = 0
plugin_dir = ""
required_plugins = []
root = "/var/lib/containerd"
state = "/run/containerd"
version = 2

[plugins]

  [plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
    [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
      base_runtime_spec = ""
      container_annotations = []
      pod_annotations = [] 
      privileged_without_host_devices = false
      runtime_engine = ""
      runtime_root = ""
      runtime_type = "io.containerd.runc.v2"

      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
        BinaryName = ""
        CriuImagePath = ""
        CriuPath = ""
        CriuWorkPath = ""
        IoGid = 0
        IoUid = 0
        NoNewKeyring = false
        NoPivotRoot = false
        Root = ""
        ShimCgroup = ""
        SystemdCgroup = true`

	return os.WriteFile("/etc/containerd/config.toml", []byte(configContent), 0644)
}

func startServices(cfg *config.Config) error {
	cmds := []string{
		"systemctl daemon-reload",
		"systemctl enable containerd",
		"systemctl restart containerd",
		"systemctl enable kubelet",
		"systemctl start kubelet",
	}
	for _, cmd := range cmds {
		if _, err := exec.Command(cmd, cfg); err != nil {
			return err
		}
	}
	return nil
}
