package configapp

import (
	"log"

	"github.com/google/wire"
	"github.com/nobuenhombre/suikat/pkg/ge"
	"goxogen/src/internal/app/goxogen/cli"
)

// ProviderSet exports Wire providers for the configapp package.
var ProviderSet = wire.NewSet(
	ProvideConfigApp,
)

// ProvideConfigApp loads the YAML configuration from the CLI config file path.
func ProvideConfigApp(cliConfig cli.Service) (Service, func(), error) {
	cleanup := func() {
		log.Println("App config cleanup")
	}

	cfg, err := New(cliConfig.(*cli.Config).Config)
	if err != nil {
		return nil, cleanup, ge.Pin(err)
	}

	return cfg, cleanup, nil
}
