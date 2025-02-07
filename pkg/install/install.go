package install

import (
	"fmt"
	"io/fs"

	"go-install-kubernetes/pkg/config"
)

func Kubernetes(cfg *config.Config, manifestFiles fs.FS) error {
	// Log the configuration
	if cfg.IsVerbose {
		fmt.Printf("Configuration:\n"+
			"Control Node: %v\n"+
			"Worker Node: %v\n"+
			"Single Node: %v\n"+
			"Log File: %s\n",
			cfg.IsControlNode, cfg.IsWorkerNode,
			cfg.IsSingleNode, cfg.LogFile)
	}

	steps := []struct {
		name string
		fn   func(*config.Config) error
	}{
		{"Check Ubuntu version", checkUbuntuVersion},
		{"Disable swap", disableSwap},
		{"Remove existing packages", removePackages},
		{"Install required packages", installPackages},
		{"Install containerd", installContainerd},
		{"Install Kubernetes packages", installKubernetesPackages},
		{"Configure system", configureSystem},
		{"Configure crictl", configureCrictl},
		{"Configure kubelet", configureKubelet},
		{"Configure containerd", configureContainerd},
		{"Start services", startServices},
	}

	for _, step := range steps {
		fmt.Printf("Executing: %s...\n", step.name)
		if err := step.fn(cfg); err != nil {
			return fmt.Errorf("%s failed: %v", step.name, err)
		}
	}

	// Control plane specific steps
	if cfg.IsControlNode || cfg.IsSingleNode {
		controlPlaneSteps := []struct {
			name string
			fn   func(*config.Config, fs.FS) error
		}{
			{"Initialize control plane", func(cfg *config.Config, _ fs.FS) error { return kubeadmInit(cfg) }},
			{"Configure kubeconfig", func(cfg *config.Config, _ fs.FS) error { return configureKubeconfig(cfg) }},
			{"Install Calico CNI", installCalicoCNI},
			{"Wait for nodes", func(cfg *config.Config, _ fs.FS) error { return waitForNodes(cfg) }},
			{"Test Kubernetes version", func(cfg *config.Config, _ fs.FS) error { return testKubernetesVersion(cfg) }},
			{"Install metrics server", installMetricsServer},
		}

		for _, step := range controlPlaneSteps {
			fmt.Printf("Executing: %s...\n", step.name)
			if err := step.fn(cfg, manifestFiles); err != nil {
				return fmt.Errorf("%s failed: %v", step.name, err)
			}
		}

		if cfg.IsSingleNode {
			singleNodeSteps := []struct {
				name string
				fn   func(*config.Config) error
			}{
				{"Configure as single node", configureAsSingleNode},
				{"Test nginx pod", testNginxPod},
				{"Wait for pods running", waitForPodsRunning},
			}

			for _, step := range singleNodeSteps {
				fmt.Printf("Executing: %s...\n", step.name)
				if err := step.fn(cfg); err != nil {
					return fmt.Errorf("%s failed: %v", step.name, err)
				}
			}
		}
	} else {
		if err := checkWorkerServices(cfg); err != nil {
			return fmt.Errorf("worker services check failed: %v", err)
		}
	}

	return nil
}
