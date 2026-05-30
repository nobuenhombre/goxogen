package domainapp

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

// generateDbRepo scans all *-repo.xo.go files and generates an aggregate a-db-repo.go
// with a Db{DbName}Repo struct that holds all individual repositories.
// The constructor creates a DB connection internally via pgxdb.NewDB. If it fails,
// it wraps the error with ge.Pin and returns it.
func (d *AppDomain) generateDbRepo(outdir string, cfg *XOConfig) error {

	repoFiles, err := filepath.Glob(filepath.Join(outdir, "*-repo.xo.go"))
	if err != nil {
		return fmt.Errorf("listing repo files: %w", err)
	}

	if len(repoFiles) == 0 {
		log.Println("[xo] No repo files found, skipping db-repo generation")
		return nil
	}

	sort.Strings(repoFiles)

	// Extract repository names from each file
	var repos []string

	for _, file := range repoFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("reading %s: %w", file, err)
		}

		name := extractRepoInterfaceName(string(data))
		if name == "" {
			log.Printf("[xo] Warning: could not extract repo name from %s", filepath.Base(file))
			continue
		}

		repos = append(repos, name)
	}

	if len(repos) == 0 {
		log.Println("[xo] No repository names extracted, skipping db-repo generation")
		return nil
	}

	dbName := cfg.resolveDbName()

	// Load and execute the template
	tplContent, err := templateFS.ReadFile("templates/a-db-repo.go.tpl")
	if err != nil {
		return fmt.Errorf("reading embedded template a-db-repo.go.tpl: %w", err)
	}

	tmpl, err := template.New("a-db-repo.go.tpl").Parse(string(tplContent))
	if err != nil {
		return fmt.Errorf("parsing db-repo template: %w", err)
	}

	data := struct {
		Package string
		DbName  string
		Repos   []string
	}{
		Package: cfg.Config.Codegen.Package,
		DbName:  dbName,
		Repos:   repos,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing db-repo template: %w", err)
	}

	outputFile := filepath.Join(outdir, "a-db-repo.go")
	if err := os.WriteFile(outputFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", outputFile, err)
	}

	log.Printf("[xo] Created aggregate db-repo file: a-db-repo.go")
	return nil
}

// extractRepoInterfaceName extracts the repository name from an I{Name}Repository interface
// declaration in a *-repo.xo.go file.
func extractRepoInterfaceName(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Match "type I{Name}Repository interface {"
		if strings.HasPrefix(trimmed, "type I") && strings.Contains(trimmed, "Repository interface") {
			trimmed = strings.TrimPrefix(trimmed, "type I")
			idx := strings.Index(trimmed, "Repository")
			if idx > 0 {
				return trimmed[:idx]
			}
		}
	}
	return ""
}
