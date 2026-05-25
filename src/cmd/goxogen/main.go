package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"goxogen/src/internal/app/goxogen/version"
)

func main() {
	// 1. Panic recovery
	defer func() {
		r := recover()
		if r != nil {
			log.Printf("PANIC: %v\nStack trace: %s", r, debug.Stack())
		}
	}()

	// 2. Fast version check before heavy Wire initialization
	for _, arg := range os.Args[1:] {
		if arg == "-version" || arg == "--version" {
			fmt.Println(version.Version)
			os.Exit(0)
		}
	}

	// 3. Wire initialization (all dependencies created here)
	app, cleanup, err := initializeApp()
	if err != nil {
		log.Fatalf("Initialization error: %v", err)
	}
	defer cleanup()

	err = app.Run()
	if err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
