package domainapp

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nobuenhombre/suikat/pkg/ge"
)

//go:embed templates/*.tpl
var templateFS embed.FS

// TemplatesDir extracts embedded templates to a temp directory and returns the path.
// The caller should not remove the directory — it lives for the process lifetime.
func TemplatesDir() (string, error) {
	dir, err := os.MkdirTemp("", "goxogen-templates-*")
	if err != nil {
		return "", ge.Pin(fmt.Errorf("creating temp dir for templates: %w", err))
	}

	entries, err := templateFS.ReadDir("templates")
	if err != nil {
		return "", ge.Pin(fmt.Errorf("reading embedded templates dir: %w", err))
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := templateFS.ReadFile(filepath.Join("templates", entry.Name()))
		if err != nil {
			return "", ge.Pin(fmt.Errorf("reading embedded template %s: %w", entry.Name(), err))
		}

		outPath := filepath.Join(dir, entry.Name())
		err = os.WriteFile(outPath, data, 0644)
		if err != nil {
			return "", ge.Pin(fmt.Errorf("writing template %s: %w", entry.Name(), err))
		}
	}

	return dir, nil
}
