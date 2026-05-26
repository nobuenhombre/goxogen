package domainapp

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/nobuenhombre/suikat/pkg/ge"

	"goxogen/src/internal/app/gobp/cli"
)

// DomainService is the business-logic orchestrator for gobp.
type DomainService interface {
	Run() error
}

// AppDomain implements DomainService by running go build -x and showing progress.
type AppDomain struct {
	cliConfig cli.Service
}

// New creates a new gobp domain service.
func New(cliConfig cli.Service) (DomainService, error) {
	return &AppDomain{
		cliConfig: cliConfig,
	}, nil
}

// countSteps runs go build -n (dry-run) to count the total number of build steps.
func countSteps(args []string) (int, error) {
	// Insert -n right after "build" (args[0] = "build")
	dryArgs := make([]string, 0, len(args)+1)
	dryArgs = append(dryArgs, args[0])
	dryArgs = append(dryArgs, "-n")
	dryArgs = append(dryArgs, args[1:]...)

	cmd := exec.Command("go", dryArgs...)

	// go build -n outputs to stderr, not stdout
	out, err := cmd.StderrPipe()
	if err != nil {
		return 0, ge.Pin(err)
	}

	if err := cmd.Start(); err != nil {
		return 0, ge.Pin(err)
	}

	// Use large buffer — go build -n with CGO and long gcc invocations can
	// produce lines that exceed the default 64KB scanner limit, causing
	// the scanner to silently stop and the build command to hang on write.
	steps := 0
	scanner := bufio.NewScanner(out)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		// Count all command lines (mkdir, gcc, ar, compile, link, cd, cat, etc.)
		// — not just mkdir. This gives a smooth, meaningful progress estimate.
		if strings.HasPrefix(line, "mkdir") ||
			strings.HasPrefix(line, "/") || // absolute path = gcc/ar/compile/link
			strings.HasPrefix(line, "cd ") ||
			strings.HasPrefix(line, "cat ") {
			steps++
		}
	}

	if err := cmd.Wait(); err != nil {
		return 0, ge.Pin(fmt.Errorf("failed to count build steps: %v", err))
	}

	if err := scanner.Err(); err != nil {
		return steps, ge.Pin(fmt.Errorf("scanner error: %v (partial count: %d)", err, steps))
	}

	return steps, nil
}

// isBuildCommandLine returns true if the line from go build -x output
// represents a command execution (mkdir, compile, link, catalog, etc.).
func isBuildCommandLine(line string) bool {
	return strings.HasPrefix(line, "mkdir") ||
		strings.HasPrefix(line, "/") ||
		strings.HasPrefix(line, "cd ") ||
		strings.HasPrefix(line, "cat ")
}

// Run executes the build progress display.
func (d *AppDomain) Run() error {
	binary := d.cliConfig.GetBinary()
	out := d.cliConfig.GetOut()

	// Build the go build command arguments
	commonArgs := []string{"build"}
	if d.cliConfig.GetFullRebuild() {
		commonArgs = append(commonArgs, "-a")
	}
	commonArgs = append(commonArgs, "-o", out, binary)

	// Phase 1: Pre-count total mkdir steps with dry-run
	totalSteps, err := countSteps(commonArgs)
	if err != nil {
		return ge.Pin(err)
	}

	if totalSteps == 0 {
		// Binary is up to date in cache — run a plain go build to ensure output exists
		fmt.Print(colorSuccess + "✓" + colorReset + " Binary is up to date — running go build...")
		buildCmd := exec.Command("go", commonArgs...)
		buildOut, buildErr := buildCmd.CombinedOutput()
		if buildErr != nil {
			fmt.Println(" " + colorError + "[FAILED]" + colorReset)
			return ge.Pin(fmt.Errorf("build failed: %v\n%s", buildErr, buildOut))
		}
		fmt.Println(" " + colorSuccess + "[OK]" + colorReset)
		return nil
	}

	// Print the header
	fmt.Print(StartLine(Title, binary))

	startTime := time.Now()
	state := &ProgressState{
		Title:       Title,
		ProjectName: binary,
		StartTime:   startTime.Unix(),
		Total:       totalSteps,
	}

	// Phase 2: Actual build with -x to show mkdir steps
	// Insert -x right after "build"
	buildArgs := make([]string, 0, len(commonArgs)+1)
	buildArgs = append(buildArgs, commonArgs[0])
	buildArgs = append(buildArgs, "-x")
	buildArgs = append(buildArgs, commonArgs[1:]...)
	cmd := exec.Command("go", buildArgs...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return ge.Pin(err)
	}

	if err := cmd.Start(); err != nil {
		return ge.Pin(err)
	}

	// Read only from stderr — go build -x outputs everything to stderr, stdout stays empty
	scanner := bufio.NewScanner(stderr)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	step := 0
	errorCount := 0
	var errorLines []string

	for scanner.Scan() {
		line := scanner.Text()

		if isBuildCommandLine(line) {
			step++
			state.Current = step
			state.Elapsed = int(time.Since(startTime).Seconds())
			if step > 1 {
				state.Remaining = int(float64(state.Elapsed) * float64(totalSteps-step) / float64(step))
			}

			fmt.Print(state.ProgressBar())
		}

		// Collect error lines — print them after ErrorLine, not inline
		if strings.Contains(line, ".go:") {
			errorCount++
			errorLines = append(errorLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return ge.Pin(err)
	}

	err = cmd.Wait()
	elapsed := int(time.Since(startTime).Seconds())

	if err != nil {
		state.Errors = errorCount
		fmt.Print(ErrorLine(errorCount, elapsed))

		// Print collected error lines after the ⚠ Stopped after N seconds line
		for _, errorLine := range errorLines {
			fmt.Println(errorLine)
		}

		return ge.Pin(fmt.Errorf("build failed with %d error(s)", errorCount))
	}

	// Final: render 100% and success message
	state.Current = totalSteps
	state.Elapsed = elapsed
	fmt.Print(state.ProgressBar())
	fmt.Print(FinishLine(elapsed))

	return nil
}
