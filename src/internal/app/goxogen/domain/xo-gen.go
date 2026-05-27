package domainapp

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	progressbar "goxogen/src/internal/pkg/progress-bar"
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

	// Step 8: Format and vet code
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

// runXO orchestrates all xo/xouid sub-process calls.
func (d *AppDomain) runXO(cs, csuid, outdir, ignoreFields, pkg, templates, queries string) error {
	// Delete old generated files
	if err := d.deleteGlob(filepath.Join(outdir, "*.xo.go")); err != nil {
		return err
	}
	if err := d.deleteGlob(filepath.Join(outdir, "*.xouid.go")); err != nil {
		return err
	}

	// runXoBasic — generate models from database schema
	if err := d.runXoBasic(cs, outdir, ignoreFields, pkg, templates); err != nil {
		return fmt.Errorf("xo basic: %w", err)
	}

	// runXoQueriesOne
	if err := d.runXoQueries(cs, outdir, ignoreFields, queries, pkg, templates, "one", true); err != nil {
		return fmt.Errorf("xo queries one: %w", err)
	}

	// runXoQueriesMany
	if err := d.runXoQueries(cs, outdir, ignoreFields, queries, pkg, templates, "many", false); err != nil {
		return fmt.Errorf("xo queries many: %w", err)
	}

	// runXoQueriesUID — uses xouid tool
	if err := d.runXoQueriesUID(csuid, outdir, queries, pkg, templates); err != nil {
		return fmt.Errorf("xo queries uid: %w", err)
	}

	// Delete stored procedure files
	if err := d.deleteGlob(filepath.Join(outdir, "sp_*.xo.go")); err != nil {
		return err
	}

	return nil
}

// runXoBasic runs xo for basic schema generation
func (d *AppDomain) runXoBasic(cs, outdir, ignoreFields, pkg, templates string) error {
	args := []string{
		cs,
		"-o", outdir,
		"--template-path", templates,
		"--package", pkg,
	}
	if ignoreFields != "" {
		args = append(args, "--ignore-fields", ignoreFields)
	}

	cmdStr := "xo " + strings.Join(args, " ")
	log.Printf("[xo] Running: %s", cmdStr)

	cmd := exec.Command("xo", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("xo command failed: %w\n%s", err, string(output))
	}

	return nil
}

// runXoQueries runs xo in query mode for a subdirectory (one/many)
func (d *AppDomain) runXoQueries(cs, outdir, ignoreFields, querydir, pkg, templates, subdir string, onlyOne bool) error {
	qdir := filepath.Join(querydir, subdir)
	sqlFiles, err := filepath.Glob(filepath.Join(qdir, "*.sql"))
	if err != nil {
		return fmt.Errorf("listing %s queries: %w", subdir, err)
	}

	sort.Strings(sqlFiles)

	for _, filename := range sqlFiles {
		basename := filepath.Base(filename)
		noext := strings.TrimSuffix(basename, filepath.Ext(basename))

		// Parse "TypeName-FuncName" format
		parts := strings.SplitN(noext, "-", 2)
		typename := parts[0]
		funcname := ""
		if len(parts) > 1 {
			funcname = parts[1]
		} else {
			funcname = typename
		}

		goFile := filepath.Join(outdir, strings.ToLower(typename)+".xo.go")

		args := []string{
			cs,
			"-o", outdir,
			"--template-path", templates,
			"--package", pkg,
			"--query-mode",
			"--query-trim",
			"--query-strip",
			"--query-interpolate",
			"--query-type", typename,
			"--query-func", funcname,
		}
		if onlyOne {
			args = append(args, "--query-only-one")
		}
		if ignoreFields != "" {
			args = append(args, "--ignore-fields", ignoreFields)
		}

		// Check if file exists for --append
		if _, err := os.Stat(goFile); err == nil {
			args = append(args, "--append")
		}

		cmdStr := "xo " + strings.Join(args, " ") + " < " + filename
		log.Printf("[xo] Running: %s", cmdStr)

		cmd := exec.Command("xo", args...)

		// Read SQL from file
		sqlData, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("reading %s: %w", filename, err)
		}
		cmd.Stdin = strings.NewReader(string(sqlData))

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("xo query %s failed: %w\n%s", filename, err, string(output))
		}
	}

	return nil
}

