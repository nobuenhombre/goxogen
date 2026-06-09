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

const (
	crudBlockStart = "// @crud"
	crudBlockEnd   = "// @end-crud"
)

// removeCRUDBlocks strips all lines between // @crud and // @end-crud markers
// in generated .xo.go files. Used when db_is_readonly=true.
func (d *AppDomain) removeCRUDBlocks(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.xo.go"))
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
		newContent := stripCRUDBlocks(content)

		if content != newContent {
			err = os.WriteFile(file, []byte(newContent), 0644)
			if err != nil {
				return ge.Pin(fmt.Errorf("writing %s: %w", file, err))
			}
			log.Printf("[xo] Stripped CRUD blocks from %s (readonly mode)", filepath.Base(file))
		}
	}

	return nil
}

// stripCRUDBlocks removes all content between // @crud and // @end-crud markers
// (including the markers themselves). Non-overlapping, left-to-right scan.
func stripCRUDBlocks(content string) string {
	var result strings.Builder
	remaining := content

	for {
		start := strings.Index(remaining, crudBlockStart)
		if start == -1 {
			result.WriteString(remaining)
			break
		}

		// Write everything before the marker
		result.WriteString(remaining[:start])

		// Find the end marker
		end := strings.Index(remaining[start+len(crudBlockStart):], crudBlockEnd)
		if end == -1 {
			// No closing marker — keep the rest as-is
			result.WriteString(remaining[start:])
			break
		}

		// Skip everything from start marker through end marker (inclusive)
		skipTo := start + len(crudBlockStart) + end + len(crudBlockEnd)
		remaining = remaining[skipTo:]
	}

	return result.String()
}
