package domainapp

import (
	"log"

	"github.com/google/wire"
	"github.com/nobuenhombre/suikat/pkg/ge"

	"goxogen/src/internal/app/gobp/cli"
)

// ProviderSet exports Wire providers for the gobp domain package.
var ProviderSet = wire.NewSet(
	ProvideDomain,
)

// ProvideDomain creates the domain service (business-logic orchestrator).
func ProvideDomain(cliConfig cli.Service) (DomainService, func(), error) {
	cleanup := func() {
		log.Println("[wire-cleanup] Gobp domain cleanup")
	}

	dom, err := New(cliConfig)
	if err != nil {
		return nil, cleanup, ge.Pin(err)
	}

	return dom, cleanup, nil
}
