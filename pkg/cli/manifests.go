package cli

import (
	"fmt"
	"io/fs"
	"os"

	"go-install-kubernetes/pkg/config"
)

func exportEmbeddedFiles(cfg *config.Config, manifestFiles fs.FS) error {
	// Create manifests directory
	if err := os.MkdirAll("manifests", 0755); err != nil {
		return fmt.Errorf("failed to create manifests directory: %v", err)
	}

	// Walk through all embedded files
	err := fs.WalkDir(manifestFiles, "manifests", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, we'll create them when needed
		if d.IsDir() {
			return os.MkdirAll(path, 0755)
		}

		// Read and write each file
		content, err := fs.ReadFile(manifestFiles, path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %v", path, err)
		}

		if err := os.WriteFile(path, content, 0644); err != nil {
			return fmt.Errorf("failed to export file %s: %v", path, err)
		}

		fmt.Printf("Exported: %s\n", path)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to export manifests: %v", err)
	}

	fmt.Println("Successfully exported all manifests")
	return nil
}
