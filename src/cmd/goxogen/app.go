package main

import (
	"log"

	domainapp "goxogen/src/internal/app/goxogen/domain"
)

// IApp is the top-level application orchestrator interface.
type IApp interface {
	Run() error
}

// App is the top-level application orchestrator.
type App struct {
	dom domainapp.DomainService
}

// Run executes the application based on CLI configuration.
func (a *App) Run() error {
	return a.dom.Run()
}

// newApp is the Wire provider for the top-level application.
func newApp(dom domainapp.DomainService) (IApp, func(), error) {
	cleanup := func() {
		log.Println("[wire-cleanup] App cleanup")
	}

	return &App{dom: dom}, cleanup, nil
}
