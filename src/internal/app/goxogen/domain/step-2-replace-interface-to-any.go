package domainapp

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// replaceInterfaceToAny replaces all occurrences of interface{} with any in .go files.
func (d *AppDomain) replaceInterfaceToAny(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return err
	}

	sort.Strings(files)

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("reading %s: %w", file, err)
		}

		content := string(data)
		newContent := strings.ReplaceAll(content, "interface{}", "any")

		if content != newContent {
			if err := os.WriteFile(file, []byte(newContent), 0644); err != nil {
				return fmt.Errorf("writing %s: %w", file, err)
			}
			log.Printf("[xo] Replaced interface{} in %s", filepath.Base(file))
		}
	}

	return nil
}
