//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"goxogen/src/internal/app/xouid/cli"
	domainapp "goxogen/src/internal/app/xouid/domain"
	"goxogen/src/internal/app/xouid/postgres"
)

// initializeApp is the Wire injector entrypoint for xouid.
// It aggregates all ProviderSets and constructs the top-level application.
func initializeApp() (IApp, func(), error) {
	wire.Build(
		cli.ProviderSet,
		postgres.ProviderSet,
		domainapp.ProviderSet,
		newApp,
	)
	return nil, nil, nil
}