// runXoQueriesUID runs xouid for UPDATE/INSERT/DELETE queries
func (d *AppDomain) runXoQueriesUID(dsn, outdir, querydir, pkg, templates string) error {
	qdir := filepath.Join(querydir, "uid")
	sqlFiles, err := filepath.Glob(filepath.Join(qdir, "*.sql"))
	if err != nil {
		return fmt.Errorf("listing uid queries: %w", err)
	}

	sort.Strings(sqlFiles)

	for _, filename := range sqlFiles {
		basename := filepath.Base(filename)
		noext := strings.TrimSuffix(basename, filepath.Ext(basename))

		parts := strings.SplitN(noext, "-", 2)
		typename := parts[0]
		funcname := ""
		if len(parts) > 1 {
			funcname = parts[1]
		} else {
			funcname = typename
		}

		args := []string{
			"-out=" + outdir,
			"-dsn=" + dsn,
			"-template-path=" + templates,
			"-package=" + pkg,
			"-query-type=" + typename,
			"-query-func=" + funcname,
			"-query=" + filename,
			"-verbose=false",
		}

		cmdStr := "xouid " + strings.Join(args, " ")
		log.Printf("[xo] Running: %s", cmdStr)

		cmd := exec.Command("/usr/local/bin/xouid", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("xouid query %s failed: %w\n%s", filename, err, string(output))
		}
	}

	return nil
}

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

// extractRepo extracts @repo-start/@repo-end blocks from .xo-xouid.go files into *-repo.xo.go files.
func (d *AppDomain) extractRepo(outdir, pkg string) error {
	files, err := filepath.Glob(filepath.Join(outdir, "*.xo-xouid.go"))
	if err != nil {
		return err
	}

	sort.Strings(files)

	for _, file := range files {
		if strings.HasSuffix(file, "-repo.xo.go") {
			continue
		}

		if err := d.extractRepoFile(outdir, file, pkg); err != nil {
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
		return fmt.Errorf("reading %s: %w", file, err)
	}

	content := string(data)

	// Check for markers
	if !strings.Contains(content, "// @repo-start") || !strings.Contains(content, "// @repo-end") {
		return fmt.Errorf("markers not found in %s", filepath.Base(file))
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
		return fmt.Errorf("no repository name found in %s", filepath.Base(file))
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
			buf.WriteString(line + "\n")
		}
	}

	if err := os.WriteFile(outputFile, []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", outputFile, err)
	}

	log.Printf("[xo] Created repo file: %s", filepath.Base(outputFile))
	return nil
}

// removeXoXouid deletes all .xo-xouid.go files.
func (d *AppDomain) removeXoXouid(dir string) error {
	return d.deleteGlob(filepath.Join(dir, "*.xo-xouid.go"))
}

// cleanXoXouidSourceBlocks removes @repo-start/@repo-end blocks from .xo.go and .xouid.go files.
func (d *AppDomain) cleanXoXouidSourceBlocks(dir string) error {
	// Process .xo.go files
	if err := d.removeRepoBlocksFromFiles(dir, "*.xo.go", "*-repo.xo.go"); err != nil {
		return err
	}

	// Process .xouid.go files
	if err := d.removeRepoBlocksFromFiles(dir, "*.xouid.go", "*-repo.xo.go"); err != nil {
		return err
	}

	return nil
}

// removeRepoBlocksFromFiles removes @repo-start/@repo-end blocks from matching files.
func (d *AppDomain) removeRepoBlocksFromFiles(dir, pattern, excludePattern string) error {
	files, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return err
	}

	sort.Strings(files)

	for _, file := range files {
		// Skip -repo.xo.go files
		if strings.HasSuffix(file, "-repo.xo.go") {
			continue
		}

		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("reading %s: %w", file, err)
		}

		content := string(data)
		if !strings.Contains(content, "// @repo-start") {
			continue
		}

		// Remove lines between @repo-start and @repo-end (including markers)
		var result strings.Builder
		skip := false
		for _, line := range strings.Split(content, "\n") {
			if strings.Contains(line, "// @repo-start") {
				skip = true
				continue
			}
			if strings.Contains(line, "// @repo-end") {
				skip = false
				continue
			}
			if !skip {
				result.WriteString(line + "\n")
			}
		}

		output := result.String()
		// Remove trailing newline that we added
		if strings.HasSuffix(output, "\n") {
			output = strings.TrimSuffix(output, "\n")
		}

		if err := os.WriteFile(file, []byte(output), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", file, err)
		}

		log.Printf("[xo] Cleaned repo blocks from %s", filepath.Base(file))
	}

	return nil
}

