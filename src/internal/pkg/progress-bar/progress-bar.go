package progressbar

import (
	"fmt"
	"time"
)

// ANSI color codes for terminal output.
const (
	ColorTitle   = "\033[1;36m"
	ColorProject = "\033[1;37m"
	ColorSuccess = "\033[1;32m"
	ColorError   = "\033[1;31m"
	ColorWarning = "\033[1;33m"
	ColorTime    = "\033[1;34m"
	ColorEta     = "\033[1;35m"
	ColorReset   = "\033[0m"
)

// Progress bar display characters.
const (
	charDone    = "━"
	charPending = "━"
	charOk      = "✅"
	charError   = "✖"
	charWarning = "⚠"
	charTime    = "⏱"
	charEta     = "⌛"
)

const barLength = 50

// ProgressState holds the current state of the pipeline progress bar.
type ProgressState struct {
	Title       string
	ProjectName string
	Current     int
	Total       int
	Elapsed     int // seconds
	Remaining   int // seconds
	Errors      int
	StartTime   int64 // unix timestamp
}

// ProgressBar renders the progress bar as a single line (like gobp).
// Uses \r to overwrite the current line. currentFile is shown at the end
// of the bar line, truncated to fit on a standard terminal width.
func (p *ProgressState) ProgressBar(currentFile string) string {
	if p.Total == 0 {
		return ""
	}

	percent := int((p.Current * 100) / p.Total)
	if percent > 100 {
		percent = 100
	}
	filled := int((barLength * p.Current) / p.Total)
	if filled > barLength {
		filled = barLength
	}

	bar := "["
	for i := 0; i < barLength; i++ {
		if i < filled {
			bar += ColorSuccess + charDone + ColorReset
		} else {
			bar += ColorError + charPending + ColorReset
		}
	}
	bar += "]"

	etaStr := fmt.Sprintf("%02d:%02d", p.Remaining/60, p.Remaining%60)
	if p.Current <= 1 {
		etaStr = "--:--"
	}
	elapsedStr := fmt.Sprintf("%02d:%02d", p.Elapsed/60, p.Elapsed%60)

	line := fmt.Sprintf("\r%s %s%d%%%s (%d/%d) %s%s %s%s %s%s %s%s",
		bar,
		ColorWarning, percent, ColorReset,
		p.Current, p.Total,
		ColorTime, charTime, elapsedStr, ColorReset,
		ColorEta, charEta, etaStr, ColorReset,
	)

	if currentFile != "" {
		// Truncate currentFile to fit on one line (~35 chars)
		displayFile := currentFile
		if len(displayFile) > 35 {
			displayFile = "..." + displayFile[len(displayFile)-32:]
		}
		line += fmt.Sprintf(" %s▶%s %s", ColorProject, ColorReset, displayFile)
	}

	return line
}

// StartLine renders the initial header line.
func StartLine(title, project string) string {
	return fmt.Sprintf("\n%s%s%s [%s%s%s]\n",
		ColorTitle, title, ColorReset,
		ColorProject, project, ColorReset,
	)
}

// FinishLine renders the final success message.
func FinishLine(elapsed int) string {
	return fmt.Sprintf("\n%s%s Done in %d seconds!%s\n",
		ColorSuccess, charOk, elapsed, ColorReset,
	)
}

// ErrorLine renders the error summary as gobp does.
func ErrorLine(errors, elapsed int) string {
	return fmt.Sprintf("%s%s %d Errors detected%s\n%s%s Stopped after %d seconds%s\n",
		ColorError, charError, errors, ColorReset,
		ColorWarning, charWarning, elapsed, ColorReset,
	)
}

// ProgressTracker manages the progress bar lifecycle through the pipeline.
type ProgressTracker struct {
	state     *ProgressState
	startTime int64
	errors    []string
}

// NewProgressTracker creates a new tracker with a pre-counted total.
func NewProgressTracker(title string, total int) *ProgressTracker {
	now := time.Now().Unix()
	return &ProgressTracker{
		state: &ProgressState{
			Title:       title,
			ProjectName: "xoxgen",
			Current:     0,
			Total:       total,
			StartTime:   now,
		},
		startTime: now,
	}
}

// Increment advances the progress by one step and renders the bar.
// currentFile is shown at the end of the bar line (truncated if long).
func (pt *ProgressTracker) Increment(currentFile string) {
	pt.state.Current++
	if pt.state.Current > pt.state.Total {
		pt.state.Current = pt.state.Total
	}
	elapsed := int(time.Now().Unix() - pt.startTime)
	pt.state.Elapsed = elapsed
	if pt.state.Current > 1 {
		pt.state.Remaining = int(float64(elapsed) * float64(pt.state.Total-pt.state.Current) / float64(pt.state.Current))
	}
	fmt.Print(pt.state.ProgressBar(currentFile))
}

// AddError records an error line for display when Fail() is called.
func (pt *ProgressTracker) AddError(line string) {
	pt.state.Errors++
	pt.errors = append(pt.errors, line)
}

// Finish renders the final 100% state and success message.
func (pt *ProgressTracker) Finish() {
	pt.state.Current = pt.state.Total
	elapsed := int(time.Now().Unix() - pt.startTime)
	pt.state.Elapsed = elapsed
	pt.state.Remaining = 0

	// Final progress bar update, then success message (FinishLine has leading \n)
	fmt.Print(pt.state.ProgressBar(""))
	fmt.Print(FinishLine(elapsed))
}

// Fail renders the error state and prints all collected error lines.
// Breaks out of the \r-based progress bar display so errors are visible.
func (pt *ProgressTracker) Fail() {
	elapsed := int(time.Now().Unix() - pt.startTime)
	pt.state.Elapsed = elapsed

	// Move below the progress bar line
	fmt.Print("\n")
	fmt.Print(ErrorLine(pt.state.Errors, elapsed))
	for _, errLine := range pt.errors {
		fmt.Println(errLine)
	}
}
