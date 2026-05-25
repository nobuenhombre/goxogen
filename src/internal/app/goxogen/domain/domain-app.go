package domainapp

import (
	"goxogen/src/internal/app/goxogen/cli"
)

type DomainService interface {
	Run() error
}

type AppDomain struct {
	Cli *cli.Config
}

func New(cliConfig cli.Service) DomainService {
	return &AppDomain{
		Cli: cliConfig.(*cli.Config),
	}
}

func (d *AppDomain) Run() error {
	// Add your domain logic here
	return nil
}
