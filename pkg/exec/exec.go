package exec

import (
	"fmt"
	"os"

	"go-install-kubernetes/pkg/config"

	"github.com/bitfield/script"
)

func Command(cmd string, cfg *config.Config) (string, error) {
	pipe := script.Exec(cmd)
	output, err := pipe.String()
	if err != nil {
		return "", err
	}

	// Always append to log file
	f, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Write command and output to log file
	if _, err := fmt.Fprintf(f, "\n$ %s\n%s\n", cmd, output); err != nil {
		return "", err
	}

	// If verbose, also print to stdout
	if cfg.IsVerbose {
		fmt.Printf("$ %s\n%s\n", cmd, output)
	}

	return output, nil
}
