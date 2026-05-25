package logfile

import (
	"log"

	"github.com/google/wire"
	"goxogen/src/internal/app/goxogen/cli"
)

// ProviderSet exports Wire providers for the logfile package.
var ProviderSet = wire.NewSet(
	ProvideLogFile,
)

// ProvideLogFile opens the log file if a path was provided via CLI flags.
func ProvideLogFile(cliConfig cli.Service) (ILogFile, func(), error) {
	cleanup := func() {
		log.Println("Log File cleanup")
	}

	logFile := &LogFile{}

	if len(cliConfig.(*cli.Config).LogFile) != 0 {
		logFile.Open(cliConfig.(*cli.Config).LogFile)
		cleanup = logFile.Close
	}

	return logFile, cleanup, nil
}
