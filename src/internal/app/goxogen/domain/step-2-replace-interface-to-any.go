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

// replaceInterfaceToAny replaces all occurrences of interface{} with any in .go files.
func (d *AppDomain) replaceInterfaceToAny(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return ge.Pin(err)
	}

	sort.Strings(files)

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return ge.Pin(fmt.Errorf("reading %s: %w", file, err))
		}

		content := string(data)
		newContent := strings.ReplaceAll(content, "interface{}", "any")

		if content != newContent {
			err = os.WriteFile(file, []byte(newContent), 0644)
			if err != nil {
				return ge.Pin(fmt.Errorf("writing %s: %w", file, err))
			}
			log.Printf("[xo] Replaced interface{} in %s", filepath.Base(file))
		}
	}

	return nil
}
