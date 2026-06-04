package domainapp

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nobuenhombre/suikat/pkg/ge"
)

// goFormatCode runs go fmt, goimports, and go vet.
// Errors from fmt/goimports/vet are collected and displayed like gobp.
func (d *AppDomain) goFormatCode(path string) error {
	// Step 1: go fmt
	log.Printf("[xo] Running go fmt %s", path)

	cmd := exec.Command("go", "fmt", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// go fmt returns non-zero on parse errors
		errMsg := strings.TrimSpace(string(output))
		if errMsg != "" {
			log.Printf("[xo] go fmt error: %s", errMsg)
		}
		return ge.Pin(fmt.Errorf("go fmt failed: %w", err))
	}
	// Print formatting changes (modified filenames) if any
	if trimmed := strings.TrimSpace(string(output)); trimmed != "" {
		log.Printf("[xo] go fmt modified: %s", trimmed)
	}

	// Step 2: goimports
	log.Printf("[xo] Running goimports -w on %s", path)

	goFiles, err := filepath.Glob(filepath.Join(path, "*.go"))
	if err != nil {
		return ge.Pin(fmt.Errorf("listing go files: %w", err))
	}

	if len(goFiles) > 0 {
		args := append([]string{"-w"}, goFiles...)
		cmd = exec.Command("goimports", args...)
		output, err = cmd.CombinedOutput()
		if err != nil {
			errMsg := strings.TrimSpace(string(output))
			// goimports may not be installed — log as warning
			if errMsg != "" {
				log.Printf("[xo] goimports error: %s", errMsg)
			}
			log.Printf("[xo] Warning: goimports failed (may not be installed): %v", err)
		} else {
			if trimmed := strings.TrimSpace(string(output)); trimmed != "" {
				log.Printf("[xo] goimports output: %s", trimmed)
			}
		}
	}

	// Step 3: go vet
	log.Printf("[xo] Running go vet %s", path)

	cmd = exec.Command("go", "vet", path)
	output, err = cmd.CombinedOutput()
	if err != nil {
		errMsg := strings.TrimSpace(string(output))
		if errMsg != "" {
			log.Printf("[xo] go vet error: %s", errMsg)
			return ge.Pin(fmt.Errorf("go vet failed: %s", errMsg))
		}
		return ge.Pin(fmt.Errorf("go vet failed: %w", err))
	}
	if trimmed := strings.TrimSpace(string(output)); trimmed != "" {
		log.Printf("[xo] go vet output: %s", trimmed)
	}

	return nil
}
