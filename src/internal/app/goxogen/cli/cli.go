package cli

import (
	"github.com/nobuenhombre/suikat/pkg/clivar"
	"github.com/nobuenhombre/suikat/pkg/ge"
)

type Service interface {
}

// Config represents the command-line interface configuration structure.
type Config struct {
	Config  string `cli:"config[Path to YAML config]:string=config.yaml"`
	LogFile string `cli:"log[Path to log file]:string="`
	Version bool   `cli:"version[Show version and exit]:bool=false"`
}

// New creates a new Config instance by loading values from command-line arguments.
func New() (Service, error) {
	cfg := &Config{}

	err := clivar.Load(cfg)
	if err != nil {
		return nil, ge.Pin(err)
	}

	return cfg, nil
}
