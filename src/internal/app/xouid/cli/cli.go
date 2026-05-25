package cli

import (
	"github.com/nobuenhombre/suikat/pkg/clivar"
	"github.com/nobuenhombre/suikat/pkg/ge"
)

// Service defines the CLI configuration interface.
type Service interface {
	GetOut() string
	GetDSN() string
	GetTemplatePath() string
	GetPackage() string
	GetSchema() string
	GetQueryType() string
	GetQueryFunc() string
	GetQuery() string
	GetVerbose() bool
}

// Config represents the command-line interface configuration for xouid.
type Config struct {
	Out          string `cli:"out[output path]:string="`
	DSN          string `cli:"dsn[PostgreSQL DSN]:string="`
	TemplatePath string `cli:"template-path[user supplied template path]:string="`
	Package      string `cli:"package[package name used in generated Go code]:string="`
	Schema       string `cli:"schema[schema name to generate Go types for]:string=public"`
	QueryType    string `cli:"query-type[query generated Go type filename.xo.go]:string="`
	QueryFunc    string `cli:"query-func[query generated Go func name]:string="`
	Query        string `cli:"query[query file to generate Go type and func from]:string="`
	Verbose      bool   `cli:"verbose[dont view hello message]:bool=false"`
	Version      bool   `cli:"version[Show version and exit]:bool=false"`
}

// GetOut returns the output path.
func (c *Config) GetOut() string { return c.Out }

// GetDSN returns the PostgreSQL DSN.
func (c *Config) GetDSN() string { return c.DSN }

// GetTemplatePath returns the user supplied template path.
func (c *Config) GetTemplatePath() string { return c.TemplatePath }

// GetPackage returns the package name.
func (c *Config) GetPackage() string { return c.Package }

// GetSchema returns the schema name.
func (c *Config) GetSchema() string { return c.Schema }

// GetQueryType returns the query type.
func (c *Config) GetQueryType() string { return c.QueryType }

// GetQueryFunc returns the query func name.
func (c *Config) GetQueryFunc() string { return c.QueryFunc }

// GetQuery returns the query file path.
func (c *Config) GetQuery() string { return c.Query }

// GetVerbose returns the verbose flag.
func (c *Config) GetVerbose() bool { return c.Verbose }

// New creates a new Config instance by loading values from command-line arguments.
func New() (Service, error) {
	cfg := &Config{}

	err := clivar.Load(cfg)
	if err != nil {
		return nil, ge.Pin(err)
	}

	return cfg, nil
}
