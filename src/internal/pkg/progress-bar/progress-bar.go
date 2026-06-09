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
	ColorXoLine  = "\033[38;5;250m" // subtle grey for xo subprocess output
)

// ANSI cursor-control sequences.
const (
	AnsiCursorUp   = "\033[%dA"
	AnsiClearLine  = "\033[2K\r"
	AnsiEraseBelow = "\033[J"
)

// XoWindowLines is the number of fixed lines reserved below the progress
// bar for displaying xo subprocess output (scrolls bottom-to-top).
const XoWindowLines = 5

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

// ProgressBar renders the progress bar as a single line (like gobp), returning a string
// that includes \r so it can be printed directly.  currentFile is shown at the end
// of the bar line, truncated to fit on a standard terminal width.
//
// Deprecated for use with xo-window display: when xo subprocess output is visible
// under the bar the caller should use renderWithXoLines instead, which manages the
// full 6-line block.
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

	// XoOutputWindow: ring buffer of XoWindowLines lines for xo subprocess output.
	xoLines  [XoWindowLines]string
	xoCount  int  // how many lines currently filled (0..XoWindowLines)
	xoActive bool // whether the xo window has been drawn below the bar
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
// When the xo window is active (PushXoLine has been called), the full
// 6-line block (progress bar + 5 xo lines) is redrawn.
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

	if pt.xoActive {
		pt.renderWithXoLines()
	} else {
		fmt.Print(pt.state.ProgressBar(currentFile))
	}
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

	// If the xo window is active, first clear the 5 xo lines so
	// the error summary isn't buried inside the scrolling area.
	if pt.xoActive {
		pt.clearXoLines()
	}

	// Move below the progress bar line
	fmt.Print("\n")
	fmt.Print(ErrorLine(pt.state.Errors, elapsed))
	for _, errLine := range pt.errors {
		fmt.Println(errLine)
	}
}

// renderWithXoLines redraws the progress bar and the 5 xo output lines
// below it as a single block (6 lines total).  On first call (xoActive is
// false) the area below the progress bar is first cleared so any prior
// garbage (e.g. from a previous xo run) doesn't show through.  On
// subsequent calls the block is overwritten in-place using ANSI cursor-up.
func (pt *ProgressTracker) renderWithXoLines() {
	// Move cursor to the start of the progress bar line.
	// On first call (xoActive == false) we're still on the progress bar
	// line from the last Increment — no need to go up, but we must clear
	// whatever trash is below the bar.
	if pt.xoActive {
		// After a full render, cursor is on the 5th xo line (XoWindowLines
		// lines below the progress bar).  Move up by XoWindowLines lines to
		// reach the progress bar line again — NOT XoWindowLines+1, that
		// would overshoot and cascade upward.
		fmt.Printf(AnsiCursorUp, XoWindowLines)
	} else {
		// First call: clear everything from the progress bar line down to
		// the bottom of the terminal so no stale xo output shows through.
		fmt.Print(AnsiEraseBelow)
	}

	// Render progress bar (no file label — the xo window below shows
	// what's being processed).
	fmt.Print(pt.state.ProgressBar(""))

	// Render the xo lines — always write full XoWindowLines to keep the
	// terminal scrolling clean.  The ring buffer fills from the bottom
	// (last xoCount slots have data), so start displaying from the right offset.
	startIdx := XoWindowLines - pt.xoCount
	if startIdx < 0 {
		startIdx = 0
	}
	for i := 0; i < XoWindowLines; i++ {
		fmt.Print("\n") // move to next line
		fmt.Print(AnsiClearLine)
		lineIdx := startIdx + i
		if lineIdx < XoWindowLines && pt.xoLines[lineIdx] != "" {
			// Truncate long lines to avoid wrapping
			line := pt.xoLines[lineIdx]
			if len(line) > 100 {
				line = line[:100]
			}
			fmt.Print(ColorXoLine + line + ColorReset)
		}
	}

	pt.xoActive = true
}

// PushXoLine pushes a new xo subprocess output line into the ring buffer
// and redraws the progress bar + xo window.  Lines scroll bottom-to-top:
// old lines shift up, the new line appears at the bottom.
func (pt *ProgressTracker) PushXoLine(line string) {
	// Shift existing lines up
	for i := 0; i < XoWindowLines-1; i++ {
		pt.xoLines[i] = pt.xoLines[i+1]
	}
	pt.xoLines[XoWindowLines-1] = line
	if pt.xoCount < XoWindowLines {
		pt.xoCount++
	}

	pt.renderWithXoLines()
}

// clearXoLines erases the 5 xo output lines from the terminal without
// disturbing the progress bar line.  Called by ClearXoOutput and Fail.
func (pt *ProgressTracker) clearXoLines() {
	// After renderWithXoLines cursor is on the 5th xo line (line 5,
	// XoWindowLines \n's below the progress bar).  Go up 5 to line 0.
	fmt.Printf(AnsiCursorUp, XoWindowLines)

	// Clear 5 lines going down.  The loop clears lines 0-4, and the
	// trailing \n after each clear leaves cursor on line 5 — which
	// was NOT cleared.  Clear it explicitly below.
	for i := 0; i < XoWindowLines; i++ {
		fmt.Print(AnsiClearLine)
		fmt.Print("\n")
	}
	// Now on line 5 — clear the orphan.
	fmt.Print(AnsiClearLine)

	// Return to the progress bar line.
	fmt.Printf(AnsiCursorUp, XoWindowLines)
}

// ClearXoOutput clears the scrollable xo window from the terminal and
// resets the internal state so that subsequent Increment calls render
// a simple single-line progress bar again.  The progress bar itself is
// redrawn immediately so the screen doesn't go blank during the several
// seconds between xo basic and the next pipeline step.
func (pt *ProgressTracker) ClearXoOutput() {
	if !pt.xoActive {
		return
	}
	pt.clearXoLines()
	// Redraw the progress bar — clearXoLines clears everything including
	// the bar line, and without this redraw the terminal stays blank until
	// the next Increment.
	fmt.Print(pt.state.ProgressBar(""))
	pt.xoActive = false
	pt.xoCount = 0
	for i := 0; i < XoWindowLines; i++ {
		pt.xoLines[i] = ""
	}
}
