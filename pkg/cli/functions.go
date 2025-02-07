package cli

import (
	"fmt"

	"go-install-kubernetes/pkg/config"
)

func showHelp() {
	fmt.Println("USAGE:")
	fmt.Println("  go-install-kubernetes [options]")
	fmt.Println("\nOPTIONS:")
	fmt.Println("  -c  Configure as a control plane node")
	fmt.Println("  -w  Configure as a worker node")
	fmt.Println("  -s  Configure as a single node (control plane + worker)")
	fmt.Println("  -v  Enable verbose output")
	fmt.Println("  -h  Show this help message")
	fmt.Println("  --version  Show version information")
	fmt.Println("  --export-manifests  Export embedded Calico manifests to disk")
	fmt.Println("\nAt least one of -c, -w, or -s must be specified")
}

func printVersion() {
	fmt.Printf("Install Kubernetes Version: %s\n", config.CLIVersion)
	fmt.Printf("Kubernetes Version: %s\n", config.KubeVersion)
	fmt.Printf("Containerd Version: %s\n", config.ContainerdVersion)
	fmt.Printf("Calico Version: %s\n", config.CalicoVersion)
	fmt.Printf("Ubuntu Version: %s\n", config.UbuntuVersion)
}
