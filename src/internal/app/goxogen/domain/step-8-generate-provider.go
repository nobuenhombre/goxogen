package domainapp

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

// generateProvider creates a provider.go file with a Wire ProviderSet
// that exposes a Provider{DbName} constructor for the aggregate DbRepo.
// This file is generated after generateDbRepo (step 7) so the DbRepo
// and its constructor already exist in the output directory.
func (d *AppDomain) generateProvider(outdir string, cfg *XOConfig) error {
	dbName := cfg.resolveDbName()

	// Check if a-db-repo.go exists — no repo means nothing to provide
	repoFile := filepath.Join(outdir, "a-db-repo.go")
	if _, err := os.Stat(repoFile); os.IsNotExist(err) {
		log.Println("[xo] No a-db-repo.go found, skipping provider generation")
		return nil
	}

	// Load and execute the embedded template
	tplContent, err := templateFS.ReadFile("templates/provider.go.tpl")
	if err != nil {
		return fmt.Errorf("reading embedded template provider.go.tpl: %w", err)
	}

	tmpl, err := template.New("provider.go.tpl").Parse(string(tplContent))
	if err != nil {
		return fmt.Errorf("parsing provider template: %w", err)
	}

	data := struct {
		Package string
		DbName  string
	}{
		Package: cfg.Config.Codegen.Package,
		DbName:  dbName,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing provider template: %w", err)
	}

	outputFile := filepath.Join(outdir, "provider.go")
	if err := os.WriteFile(outputFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", outputFile, err)
	}

	log.Printf("[xo] Created Wire provider file: provider.go")
	return nil
}
