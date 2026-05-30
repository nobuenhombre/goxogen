package domainapp

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// cleanXoXouidSourceBlocks removes @repo-start/@repo-end blocks from .xo.go and .xouid.go files.
func (d *AppDomain) cleanXoXouidSourceBlocks(dir string) error {
	// Process .xo.go files
	if err := d.removeRepoBlocksFromFiles(dir, "*.xo.go", "*-repo.xo.go"); err != nil {
		return err
	}

	// Process .xouid.go files
	if err := d.removeRepoBlocksFromFiles(dir, "*.xouid.go", "*-repo.xo.go"); err != nil {
		return err
	}

	return nil
}

// removeRepoBlocksFromFiles removes @repo-start/@repo-end blocks from matching files.
func (d *AppDomain) removeRepoBlocksFromFiles(dir, pattern, excludePattern string) error {
	files, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return err
	}

	sort.Strings(files)

	for _, file := range files {
		// Skip -repo.xo.go files
		if strings.HasSuffix(file, "-repo.xo.go") {
			continue
		}

		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("reading %s: %w", file, err)
		}

		content := string(data)
		if !strings.Contains(content, "// @repo-start") {
			continue
		}

		// Remove lines between @repo-start and @repo-end (including markers)
		var result strings.Builder
		skip := false
		for _, line := range strings.Split(content, "\n") {
			if strings.Contains(line, "// @repo-start") {
				skip = true
				continue
			}
			if strings.Contains(line, "// @repo-end") {
				skip = false
				continue
			}
			if !skip {
				result.WriteString(line + "\n")
			}
		}

		output := result.String()
		// Remove trailing newline that we added
		if strings.HasSuffix(output, "\n") {
			output = strings.TrimSuffix(output, "\n")
		}

		if err := os.WriteFile(file, []byte(output), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", file, err)
		}

		log.Printf("[xo] Cleaned repo blocks from %s", filepath.Base(file))
	}

	return nil
}
