package domainapp

import "fmt"

// ANSI color codes for terminal output.
const (
	colorTitle   = "\033[1;36m"
	colorProject = "\033[1;37m"
	colorSuccess = "\033[1;32m"
	colorError   = "\033[1;31m"
	colorWarning = "\033[1;33m"
	colorTime    = "\033[1;34m"
	colorEta     = "\033[1;35m"
	colorReset   = "\033[0m"
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

// Default progress bar width.
const barLength = 50

// title is the default build title shown on the progress bar.
const Title = "Building project"

// ProgressState holds the current state of the build progress bar.
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

// ProgressBar renders the current progress bar to a string.
func (p *ProgressState) ProgressBar() string {
	if p.Total == 0 {
		return ""
	}

	percent := int((p.Current * 100) / p.Total)
	filled := int((barLength * p.Current) / p.Total)

	bar := "["
	for i := 0; i < barLength; i++ {
		if i < filled {
			bar += colorSuccess + charDone + colorReset
		} else {
			bar += colorError + charPending + colorReset
		}
	}
	bar += "]"

	etaStr := fmt.Sprintf("%02d:%02d", p.Remaining/60, p.Remaining%60)
	if p.Current <= 1 {
		etaStr = "--:--"
	}
	elapsedStr := fmt.Sprintf("%02d:%02d", p.Elapsed/60, p.Elapsed%60)

	return fmt.Sprintf("\r%s %s%d%%%s (%d/%d) %s%s %s%s %s%s %s%s",
		bar,
		colorWarning, percent, colorReset,
		p.Current, p.Total,
		colorTime, charTime, elapsedStr, colorReset,
		colorEta, charEta, etaStr, colorReset,
	)
}

// StartLine renders the initial header line.
func StartLine(title, project string) string {
	return fmt.Sprintf("\n%s%s%s [%s%s%s]\n",
		colorTitle, title, colorReset,
		colorProject, project, colorReset,
	)
}

// FinishLine renders the final success message.
func FinishLine(elapsed int) string {
	return fmt.Sprintf("\n%s%s Done in %d seconds!%s\n",
		colorSuccess, charOk, elapsed, colorReset,
	)
}

// ErrorLine renders the error summary.
func ErrorLine(errors, elapsed int) string {
	return fmt.Sprintf("\n%s%s %d Errors detected%s\n%s%s Stopped after %d seconds%s\n",
		colorError, charError, errors, colorReset,
		colorWarning, charWarning, elapsed, colorReset,
	)
}
