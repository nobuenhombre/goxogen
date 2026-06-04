package domainapp

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nobuenhombre/suikat/pkg/ge"
)

// runXO orchestrates all xo/xouid sub-process calls.
func (d *AppDomain) runXO(cs, csuid, outdir, ignoreFields, pkg, templates, queries string) error {
	// Delete old generated files
	err := d.deleteGlob(filepath.Join(outdir, "*.xo.go"))
	if err != nil {
		return ge.Pin(err)
	}
	err = d.deleteGlob(filepath.Join(outdir, "*.xouid.go"))
	if err != nil {
		return ge.Pin(err)
	}

	// runXoBasic — generate models from database schema
	err = d.runXoBasic(cs, outdir, ignoreFields, pkg, templates)
	if err != nil {
		return ge.Pin(fmt.Errorf("xo basic: %w", err))
	}

	// runXoQueriesOne
	err = d.runXoQueries(cs, outdir, ignoreFields, queries, pkg, templates, "one", true)
	if err != nil {
		return ge.Pin(fmt.Errorf("xo queries one: %w", err))
	}

	// runXoQueriesMany
	err = d.runXoQueries(cs, outdir, ignoreFields, queries, pkg, templates, "many", false)
	if err != nil {
		return ge.Pin(fmt.Errorf("xo queries many: %w", err))
	}

	// runXoQueriesUID — uses xouid tool
	err = d.runXoQueriesUID(csuid, outdir, queries, pkg, templates)
	if err != nil {
		return ge.Pin(fmt.Errorf("xo queries uid: %w", err))
	}

	// Delete stored procedure files
	err = d.deleteGlob(filepath.Join(outdir, "sp_*.xo.go"))
	if err != nil {
		return ge.Pin(err)
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
		return ge.Pin(fmt.Errorf("xo command failed: %w\n%s", err, string(output)))
	}

	return nil
}

// runXoQueries runs xo in query mode for a subdirectory (one/many)
func (d *AppDomain) runXoQueries(cs, outdir, ignoreFields, querydir, pkg, templates, subdir string, onlyOne bool) error {
	qdir := filepath.Join(querydir, subdir)
	sqlFiles, err := filepath.Glob(filepath.Join(qdir, "*.sql"))
	if err != nil {
		return ge.Pin(fmt.Errorf("listing %s queries: %w", subdir, err))
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
			return ge.Pin(fmt.Errorf("reading %s: %w", filename, err))
		}
		cmd.Stdin = strings.NewReader(string(sqlData))

		output, err := cmd.CombinedOutput()
		if err != nil {
			return ge.Pin(fmt.Errorf("xo query %s failed: %w\n%s", filename, err, string(output)))
		}
	}

	return nil
}

// runXoQueriesUID runs xouid for UPDATE/INSERT/DELETE queries
func (d *AppDomain) runXoQueriesUID(dsn, outdir, querydir, pkg, templates string) error {
	qdir := filepath.Join(querydir, "uid")
	sqlFiles, err := filepath.Glob(filepath.Join(qdir, "*.sql"))
	if err != nil {
		return ge.Pin(fmt.Errorf("listing uid queries: %w", err))
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
			return ge.Pin(fmt.Errorf("xouid query %s failed: %w\n%s", filename, err, string(output)))
		}
	}

	return nil
}
