package main

import (
	"log"

	"goxogen/src/internal/app/xouid/domain"
)

// IApp is the top-level application orchestrator interface for xouid.
type IApp interface {
	Run() error
}

// App is the top-level application orchestrator for xouid.
type App struct {
	dom domainapp.DomainService
}

// Run executes the xouid code generation pipeline.
func (a *App) Run() error {
	return a.dom.Run()
}

// newApp is the Wire provider for the top-level xouid application.
func newApp(dom domainapp.DomainService) (IApp, func(), error) {
	cleanup := func() {
		log.Println("[wire-cleanup] Xouid App cleanup")
	}

	return &App{dom: dom}, cleanup, nil
}
