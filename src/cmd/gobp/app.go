package main

import (
	"log"

	"goxogen/src/internal/app/gobp/domain"
)

// IApp is the top-level application orchestrator interface for gobp.
type IApp interface {
	Run() error
}

// App is the top-level application orchestrator for gobp.
type App struct {
	dom domainapp.DomainService
}

// Run executes the gobp build progress pipeline.
func (a *App) Run() error {
	return a.dom.Run()
}

// newApp is the Wire provider for the top-level gobp application.
func newApp(dom domainapp.DomainService) (IApp, func(), error) {
	cleanup := func() {
		log.Println("[wire-cleanup] Gobp App cleanup")
	}

	return &App{dom: dom}, cleanup, nil
}
