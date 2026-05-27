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

// countPipelineSteps returns the total number of top-level pipeline steps (8).
// Matches the 8 steps in Run(): runXO, replaceInterfaceToAny, glueXoXouid,
// extractRepo, removeXoXouid, cleanXoXouidSourceBlocks, generateDbRepo, goFormatCode.
func (d *AppDomain) countPipelineSteps(cfg *XOConfig) (int, error) {
	return 8, nil
}
