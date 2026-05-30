# goxogen — Code Generation Scaffolder

[![Go Version](https://img.shields.io/badge/Go-1.26.1-blue)](https://go.dev/dl/)
[![Wire DI](https://img.shields.io/badge/DI-Google_Wire-green)](https://github.com/google/wire)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

**goxogen** is a multi-tool monorepo containing three CLI applications for Go code generation (XO pipeline), build progress visualization, and PostgreSQL query scaffolding.

## Applications

| Application | Version | Description |
|-------------|---------|-------------|
| **goxogen** | v0.16.0 | XO code generation pipeline — generates Go models and query functions from PostgreSQL schema with a 9-step automated pipeline |
| **gobp** | v0.7.0 | Build pipeline with progress bar — wraps `go build` with dry-run step counting, visual progress, ETA, and error aggregation |
| **xouid** | v0.1.0 | PostgreSQL XOID query generator — generates typed Go functions from UPDATE/INSERT/DELETE SQL with EXPLAIN validation |

---

## Architecture

```
goxogen/
├── AGENTS.md                 # Dev agent context (23 nested AGENTS.md files)
├── Makefile                  # Root build targets (deps, wire)
├── go.mod                    # Go module (Go 1.26.1)
├── src/
│   ├── cmd/
│   │   ├── goxogen/          # XO code generation pipeline CLI
│   │   ├── gobp/             # Build progress CLI
│   │   └── xouid/            # XOID query generator CLI
│   ├── internal/
│   │   ├── pkg/
│   │   │   └── progress-bar/ # Shared ANSI progress bar (used by goxogen + gobp)
│   │   └── app/
│   │       ├── goxogen/      # Config, CLI, domain (XO pipeline), log, version
│   │       ├── gobp/         # CLI, domain, version
│   │       └── xouid/        # CLI, domain, postgres, version
├── service/
│   └── deployments/          # Per-app Linux build & install Makefiles
├── bin/                      # Compiled binaries (gitkeep'd)
├── goxogen                   # Compiled goxogen binary
├── gobp                      # Compiled gobp binary
├── xouid                     # Compiled xouid binary
├── .idea/                    # GoLand/IntelliJ project configuration
└── LICENSE                   # Apache 2.0
```

Every package has its own AGENTS.md describing purpose, key types, Wire integration, and change rules (23 total).

## Technology Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.26.1 |
| DI Framework | Google Wire | v0.7.0 |
| CLI Parsing | nobuenhombre/suikat/pkg/clivar | v0.0.170 |
| PostgreSQL | jackc/pgx/v5 | v5.9.2 |
| YAML | gopkg.in/yaml.v3 | v3.0.1 |
| UUID | github.com/google/uuid | v1.6.0 |
| Error Wrapping | suikat/pkg/ge (ge.Pin) | — |
| File I/O | suikat/pkg/fico | — |
| File Utils | suikat/pkg/futi | — |
| DB Abstraction | suikat/pkg/db/connectors/postgres-pgx-db | — |

## Features

- **XO Code Generation Pipeline** — 9-step pipeline: schema → models → query functions (one/many/UID) → repo extraction → aggregate repo → Wire provider → formatting/vetting
- **Multi-DB Template Support** — 14 embedded templates for PostgreSQL, MSSQL, MySQL, and Oracle type generation
- **Build Progress Bar** — `gobp` wraps `go build` with dry-run step counting, ANSI progress bar, time/ETA display, and error aggregation
- **PostgreSQL Query Generation** — `xouid` validates SQL via EXPLAIN, parses `%%param type%%` descriptors, and generates typed Go query functions using Go templates
- **Google Wire DI** — clean dependency injection across all three apps with proper cleanup ordering
- **Shared Progress Bar** — common ANSI progress bar package used by both goxogen and gobp

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

## XO Code Generation Pipeline

goxogen runs its 9-step code generation pipeline when executed with `-config=config.yaml`. It connects to a PostgreSQL database, generates models and query functions, extracts repository interfaces, and produces clean, formatted Go code.

### Config YAML Structure

```yaml
config:
  db:
    host: localhost
    port: 5432
    name: dbname
    user: user
    pass: pass
    sslmode: disable
    pool_max_conns: 10
    backups:
      path: /path/to/backups
  codegen:
    path: ./gen          # output directory for generated code
    package: gen         # Go package name for generated code
    queries: ./queries   # SQL queries directory
    ignore_fields: created_at,updated_at  # fields to skip in generation
    db_name: Mydb                         # optional, overrides db.name for Db{Name}Repo struct
```

### Pipeline Steps

| Step | Action | Description |
|------|--------|-------------|
| 1 | **runXO** | Deletes old `.xo.go` / `.xouid.go`, runs xo for models + queries (one/many/uid), removes stored procedures |
| 2 | **replaceInterfaceToAny** | Replaces `interface{}` with `any` in all generated files |
| 3 | **glueXoXouid** | Merges `.xo.go` + `.xouid.go` → `.xo-xouid.go` temporary files |
| 4 | **extractRepo** | Extracts `@repo-start`/`@repo-end` blocks → `*-repo.xo.go` repository files |
| 5 | **removeXoXouid** | Deletes temporary `.xo-xouid.go` files |
| 6 | **cleanXoXouidSourceBlocks** | Removes `@repo-start`/`@repo-end` markers from `.xo.go` and `.xouid.go` |
|| 7 | **generateDbRepo** | Scans `*-repo.xo.go`, generates `a-db-repo.go` via embedded `a-db-repo.go.tpl` template (`text/template`) with aggregate `Db{DbName}Repo` struct + `NewDb{DbName}Repository` constructor (creates DB connection internally via `pgxdb.NewDB`) + `Close()` method |
| 8 | **generateProvider** | Generates `provider.go` with Wire `ProviderSet`, exposes `Provider{DbName}` with cleanup |
| 9 | **goFormatCode** | Runs `go fmt`, `goimports -w`, `go vet` for clean output |

SQL query files follow the naming convention `TypeName-FuncName.sql` and are organized in subdirectories:
- `queries/one/` — single-row queries
- `queries/many/` — multi-row queries
- `queries/uid/` — UPDATE/INSERT/DELETE queries (processed by xouid)

### Embedded Templates

15 Go templates are embedded into the goxogen binary via `//go:embed templates/*.tpl`. 14 XO templates are extracted at runtime for the xo CLI. The `a-db-repo.go.tpl` template is loaded directly from the embedded FS via `text/template` for step 7.

| Template | Purpose |
|----------|---------|
| `mssql.type.go.tpl` | MSSQL type generation |
| `mysql.type.go.tpl` | MySQL type generation |
| `oracle.type.go.tpl` | Oracle type generation |
| `postgres.enum.go.tpl` | PostgreSQL enum generation |
| `postgres.foreignkey.go.tpl` | PostgreSQL foreign key queries |
| `postgres.index.go.tpl` | PostgreSQL index queries |
| `postgres.proc.go.tpl` | PostgreSQL stored procedures |
| `postgres.query.go.tpl` | PostgreSQL query functions |
| `postgres.querytype.go.tpl` | PostgreSQL query type helpers |
| `postgres.type.go.tpl` | PostgreSQL type generation |
| `xo_db.go.tpl` | Database connection wrapper |
| `xo_package.go.tpl` | Package header |
| `xouid_package.go.tpl` | XOID package header |
|| `xouid_query.go.tpl` | XOID query function |
|| `a-db-repo.go.tpl` | Aggregate DbRepo struct template (used by generateDbRepo) |

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
- All `Provide*` functions return `(T, func(), error)` for proper cleanup chaining

## Shared Packages

### progress-bar (`src/internal/pkg/progress-bar/`)

Shared ANSI progress bar used by both goxogen and gobp. Provides:

- **ProgressState** — bar state with title, project name, current/total, elapsed time, ETA, error count
- **ProgressTracker** — lifecycle manager: Increment, AddError, Finish, Fail
- Unicode progress bar with 8 ANSI colors, 50-character bar length
- ETA calculation with remaining time estimate

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
- Installed via symlink to `/usr/local/bin/{app}`

## Development

### Conventions

- **Errors:** wrap with `ge.Pin(err)` from `suikat/pkg/ge`
- **Cleanup:** all `Provide*` functions return `(T, func(), error)` — Wire invokes cleanup in reverse order
- **Aliases:** conflicting package names are aliased (e.g., `domainapp`, `pgxdb`, `configapp`, `logfile`)
- **Versioning:** only via `-version` flag before Wire init (avoids unnecessary DB/log connections)
- **Linking:** ldflags via `GOFLAGS` environment variable, not hardcoded
- **Nested AGENTS.md:** read the relevant AGENTS.md before modifying any package

### Common Pitfalls

- `go vet file.go` produces `undefined: Service` in provider.go — always use `go vet ./...`
- gobp progress bar `\r` conflicts with compiler error output — print `\n` before error lines
- gobp scanner buffer must be at least 1MB (`scanner.Buffer(make([]byte, 64*1024), 1024*1024)`) for long CGO lines
- gobp dry-run (`go build -n`) may count fewer steps than actual (`go build -x`) due to CGO diagnostics — progress bar caps at 100%
- xouid templates expect `xouid_package.go.tpl` (template name: `xouidpackage`) and `xouid_query.go.tpl` (template name: `xouidquery`) in `-template-path`
- xouid SQL parameters use `%%paramName type%%` format (types: int, int32, int64, float, float32, float64, string, bool, uuid.UUID, time.Time, arrays)
- `bin/.gitkeep`: don't remove binary directories (deployment Makefiles reference them)

## License

Apache 2.0 — see [LICENSE](LICENSE)