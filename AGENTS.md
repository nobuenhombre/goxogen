# goxogen — Code Generation Scaffolder

## Overview

- **Purpose:** multi-tool monorepo (goxogen, gobp, xouid) for Go code generation, build progress display, and PostgreSQL XOID query scaffolding.
- **Domain:** code scaffolding, build tooling, SQL-to-Go generation
- **Module:** `goxogen` (Go 1.26.1)

## Архитектура

```
goxogen/
├── AGENTS.md                   # This file — dev agent context
├── src/
│   ├── AGENTS.md               # Root src agents context
│   ├── cmd/
│   │   ├── AGENTS.md           # cmd agents context
│   │   ├── goxogen/            # scaffolder CLI (code generation templates)
│   │   │   ├── AGENTS.md
│   │   │   ├── main.go
│   │   │   ├── app.go
│   │   │   ├── wire.go
│   │   │   └── wire_gen.go
│   │   ├── gobp/               # build pipeline with progress bar
│   │   │   ├── AGENTS.md
│   │   │   ├── main.go
│   │   │   ├── app.go
│   │   │   ├── wire.go
│   │   │   └── wire_gen.go
│   │   └── xouid/              # PostgreSQL XOID query generator
│   │       ├── AGENTS.md
│   │       ├── main.go
│   │       ├── app.go
│   │       ├── wire.go
│   │       └── wire_gen.go
│   └── internal/
│       ├── AGENTS.md           # Internal packages context
│       └── app/
│           ├── goxogen/        # config + cli + domain + log + version
│           │   ├── AGENTS.md
│           │   ├── cli/        # CLI flag parsing (runtype, config, log, version)
│           │   ├── config/     # YAML config load/save (app-level)
│           │   ├── domain/     # Business logic: XO pipeline + app run
│           │   ├── log/        # Log file redirection
│           │   └── version/    # v0.1.0
│           ├── gobp/           # cli + domain + version
│           │   ├── AGENTS.md
│           │   ├── cli/        # CLI flags: binary, out, verbose, full-rebuild
│           │   ├── domain/     # Build progress bar with dry-run counting
│           │   └── version/    # v0.6.0
│           └── xouid/          # cli + domain + postgres + version
│               ├── AGENTS.md
│               ├── cli/        # CLI flags: out, dsn, template-path, etc.
│               ├── domain/     # SQL-to-Go generation with EXPLAIN validation
│               ├── postgres/   # pgx/v5 connection pool
│               └── version/    # v0.1.0
├── service/
│   ├── AGENTS.md               # Service infrastructure context
│   └── deployments/
│       ├── AGENTS.md           # Deployments context
│       ├── goxogen/linux/Makefile
│       ├── gobp/linux/Makefile
│       └── xouid/linux/Makefile
├── bin/                        # Compiled binaries (gitkeep'd)
├── gobp                        # Compiled gobp binary (used for build-app-progress)
├── goxogen                     # Compiled goxogen binary
├── xouid                       # Compiled xouid binary
├── Makefile                    # Root build targets: deps, wire
├── go.mod                      # Go 1.26.1
├── go.sum
├── LICENSE                     # Apache 2.0
├── README.md                   # English documentation
├── README.RU.md                # Russian documentation
└── .gitignore
```

**Nested AGENTS.md:** every `src/cmd/{app}/`, `src/internal/app/{app}/`, and every internal package has its own AGENTS.md describing purpose, files, key types, Wire integration, and change rules. Read the relevant nested AGENTS.md before modifying any package.

## Технологический стек

| Компонент | Технология | Версия |
|-----------|-----------|--------|
| Язык | Go | 1.26.1 |
| DI | Google Wire | v0.7.0 |
| CLI-парсинг | nobuenhombre/suikat/pkg/clivar | v0.0.170 |
| PostgreSQL | jackc/pgx/v5 | v5.9.2 |
| YAML | gopkg.in/yaml.v3 | v3.0.1 |
| UUID | github.com/google/uuid | v1.6.0 |
| Error-wrapping | suikat/pkg/ge (ge.Pin) | — |
| File I/O | suikat/pkg/fico | — |
| File utils | suikat/pkg/futi | — |
| DB abstraction | suikat/pkg/db/connectors/postgres-pgx-db | — |

## Версионирование

- Source: `src/internal/app/{app}/version/version.go`
- Format: `vMAJOR.MINOR.PATCH` (SemVer)
- Flag: `-version` / `--version` (before Wire init)

| App | Current version |
|-----|----------------|
| goxogen | v0.4.0 |
| gobp | v0.6.0 |
| xouid | v0.1.0 |

## CLI-флаги

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

## Wire DI граф

```
goxogen:  cli.ProviderSet → domainapp.ProviderSet → App
gobp:     cli.ProviderSet → domainapp.ProviderSet → App
xouid:    cli.ProviderSet → postgres.ProviderSet → domainapp.ProviderSet → App
```

- Все `provider.go` экспортируют `var ProviderSet = wire.NewSet(ProvideXxx)`
- `wire.go` в `package main`: только `wire.Build()`, никакой логики
- `wire_gen.go` — автосгенерирован, не редактировать вручную
- `go vet ./...` — избегать `go vet файл.go` (false positive `undefined: Service`)

