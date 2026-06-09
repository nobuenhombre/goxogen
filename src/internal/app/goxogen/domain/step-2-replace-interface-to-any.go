package domainapp

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nobuenhombre/suikat/pkg/ge"
)

// replaceInterfaceToAny replaces interface{} with any and xo-specific type aliases
// with their standard library equivalents in .go files.
func (d *AppDomain) replaceInterfaceToAny(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return ge.Pin(err)
	}

	sort.Strings(files)

	replacements := map[string]string{
		"interface{}":               "any",
		"Timestamp0WithoutTimeZone": "sql.NullTime",
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return ge.Pin(fmt.Errorf("reading %s: %w", file, err))
		}

		content := string(data)
		newContent := content
		for old, new := range replacements {
			newContent = strings.ReplaceAll(newContent, old, new)
		}

		if content != newContent {
			err = os.WriteFile(file, []byte(newContent), 0644)
			if err != nil {
				return ge.Pin(fmt.Errorf("writing %s: %w", file, err))
			}
			log.Printf("[xo] Replaced type aliases in %s", filepath.Base(file))
		}
	}

	return nil
}
