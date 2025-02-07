package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"go-install-kubernetes/pkg/cli"
	"go-install-kubernetes/pkg/install"
)

//go:embed manifests/*
//go:embed manifests/calico/*
var manifestFiles embed.FS

func main() {
	config := cli.ParseFlags(manifestFiles)

	// Allow -h without root check
	if len(os.Args) == 2 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		return
	}

	// Check if running as root
	if os.Geteuid() != 0 {
		log.Fatal("This script must be run as root")
	}

	// Create temp directory for logs
	tmpDir, err := os.MkdirTemp("", "install-kubernetes-*")
	if err != nil {
		log.Fatal(err)
	}
	config.LogFile = filepath.Join(tmpDir, "install.log")

	// Create the log file
	if _, err := os.Create(config.LogFile); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Writing all output to: %s\n", config.LogFile)

	defer os.RemoveAll(tmpDir)

	if err := install.Kubernetes(config, manifestFiles); err != nil {
		// Print log file on error
		fmt.Println("\n### Error Log ###")
		content, _ := os.ReadFile(config.LogFile)
		fmt.Println(string(content))
		log.Fatal(err)
	}

	// Print log file if verbose
	if config.IsVerbose {
		fmt.Println("\n### Log file ###")
		content, _ := os.ReadFile(config.LogFile)
		fmt.Println(string(content))
	}

	// Print join command for control plane
	if config.IsControlNode || config.IsSingleNode {
		fmt.Println("\n### Command to add a worker node ###")
		output, err := exec.Command("kubeadm", "token", "create", "--print-join-command", "--ttl", "0").Output()
		if err != nil {
			log.Printf("Failed to create join token: %v", err)
		} else {
			fmt.Println(string(output))
		}
	} else {
		fmt.Println("\n### To add this node as a worker node ###")
		fmt.Println("Run the below on the control plane node:")
		fmt.Println("kubeadm token create --print-join-command --ttl 0")
		fmt.Println("and execute the output on the worker nodes")
	}
}
