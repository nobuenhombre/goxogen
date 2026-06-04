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

// extractRepo extracts @repo-start/@repo-end blocks from .xo-xouid.go files into *-repo.xo.go files.
func (d *AppDomain) extractRepo(outdir, pkg string) error {
	files, err := filepath.Glob(filepath.Join(outdir, "*.xo-xouid.go"))
	if err != nil {
		return ge.Pin(err)
	}

	sort.Strings(files)

	for _, file := range files {
		if strings.HasSuffix(file, "-repo.xo.go") {
			continue
		}

		err = d.extractRepoFile(outdir, file, pkg)
		if err != nil {
			log.Printf("[xo] Warning: %v (skipping)", err)
			continue
		}
	}

	return nil
}

// extractRepoFile processes a single file for repository extraction.
func (d *AppDomain) extractRepoFile(outdir, file, pkg string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return ge.Pin(fmt.Errorf("reading %s: %w", file, err))
	}

	content := string(data)

	// Check for markers
	if !strings.Contains(content, "// @repo-start") || !strings.Contains(content, "// @repo-end") {
		return ge.Pin(fmt.Errorf("markers not found in %s", filepath.Base(file)))
	}

	// Extract repository name
	repoName := ""
	lines := strings.Split(content, "\n")
	inBlock := false
	for _, line := range lines {
		if strings.Contains(line, "// @repo-start") {
			inBlock = true
			continue
		}
		if strings.Contains(line, "// @repo-end") {
			inBlock = false
			continue
		}
		if inBlock {
			if strings.HasPrefix(strings.TrimSpace(line), "type ") && strings.Contains(line, "Repository") {
				// Parse "type XXXRepository struct {"
				trimmed := strings.TrimSpace(line)
				trimmed = strings.TrimPrefix(trimmed, "type ")
				idx := strings.Index(trimmed, "Repository")
				if idx > 0 {
					repoName = trimmed[:idx]
				}
				break
			}
		}
	}

	if repoName == "" {
		return ge.Pin(fmt.Errorf("no repository name found in %s", filepath.Base(file)))
	}

	// Determine output file name
	baseName := strings.TrimSuffix(filepath.Base(file), ".xo-xouid.go")
	outputFile := filepath.Join(outdir, baseName+"-repo.xo.go")

	// Generate the repository file
	var buf strings.Builder

	// Header
	buf.WriteString("// Code generated from " + filepath.Base(file) + ". DO NOT EDIT.\n")
	buf.WriteString("package " + pkg + "\n\n")
	buf.WriteString("import (\n")
	buf.WriteString("\t\"time\"\n\n")
	buf.WriteString("\t\"github.com/google/uuid\"\n")
	buf.WriteString("\tpgxdb \"github.com/nobuenhombre/suikat/pkg/db/connectors/postgres-pgx-db\"\n")
	buf.WriteString(")\n\n")
	buf.WriteString("// I" + repoName + "Repository defines the repository interface\n")
	buf.WriteString("type I" + repoName + "Repository interface {\n")

	// Extract method signatures from @repo-start/@repo-end blocks
	inBlock = false
	for _, line := range lines {
		if strings.Contains(line, "// @repo-start") {
			inBlock = true
			continue
		}
		if strings.Contains(line, "// @repo-end") {
			inBlock = false
			continue
		}
		if inBlock {
			trimmed := strings.TrimSpace(line)
			// Match only methods: func (receiver) MethodName(...)
			if strings.HasPrefix(trimmed, "func (") {
				// Remove opening brace and everything after
				if braceIdx := strings.Index(trimmed, "{"); braceIdx >= 0 {
					trimmed = trimmed[:braceIdx]
				}
				// Strip receiver
				if idx := strings.Index(trimmed, ") "); idx >= 0 {
					trimmed = "\t" + strings.TrimSpace(trimmed[idx+2:])
				}
				buf.WriteString(trimmed + "\n")
			}
		}
	}

	buf.WriteString("}\n\n")

	// Copy method implementations from @repo-start/@repo-end blocks (excluding markers)
	// Normalize receiver types to match the struct name (fixes mismatch
	// when xo-generated code uses {{ .Name }}Repository and xouid-generated
	// code uses {{ .Type }}Repository with different casing).
	inBlock = false
	for _, line := range lines {
		if strings.Contains(line, "// @repo-start") {
			inBlock = true
			continue
		}
		if strings.Contains(line, "// @repo-end") {
			inBlock = false
			continue
		}
		if inBlock {
			// Normalize receiver type: func (repo *XXXRepository) -> func (repo *repoNameRepository)
			if strings.Contains(line, "func (repo *") && strings.Contains(line, "Repository)") {
				prefix := "func (repo *"
				suffix := "Repository)"
				funcIdx := strings.Index(line, prefix)
				if funcIdx >= 0 {
					afterPrefix := line[funcIdx+len(prefix):]
					endIdx := strings.Index(afterPrefix, suffix)
					if endIdx > 0 {
						currentName := afterPrefix[:endIdx]
						if currentName != repoName {
							oldReceiver := prefix + currentName + suffix
							newReceiver := prefix + repoName + suffix
							line = strings.Replace(line, oldReceiver, newReceiver, 1)
						}
					}
				}
			}
			buf.WriteString(line + "\n")
		}
	}

	err = os.WriteFile(outputFile, []byte(buf.String()), 0644)
	if err != nil {
		return ge.Pin(fmt.Errorf("writing %s: %w", outputFile, err))
	}

	log.Printf("[xo] Created repo file: %s", filepath.Base(outputFile))
	return nil
}