## Conventions

- **Errors:** оборачивать через `ge.Pin(err)` из suikat/pkg/ge
- **Cleanup:** все Provide\* возвращают `(T, func(), error)` — Wire вызывает cleanup в обратном порядке
- **Aliases:** конфликтующие имена пакетов алиасить (domainapp, pgxdb, configapp, logfile)
- **Wire-бинарник:** goxogen + gobp + xouid — три отдельных `main` в `src/cmd/`
- **main.go:** panic recovery → fast version check → Wire init → Run
- **Линковка:** ldflags через `GOFLAGS` envvar (не hardcode в commonArgs)
- **Пути:** `src/internal/app/{app}/` для каждого приложения

## Команды

```bash
make deps          # Re-init go.mod, re-download all deps
make wire          # Regenerate wire_gen.go (all apps)
go build ./src/cmd/goxogen
go build ./src/cmd/gobp
go build ./src/cmd/xouid
```

## Deployment

Per-app deployment Makefiles in `service/deployments/{app}/linux/`:

```bash
# Build + install one app
cd service/deployments/goxogen/linux && make all

# Build with progress bar (requires gobp binary)
cd service/deployments/goxogen/linux && make build-app-progress
```

- CGO_ENABLED=0, GOOS=linux, GOARCH=amd64
- ldflags `-s -w` для уменьшения размера
- Установка в `/usr/local/bin/{app}` + симлинк через `sudo ln -sf`

## Тестирование

```bash
go test ./... -v
```

- Единственный тест: `config-app_test.go` (Load/Save круг)
- Test fixtures: `config-app_test_load.yaml`, `config-app_test_save.yaml`

## XO Code Generation Pipeline (goxogen -runtype=xo)

goxogen supports a full 8-step code generation pipeline when run with `-runtype=xo`:

1. **runXO** — Deletes old `.xo.go` and `.xouid.go` files, then runs:
   - `xo basic` — schema-based model generation
   - `xo queries one` — single-row query generation (from `queries/one/*.sql`)
   - `xo queries many` — multi-row query generation (from `queries/many/*.sql`)
   - `xo queries uid` — UPDATE/INSERT/DELETE via xouid (from `queries/uid/*.sql`)
   - Deletes `sp_*.xo.go` (stored procedure artifacts)
2. **replaceInterfaceToAny** — Replaces `interface{}` with `any` in all `.go` files
3. **glueXoXouid** — Merges `.xo.go` + `.xouid.go` → `.xo-xouid.go`
4. **extractRepo** — Extracts `@repo-start`/`@repo-end` blocks → `*-repo.xo.go`
5. **removeXoXouid** — Deletes temp `.xo-xouid.go` files
6. **cleanXoXouidSourceBlocks** — Removes `@repo-start`/`@repo-end` markers from `.xo.go` and `.xouid.go`
7. **generateDbRepo** — Scans `*-repo.xo.go`, generates `a-db-repo.go` with aggregate `Db{DbName}Repo` struct + `NewDb{DbName}Repository` constructor
8. **goFormatCode** — Runs `go fmt`, `goimports -w`, `go vet`

Config YAML structure (`-config config.yaml`):
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
    path: ./gen
    package: gen
    queries: ./queries
    ignore_fields: created_at,updated_at
    db_name: Mydb  # optional, overrides db.name for Db{Name}Repo struct
```

## Gotchas

- **Embedded шаблоны:** XO-шаблоны встроены в бинарник через `//go:embed` и извлекаются в `os.MkdirTemp` при запуске `-runtype=xo`. Не нужно указывать `templates` в конфиге — всегда используются шаблоны из `src/internal/app/goxogen/domain/templates/`
- **`go vet файл.go` даёт `undefined: Service`** в provider.go — всегда использовать `go vet ./...`
- **Версионирование:** только через `-version` перед Wire init — после Wire подключена БД/лог
- **Wire cleanup:** порядок cleanup обратный порядку создания — важно для закрытия подключений
- **gobp прогресс-бар:** `\r` конфликтует с выводом ошибок компилятора — печатать `\n` перед строками ошибок
- **gobp scanner buffer:** `scanner.Buffer(make([]byte, 64*1024), 1024*1024)` — обязателен в countSteps() и прогресс-цикле, иначе длинные строки gcc/link (>64KB) убивают сканер
- **gobp dry-run vs real:** `go build -x` может выдать на 1–5 строк больше, чем `go build -n` (CGO diagnostics). Фикс: `ProgressBar()` ограничивает percent ≤ 100 и filled ≤ barLength; `domain.go` зажимает `step` до `totalSteps`
- **xouid templates:** ожидает файлы `xouid_package.go.tpl` (template name: `xouidpackage`) и `xouid_query.go.tpl` (template name: `xouidquery`) в пути `-template-path`
- **xouid SQL params:** формат `%%paramName type%%` (2 parts: name + type); валидируется регуляркой `(%%(\s|\S)*?%%)`
- **xouid supports types:** int, int32, int64, float, float32, float64, string, bool, uuid.UUID, time.Time, []int/[]int32/[]int64
- **bin/.gitkeep:** директории бинарников не удалять (биндинги в deployment Makefile)