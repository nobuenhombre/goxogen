package postgres

import (
	"log"

	"github.com/google/wire"
	pgxdb "github.com/nobuenhombre/suikat/pkg/db/connectors/postgres-pgx-db"
	"github.com/nobuenhombre/suikat/pkg/ge"

	"goxogen/src/internal/app/xouid/cli"
)

// ProviderSet exports Wire providers for the postgres package.
var ProviderSet = wire.NewSet(
	ProvideDB,
)

// ProvideDB creates the PostgreSQL connection from CLI config DSN.
func ProvideDB(cliConfig cli.Service) (pgxdb.DBQuery, func(), error) {
	db, cleanup, err := NewDB(cliConfig.GetDSN())
	if err != nil {
		return nil, func() { log.Println("[wire-cleanup] DB cleanup (no-op)") }, ge.Pin(err)
	}

	return db, cleanup, nil
}
