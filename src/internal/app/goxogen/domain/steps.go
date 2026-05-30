package domainapp

import (
	"fmt"
	progressbar "goxogen/src/internal/pkg/progress-bar"
	"io"
	"log"
	"os"
	"path/filepath"
)

const PipelineTitle = "Generating code"

// RunXO executes the full xo code generation pipeline with a progress bar.
func (d *AppDomain) Run() error {
	cfg, err := LoadXOConfig(d.Cli.Config)
	if err != nil {
		return fmt.Errorf("loading xo config: %w", err)
	}

	cs := cfg.XoConnectionString()
	csuid := cfg.XouidConnectionString()
	outdir := cfg.Config.Codegen.Path
	ignoreFields := cfg.Config.Codegen.IgnoreFields
	pkg := cfg.Config.Codegen.Package
	queries := cfg.Config.Codegen.Queries

	fmt.Printf("%sConnection string: %s%s\n", progressbar.ColorProject, cs, progressbar.ColorReset)
	fmt.Printf("%sOutput: %s%s\n", progressbar.ColorProject, outdir, progressbar.ColorReset)
	fmt.Printf("%sPackage: %s%s\n", progressbar.ColorProject, pkg, progressbar.ColorReset)
	fmt.Printf("%sQueries: %s%s\n", progressbar.ColorProject, queries, progressbar.ColorReset)

	// Pre-count pipeline steps for the progress bar
	totalSteps, err := d.countPipelineSteps(cfg)
	if err != nil {
		return fmt.Errorf("counting pipeline steps: %w", err)
	}

	pt := progressbar.NewProgressTracker(PipelineTitle, totalSteps)
	fmt.Print(progressbar.StartLine(PipelineTitle, "xoxgen"))

	// Redirect log output to discard during pipeline to prevent
	// log.Printf lines from pushing the progress bar cursor down.
	// Use -log flag to capture logs to a file if needed.
	logOutput := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(logOutput)

	// Extract embedded templates to a temp directory
	templates, err := TemplatesDir()
	if err != nil {
		return fmt.Errorf("extracting embedded templates: %w", err)
	}
	log.Printf("[xo] Embedded templates extracted to: %s", templates)

	// Step 1: Run xo generation
	pt.Increment("XO code generation")
	if err := d.runXO(cs, csuid, outdir, ignoreFields, pkg, templates, queries); err != nil {
		pt.AddError(err.Error())
		pt.Fail()
		return fmt.Errorf("xo generation: %w", err)
	}

	// Step 2: Replace interface{} with any
	pt.Increment("Replace interface{} with any")
	if err := d.replaceInterfaceToAny(outdir); err != nil {
		pt.AddError(err.Error())
		pt.Fail()
		return fmt.Errorf("replace interface{}: %w", err)
	}

	// Step 3: Glue .xo.go + .xouid.go -> .xo-xouid.go
	pt.Increment("Glue .xo.go + .xouid.go files")
	if err := d.glueXoXouid(outdir); err != nil {
		pt.AddError(err.Error())
		pt.Fail()
		return fmt.Errorf("glue xo/xouid: %w", err)
	}

	// Step 4: Extract repos from .xo-xouid.go into *-repo.xo.go
	pt.Increment("Extract repository interfaces")
	if err := d.extractRepo(outdir, pkg); err != nil {
		pt.AddError(err.Error())
		pt.Fail()
		return fmt.Errorf("extract repo: %w", err)
	}

	// Step 5: Remove .xo-xouid.go temp files
	pt.Increment("Remove temporary .xo-xouid.go files")
	if err := d.removeXoXouid(outdir); err != nil {
		pt.AddError(err.Error())
		pt.Fail()
		return fmt.Errorf("remove xo-xouid: %w", err)
	}

	// Step 6: Clean @repo blocks from .xo.go and .xouid.go
	pt.Increment("Clean @repo markers from source files")
	if err := d.cleanXoXouidSourceBlocks(outdir); err != nil {
		pt.AddError(err.Error())
		pt.Fail()
		return fmt.Errorf("clean repo blocks: %w", err)
	}

	// Step 7: Generate aggregate Db{Name}Repo struct (a-db-repo.go)
	pt.Increment("Generate aggregate DbRepo struct")
	if err := d.generateDbRepo(outdir, cfg); err != nil {
		pt.AddError(err.Error())
		pt.Fail()
		return fmt.Errorf("generate db repo: %w", err)
	}

	// Step 8: Generate Wire provider.go with ProviderSet (provider.go)
	pt.Increment("Generate Wire provider file")
	if err := d.generateProvider(outdir, cfg); err != nil {
		pt.AddError(err.Error())
		pt.Fail()
		return fmt.Errorf("generate provider: %w", err)
	}

	// Step 9: Format and vet code
	pt.Increment("Format and vet generated code")
	if err := d.goFormatCode(outdir); err != nil {
		pt.Fail()
		return fmt.Errorf("format code: %w", err)
	}

	pt.Finish()
	// Flush stdout so "✅ Done" appears before Wire cleanup log lines
	os.Stdout.Sync()

	return nil
}

// deleteGlob deletes files matching a glob pattern.
func (d *AppDomain) deleteGlob(pattern string) error {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return fmt.Errorf("removing %s: %w", f, err)
		}
	}

	return nil
}
