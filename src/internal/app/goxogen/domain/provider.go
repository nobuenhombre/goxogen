package domainapp

import (
	"log"

	"github.com/google/wire"
	"goxogen/src/internal/app/goxogen/cli"
)

// ProviderSet exports Wire providers for the domainapp package.
var ProviderSet = wire.NewSet(
	ProvideDomain,
)

// ProvideDomain creates the domain service (business-logic orchestrator).
func ProvideDomain(cliConfig cli.Service) (DomainService, func(), error) {
	cleanup := func() {
		log.Println("[wire-cleanup] Domain cleanup")
	}

	dom := New(cliConfig)

	return dom, cleanup, nil
}
