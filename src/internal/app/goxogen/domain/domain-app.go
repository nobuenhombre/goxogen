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

// countPipelineSteps returns the total number of top-level pipeline steps.
// Base count is 9: runXO, replaceInterfaceToAny, glueXoXouid,
// extractRepo, removeXoXouid, cleanXoXouidSourceBlocks, generateDbRepo,
// generateProvider, goFormatCode.
// When DbIsReadonly is true, there's an extra step: removeCRUDBlocks.
func (d *AppDomain) countPipelineSteps(cfg *XOConfig) (int, error) {
	steps := 9
	if cfg.Config.Codegen.DbIsReadonly {
		steps++
	}
	return steps, nil
}
