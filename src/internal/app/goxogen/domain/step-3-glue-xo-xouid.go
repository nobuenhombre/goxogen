package domainapp

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// glueXoXouid merges .xo.go and matching .xouid.go files into .xo-xouid.go files.
func (d *AppDomain) glueXoXouid(dir string) error {
	xoFiles, err := filepath.Glob(filepath.Join(dir, "*.xo.go"))
	if err != nil {
		return err
	}

	sort.Strings(xoFiles)

	for _, xoFile := range xoFiles {
		basePath := strings.TrimSuffix(xoFile, ".xo.go")
		xouidFile := basePath + ".xouid.go"
		targetFile := basePath + ".xo-xouid.go"

		xoData, err := os.ReadFile(xoFile)
		if err != nil {
			return fmt.Errorf("reading %s: %w", xoFile, err)
		}

		var combined strings.Builder
		combined.Write(xoData)
		combined.WriteString("\n")

		if _, err := os.Stat(xouidFile); err == nil {
			xouidData, err := os.ReadFile(xouidFile)
			if err != nil {
				return fmt.Errorf("reading %s: %w", xouidFile, err)
			}
			combined.Write(xouidData)
		}

		if err := os.WriteFile(targetFile, []byte(combined.String()), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", targetFile, err)
		}

		log.Printf("[xo] Glued %s", filepath.Base(targetFile))
	}

	return nil
}
