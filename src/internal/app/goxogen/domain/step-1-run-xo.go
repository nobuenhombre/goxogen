package domainapp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	progressbar "goxogen/src/internal/pkg/progress-bar"

	"github.com/nobuenhombre/suikat/pkg/ge"
)

// runXO orchestrates all xo/xouid sub-process calls.
func (d *AppDomain) runXO(cs, csuid, outdir, ignoreFields, pkg, templates, queries string, pt *progressbar.ProgressTracker) error {
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
	err = d.runXoBasic(cs, outdir, ignoreFields, pkg, templates, pt)
	if err != nil {
		return ge.Pin(fmt.Errorf("xo basic: %w", err))
	}

	// runXoQueriesOne
	err = d.runXoQueries(cs, outdir, ignoreFields, queries, pkg, templates, "one", true, pt)
	if err != nil {
		return ge.Pin(fmt.Errorf("xo queries one: %w", err))
	}

	// runXoQueriesMany
	err = d.runXoQueries(cs, outdir, ignoreFields, queries, pkg, templates, "many", false, pt)
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

// runXoBasic runs xo for basic schema generation, streaming its output
// line-by-line into the xo window below the progress bar.
func (d *AppDomain) runXoBasic(cs, outdir, ignoreFields, pkg, templates string, pt *progressbar.ProgressTracker) error {
	args := []string{
		cs,
		"-o", outdir,
		"--template-path", templates,
		"--package", pkg,
		"-v",
	}
	if ignoreFields != "" {
		args = append(args, "--ignore-fields", ignoreFields)
	}

	cmdStr := "xo " + strings.Join(args, " ")
	log.Printf("[xo] Running: %s", cmdStr)

	cmd := exec.Command("xo", args...)

	outputBuf, err := d.runXoStreamCmd(cmd, pt, cmdStr)
	if err != nil {
		return ge.Pin(fmt.Errorf("xo command failed: %w\n%s", err, outputBuf))
	}

	// Clear the xo window so the next pipeline step starts clean
	pt.ClearXoOutput()

	return nil
}

// runXoQueries runs xo in query mode for a subdirectory (one/many)
func (d *AppDomain) runXoQueries(cs, outdir, ignoreFields, querydir, pkg, templates, subdir string, onlyOne bool, pt *progressbar.ProgressTracker) error {
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
			"-v",
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

		outputBuf, err := d.runXoStreamCmd(cmd, pt, cmdStr)
		if err != nil {
			return ge.Pin(fmt.Errorf("xo query %s failed: %w\n%s", filename, err, outputBuf))
		}
	}

	// Clear xo window so the next subdir starts clean
	pt.ClearXoOutput()

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

// isXoObjectName filters xo -v output lines to keep only lines that
// represent a database object name (table, index, sequence, etc.).
// The rest (empty lines, SQL queries, header markers like "SQL:" /
// "PARAMS:", and empty-param lines like "[public]" / "[public r]")
// are discarded so the 5-line xo window stays readable.
// Returns the cleaned line (with is_view suffix stripped) and true, or
// "" and false if the line should be dropped.
func isXoObjectName(line string) (string, bool) {
	if line == "" {
		return "", false
	}
	// Header markers
	if line == "SQL:" || line == "PARAMS:" || line == "[]" {
		return "", false
	}
	// Full SQL statements and query fragments
	if strings.HasPrefix(line, "SELECT ") ||
		strings.HasPrefix(line, "FROM ") ||
		strings.HasPrefix(line, "WHERE ") ||
		strings.HasPrefix(line, "AND ") ||
		strings.HasPrefix(line, "OR ") ||
		strings.HasPrefix(line, "ORDER BY ") ||
		strings.HasPrefix(line, "GROUP BY ") ||
		strings.HasPrefix(line, "LIMIT ") ||
		strings.HasPrefix(line, "CREATE ") ||
		strings.HasPrefix(line, "INNER JOIN ") ||
		strings.HasPrefix(line, "LEFT JOIN ") ||
		strings.HasPrefix(line, "RIGHT JOIN ") ||
		strings.HasPrefix(line, "ON (") ||
		strings.HasPrefix(line, "EXCEPT") ||
		strings.HasPrefix(line, "UNION") ||
		strings.HasPrefix(line, "HAVING ") ||
		strings.HasPrefix(line, "-- ") ||
		strings.HasPrefix(line, "COUNT(") ||
		strings.HasPrefix(line, "SUM(") {
		return "", false
	}
	// Empty-param lines — only schema or relkind noise
	if line == "[public]" || line == "[public r]" {
		return "", false
	}
	// Strip the trailing " true"/" false" (is_view flag) that xo appends
	// to table names inside the brackets:
	//   "[public tablename false]" → "[public tablename]"
	cleaned := strings.TrimSuffix(line, " true]")
	if cleaned == line {
		cleaned = strings.TrimSuffix(line, " false]")
	}
	if cleaned != line {
		line = cleaned + "]"
	}
	return line, true
}

// runXoStreamCmd starts a command and streams its stdout+stderr output
// line-by-line through the xo window (with isXoObjectName filtering).
// Returns the captured output string and any error from the command.
// The caller is responsible for calling pt.ClearXoOutput() when done.
func (d *AppDomain) runXoStreamCmd(cmd *exec.Cmd, pt *progressbar.ProgressTracker, cmdLabel string) (string, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", ge.Pin(fmt.Errorf("stdout pipe: %w", err))
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", ge.Pin(fmt.Errorf("stderr pipe: %w", err))
	}

	if err := cmd.Start(); err != nil {
		return "", ge.Pin(fmt.Errorf("start: %w", err))
	}

	// Merge stdout+stderr concurrently (MultiReader is sequential).
	// Two goroutines copy each pipe into a shared io.Pipe.
	pr, pw := io.Pipe()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(pw, stdout)
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(pw, stderr)
	}()
	go func() {
		wg.Wait()
		pw.Close()
	}()

	// Read from the merged pipe line-by-line, filtering for the xo window.
	scanner := bufio.NewScanner(pr)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	var outBuf bytes.Buffer
	outBuf.WriteString("[xo] " + cmdLabel + "\n")

	for scanner.Scan() {
		line := scanner.Text()
		outBuf.WriteString(line + "\n")

		if cleaned, ok := isXoObjectName(line); ok {
			pt.PushXoLine(cleaned)
		}
	}

	// Drain any pipe errors
	if err := scanner.Err(); err != nil {
		log.Printf("[xo] scanner error: %v", err)
	}

	waitErr := cmd.Wait()
	if waitErr != nil {
		return outBuf.String(), ge.Pin(fmt.Errorf("command failed: %w", waitErr))
	}

	return "", nil
}
