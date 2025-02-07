package cli

import (
	"flag"
	"fmt"
	"io/fs"
	"os"

	"go-install-kubernetes/pkg/config"
)

func ParseFlags(manifestFiles fs.FS) *config.Config {
	cfg := &config.Config{
		IsWorkerNode: false,
	}

	flag.BoolVar(&cfg.IsControlNode, "c", false, "Configure as a control plane node")
	flag.BoolVar(&cfg.IsWorkerNode, "w", false, "Configure as a worker node")
	flag.BoolVar(&cfg.IsSingleNode, "s", false, "Configure as a single node (control plane + worker)")
	flag.BoolVar(&cfg.IsVerbose, "v", false, "Enable verbose output")
	exportManifests := flag.Bool("export-manifests", false, "Export embedded Calico manifests to disk")
	showVersion := flag.Bool("version", false, "Show version information")

	flag.Usage = showHelp
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	if *exportManifests {
		if err := exportEmbeddedFiles(cfg, manifestFiles); err != nil {
			fmt.Fprintf(os.Stderr, "Error exporting manifests: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if !cfg.IsControlNode && !cfg.IsWorkerNode && !cfg.IsSingleNode {
		showHelp()
		os.Exit(0)
	}

	if cfg.IsControlNode {
		cfg.IsWorkerNode = false
	}
	if cfg.IsSingleNode {
		cfg.IsControlNode = true
		cfg.IsWorkerNode = true
	}

	return cfg
}

// ... (move all CLI-related functions here)
