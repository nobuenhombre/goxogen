# goxogen — Code Generation Scaffolder

[![Go Version](https://img.shields.io/badge/Go-1.26.1-blue)](https://go.dev/dl/)
[![Wire DI](https://img.shields.io/badge/DI-Google_Wire-green)](https://github.com/google/wire)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**goxogen** is a multi-tool monorepo containing three CLI applications for Go code generation, build progress visualization, and PostgreSQL query scaffolding.

## Applications

| Application | Version | Description |
|-------------|---------|-------------|
| **goxogen** | v0.1.0 | Code generation scaffolder — generates Go code from YAML templates |
| **gobp** | v0.4.0 | Build pipeline with progress bar — wraps `go build` with visual progress |
| **xouid** | v0.1.0 | PostgreSQL XOID query generator — generates typed Go query functions from SQL |

---

## Architecture

```
goxogen/
├── AGENTS.md                 # Dev agent context
├── Makefile                  # Root build targets
├── go.mod                    # Go module (Go 1.26.1)
├── src/
│   ├── cmd/
│   │   ├── goxogen/          # Scaffolder CLI
│   │   ├── gobp/             # Build progress CLI
│   │   └── xouid/            # XOID query generator CLI
│   └── internal/
│       └── app/
│           ├── goxogen/      # Config, CLI, domain, log, version
│           ├── gobp/         # CLI, domain, version
│           └── xouid/        # CLI, domain, postgres, version
├── service/
│   └── deployments/          # Per-app Linux build & install Makefiles
├── bin/                      # Compiled binaries (gitkeep'd)
├── goxogen                   # Compiled goxogen binary
├── gobp                      # Compiled gobp binary
└── xouid                     # Compiled xouid binary
```

## Technology Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.26.1 |
| DI Framework | Google Wire | v0.7.0 |
| CLI Parsing | nobuenhombre/suikat/pkg/clivar | v0.0.170 |
| PostgreSQL | jackc/pgx/v5 | v5.9.2 |
| YAML | gopkg.in/yaml.v3 | v3.0.1 |
| Error Wrapping | suikat/pkg/ge (ge.Pin) | — |
| File I/O | suikat/pkg/fico | — |
| File Utils | suikat/pkg/futi | — |

## Getting Started

### Prerequisites

- Go 1.26.1+
- Google Wire CLI (`go install github.com/google/wire/cmd/wire@latest`)

### Build

```bash
make deps          # Re-init go.mod, download all dependencies
make wire          # Regenerate wire_gen.go for all apps

# Build individual apps
go build ./src/cmd/goxogen
go build ./src/cmd/gobp
go build ./src/cmd/xouid
```

### Test

```bash
go test ./... -v
```

## CLI Flags

### goxogen

| Flag | Default | Description |
|------|---------|-------------|
| `-runtype` | `init` | Run type |
| `-config` | `config.yaml` | Path to YAML config |
| `-log` | `""` | Path to log file |
| `-version` | `false` | Show version and exit |

### gobp

| Flag | Default | Description |
|------|---------|-------------|
| `-binary` | `.` | Go binary to build (e.g. `./src/cmd/myapp`) |
| `-out` | `./build/app` | Output binary path |
| `-verbose` | `false` | Show full build output |
| `-full-rebuild` | `false` | Force full rebuild (`go build -a`) |
| `-version` | `false` | Show version and exit |

### xouid

| Flag | Default | Description |
|------|---------|-------------|
| `-out` | `""` | Output path |
| `-dsn` | `""` | PostgreSQL DSN |
| `-template-path` | `""` | Template directory path |
| `-package` | `""` | Package name in generated Go code |
| `-schema` | `public` | Schema name |
| `-query-type` | `""` | Query type (generated as `{type}.xouid.go`) |
| `-query-func` | `""` | Generated Go func name |
| `-query` | `""` | Query file path |
| `-verbose` | `false` | Don't show hello message |
| `-version` | `false` | Show version and exit |

## Dependency Injection (Google Wire)

Each application uses Google Wire for dependency injection. Wire DI graph:

```
goxogen:  cli ProviderSet → domain ProviderSet → App
gobp:     cli ProviderSet → domain ProviderSet → App
xouid:    cli ProviderSet → postgres ProviderSet → domain ProviderSet → App
```

- All `provider.go` files export `var ProviderSet = wire.NewSet(ProvideXxx)`
- `wire.go` in `package main`: only `wire.Build()`, no application logic
- `wire_gen.go` is auto-generated — never edit manually

## Deployment

Per-app deployment Makefiles are in `service/deployments/{app}/linux/`:

```bash
# Build + install one app
cd service/deployments/goxogen/linux && make all

# Build with progress bar (requires gobp binary)
cd service/deployments/goxogen/linux && make build-app-progress
```

Build settings:
- `CGO_ENABLED=0`, `GOOS=linux`, `GOARCH=amd64`
- ldflags `-s -w` for smaller binaries
- Installed to `/usr/local/bin/{app}` with logs at `/var/log/{app}/`

## Development

### Conventions

- **Errors:** wrap with `ge.Pin(err)` from `suikat/pkg/ge`
- **Cleanup:** all `Provide*` functions return `(T, func(), error)` — Wire invokes cleanup in reverse order
- **Aliases:** conflicting package names are aliased (e.g., `domainapp`, `pgxdb`, `configapp`, `logfile`)
- **Versioning:** only via `-version` flag before Wire init (avoids unnecessary DB/log connections)
- **Linking:** ldflags via `GOFLAGS` environment variable, not hardcoded

### Common Pitfalls

- `go vet file.go` produces `undefined: Service` in provider.go — always use `go vet ./...`
- gobp progress bar `\r` conflicts with compiler error output — print `\n` before error lines
- xouid templates expect `xouid_package.go.tpl` and `xouid_query.go.tpl` in `-template-path`
- `bin/.gitkeep`: don't remove binary directories (deployment Makefiles reference them)
