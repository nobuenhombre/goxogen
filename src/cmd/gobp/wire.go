//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"goxogen/src/internal/app/gobp/cli"
	domainapp "goxogen/src/internal/app/gobp/domain"
)

// initializeApp is the Wire injector entrypoint for gobp.
// It aggregates all ProviderSets and constructs the top-level application.
func initializeApp() (IApp, func(), error) {
	wire.Build(
		cli.ProviderSet,
		domainapp.ProviderSet,
		newApp,
	)
	return nil, nil, nil
}
