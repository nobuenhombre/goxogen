package postgres

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	pgxdb "github.com/nobuenhombre/suikat/pkg/db/connectors/postgres-pgx-db"
	"github.com/nobuenhombre/suikat/pkg/ge"
)

// NewDB creates a new PostgreSQL connection pool.
func NewDB(dataSourceName string) (pgxdb.DBQuery, func(), error) {
	config, err := pgxpool.ParseConfig(dataSourceName)
	if err != nil {
		return nil, nil, ge.Pin(err)
	}

	connectPool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, nil, ge.Pin(err)
	}

	cleanup := func() {
		log.Println("[wire-cleanup] PostgreSQL DB closing")
		connectPool.Close()
	}

	return &pgxdb.Conn{
		Pool: connectPool,
	}, cleanup, nil
}
