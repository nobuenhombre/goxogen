package cli

import (
	"log"

	"github.com/google/wire"
	"github.com/nobuenhombre/suikat/pkg/ge"
)

// ProviderSet exports Wire providers for the cli package.
var ProviderSet = wire.NewSet(
	ProvideCLI,
)

// ProvideCLI parses CLI flags and returns a cleanup-aware service.
func ProvideCLI() (Service, func(), error) {
	cleanup := func() {
		log.Println("CLI config cleanup")
	}

	cfg, err := New()
	if err != nil {
		return nil, cleanup, ge.Pin(err)
	}

	return cfg, cleanup, nil
}
