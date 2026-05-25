package logfile

import (
	"log"
	"os"
)

// ILogFile interface for log file management.
type ILogFile interface {
	Open(name string)
	Close()
	Get() *os.File
}

// LogFile manages a log file with redirecting the standard logger.
type LogFile struct {
	file *os.File
}

// Open opens the log file and redirects the standard logger output.
func (lf *LogFile) Open(name string) {
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf(" -[exit]- error OpenFile log file (%v): %v", name, err)
	}

	lf.file = file

	log.SetOutput(lf.file)
}

// Close closes the log file if it was opened.
func (lf *LogFile) Close() {
	if lf.file != nil {
		err := lf.file.Close()
		if err != nil {
			log.Fatalf(" -[exit]- error Closing log file (%v): %v", lf.file.Name(), err)
		}
	}
}

// Get returns the underlying file handle.
func (lf *LogFile) Get() *os.File {
	return lf.file
}
