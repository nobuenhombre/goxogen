package cli

import (
	"github.com/nobuenhombre/suikat/pkg/clivar"
	"github.com/nobuenhombre/suikat/pkg/ge"
)

// Service defines the CLI configuration interface for gobp.
type Service interface {
	GetBinary() string
	GetOut() string
	GetVerbose() bool
	GetFullRebuild() bool
}

// Config represents the command-line interface configuration for gobp.
type Config struct {
	Binary      string `cli:"binary[Go binary to build, e.g. ./src/cmd/myapp]:string=."`
	Out         string `cli:"out[output binary path]:string=./build/app"`
	Verbose     bool   `cli:"verbose[show full build output]:bool=false"`
	FullRebuild bool   `cli:"full-rebuild[force full rebuild with go build -a]:bool=false"`
}

// GetBinary returns the Go package path or directory to build.
func (c *Config) GetBinary() string { return c.Binary }

// GetOut returns the output binary path.
func (c *Config) GetOut() string { return c.Out }

// GetVerbose returns the verbose flag.
func (c *Config) GetVerbose() bool { return c.Verbose }

// GetFullRebuild returns whether to force a full rebuild with go build -a.
func (c *Config) GetFullRebuild() bool { return c.FullRebuild }

// New creates a new Config instance by loading values from command-line arguments.
func New() (Service, error) {
	cfg := &Config{}

	err := clivar.Load(cfg)
	if err != nil {
		return nil, ge.Pin(err)
	}

	return cfg, nil
}
