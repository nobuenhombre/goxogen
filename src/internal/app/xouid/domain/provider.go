package domainapp

import (
	"log"

	"github.com/google/wire"
	pgxdb "github.com/nobuenhombre/suikat/pkg/db/connectors/postgres-pgx-db"
	"github.com/nobuenhombre/suikat/pkg/ge"

	"goxogen/src/internal/app/xouid/cli"
)

// ProviderSet exports Wire providers for the xouid domain package.
var ProviderSet = wire.NewSet(
	ProvideDomain,
)

// ProvideDomain creates the domain service (business-logic orchestrator).
func ProvideDomain(cliConfig cli.Service, db pgxdb.DBQuery) (DomainService, func(), error) {
	cleanup := func() {
		log.Println("[wire-cleanup] Xouid domain cleanup")
	}

	dom, err := New(cliConfig, db)
	if err != nil {
		return nil, cleanup, ge.Pin(err)
	}

	return dom, cleanup, nil
}