// goFormatCode runs go fmt, goimports, and go vet.
// Errors from fmt/goimports/vet are collected and displayed like gobp.
func (d *AppDomain) goFormatCode(path string) error {
	// Step 1: go fmt
	log.Printf("[xo] Running go fmt %s", path)

	cmd := exec.Command("go", "fmt", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// go fmt returns non-zero on parse errors
		errMsg := strings.TrimSpace(string(output))
		if errMsg != "" {
			log.Printf("[xo] go fmt error: %s", errMsg)
		}
		return fmt.Errorf("go fmt failed: %w", err)
	}
	// Print formatting changes (modified filenames) if any
	if trimmed := strings.TrimSpace(string(output)); trimmed != "" {
		log.Printf("[xo] go fmt modified: %s", trimmed)
	}

	// Step 2: goimports
	log.Printf("[xo] Running goimports -w on %s", path)

	goFiles, err := filepath.Glob(filepath.Join(path, "*.go"))
	if err != nil {
		return fmt.Errorf("listing go files: %w", err)
	}

	if len(goFiles) > 0 {
		args := append([]string{"-w"}, goFiles...)
		cmd = exec.Command("goimports", args...)
		output, err = cmd.CombinedOutput()
		if err != nil {
			errMsg := strings.TrimSpace(string(output))
			// goimports may not be installed — log as warning
			if errMsg != "" {
				log.Printf("[xo] goimports error: %s", errMsg)
			}
			log.Printf("[xo] Warning: goimports failed (may not be installed): %v", err)
		} else {
			if trimmed := strings.TrimSpace(string(output)); trimmed != "" {
				log.Printf("[xo] goimports output: %s", trimmed)
			}
		}
	}

	// Step 3: go vet
	log.Printf("[xo] Running go vet %s", path)

	cmd = exec.Command("go", "vet", path)
	output, err = cmd.CombinedOutput()
	if err != nil {
		errMsg := strings.TrimSpace(string(output))
		if errMsg != "" {
			log.Printf("[xo] go vet error: %s", errMsg)
		}
		return fmt.Errorf("go vet failed: %w", err)
	}
	if trimmed := strings.TrimSpace(string(output)); trimmed != "" {
		log.Printf("[xo] go vet output: %s", trimmed)
	}

	return nil
}

// generateDbRepo scans all *-repo.xo.go files and generates an aggregate a-db-repo.go
// with a Db{DbName}Repo struct that holds all individual repositories.
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
	type repoInfo struct {
		Name string // e.g. "User"
	}

	var repos []repoInfo

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

		repos = append(repos, repoInfo{Name: name})
	}

	if len(repos) == 0 {
		log.Println("[xo] No repository names extracted, skipping db-repo generation")
		return nil
	}

	dbName := cfg.resolveDbName()

	outputFile := filepath.Join(outdir, "a-db-repo.go")

	var buf strings.Builder
	buf.WriteString("// Code generated by goxogen. DO NOT EDIT.\n")
	buf.WriteString("package " + cfg.Config.Codegen.Package + "\n\n")
	buf.WriteString("import (\n")
	buf.WriteString("\tpgxdb \"github.com/nobuenhombre/suikat/pkg/db/connectors/postgres-pgx-db\"\n")
	buf.WriteString(")\n\n")

	// Struct definition
	buf.WriteString(fmt.Sprintf("type Db%sRepo struct {\n", dbName))
	buf.WriteString("\tdb pgxdb.DBQuery\n")
	for _, r := range repos {
		buf.WriteString(fmt.Sprintf("\t%s I%sRepository\n", r.Name, r.Name))
	}
	buf.WriteString("}\n\n")

	// Constructor
	buf.WriteString(fmt.Sprintf("func NewDb%sRepository(db pgxdb.DBQuery) *Db%sRepo {\n", dbName, dbName))
	buf.WriteString(fmt.Sprintf("\treturn &Db%sRepo{\n", dbName))
	buf.WriteString("\t\tdb: db,\n")
	for _, r := range repos {
		buf.WriteString(fmt.Sprintf("\t\t%s: New%sRepository(db),\n", r.Name, r.Name))
	}
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")

	if err := os.WriteFile(outputFile, []byte(buf.String()), 0644); err != nil {
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
