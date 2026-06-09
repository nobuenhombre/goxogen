package domainapp

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nobuenhombre/suikat/pkg/ge"
)

// goFormatCode removes CRUD markers, then runs go fmt, goimports, and go vet.
// Errors from fmt/goimports/vet are collected and displayed like gobp.
func (d *AppDomain) goFormatCode(path string) error {
	// Step 1: collect .go files
	goFiles, err := filepath.Glob(filepath.Join(path, "*.go"))
	if err != nil {
		return ge.Pin(fmt.Errorf("listing go files: %w", err))
	}

	// Step 2: clean CRUD markers — must run before go fmt so formatting
	// doesn't touch lines that will be removed
	if len(goFiles) > 0 {
		log.Printf("[xo] Cleaning CRUD markers from %d files", len(goFiles))
		if err := d.cleanCRUDMarkers(goFiles); err != nil {
			return ge.Pin(fmt.Errorf("cleaning CRUD markers: %w", err))
		}
	}

	// Step 3: go fmt
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

	// Step 4: goimports
	if len(goFiles) > 0 {
		log.Printf("[xo] Running goimports -w on %s", path)

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

	// Step 5: go vet
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

// cleanCRUDMarkers removes lines containing `@crud` or `@end-crud` markers
// from the provided .go files. This ensures generated code doesn't contain
// development-stage scaffolding markers.
func (d *AppDomain) cleanCRUDMarkers(files []string) error {
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return ge.Pin(fmt.Errorf("reading %s: %w", f, err))
		}

		lines := strings.Split(string(data), "\n")
		var cleaned []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.Contains(trimmed, "@crud") || strings.Contains(trimmed, "@end-crud") {
				continue
			}
			cleaned = append(cleaned, line)
		}

		result := strings.Join(cleaned, "\n")
		if err := os.WriteFile(f, []byte(result), 0644); err != nil {
			return ge.Pin(fmt.Errorf("writing %s: %w", f, err))
		}
	}

	return nil
}
