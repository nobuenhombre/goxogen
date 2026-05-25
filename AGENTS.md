# goxogen — Code Generation Scaffolder

## Overview

- **Purpose:** multi-tool monorepo (goxogen, gobp, xouid) for Go code generation, build progress display, and PostgreSQL XOID query scaffolding.
- **Domain:** code scaffolding, build tooling, SQL-to-Go generation
- **Module:** `goxogen` (Go 1.26.1)

## Архитектура

```
goxogen/
├── AGENTS.md
├── src/
│   ├── cmd/
│   │   ├── goxogen/        # scaffolder CLI (code generation templates)
│   │   ├── gobp/           # build pipeline with progress bar
│   │   └── xouid/          # PostgreSQL XOID query generator
│   └── internal/
│       └── app/
│           ├── goxogen/    # config + cli + domain + log + version
│           ├── gobp/       # cli + domain + version
│           └── xouid/      # cli + domain + postgres + version
├── service/
│   └── deployments/        # per-app Linux build/install Makefiles
├── bin/                    # compiled binaries (gitkeep'd)
├── gobp                   # compiled gobp binary (project uses it for build-app-progress)
├── goxogen                # compiled goxogen binary
├── xouid                  # compiled xouid binary
├── Makefile               # root build targets
└── go.mod                 # Go 1.26.1
```

## Технологический стек

| Компонент | Технология | Версия |
|-----------|-----------|--------|
| Язык | Go | 1.26.1 |
| DI | Google Wire | v0.7.0 |
| CLI-парсинг | nobuenhombre/suikat/pkg/clivar | v0.0.170 |
| PostgreSQL | jackc/pgx/v5 | v5.9.2 |
| YAML | gopkg.in/yaml.v3 | v3.0.1 |
| Error-wrapping | suikat/pkg/ge (ge.Pin) | — |
| File I/O | suikat/pkg/fico | — |
| File utils | suikat/pkg/futi | — |

## Версионирование

- Source: `src/internal/app/{app}/version/version.go`
- Format: `vMAJOR.MINOR.PATCH` (SemVer)
- Flag: `-version` / `--version` (before Wire init)

| App | Current version |
|-----|----------------|
| goxogen | v0.1.0 |
| gobp | v0.4.0 |
| xouid | v0.1.0 |

## CLI-флаги

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
- **Wire-бинарник:** goxogen (gobp) + goxogen (scaffolder) + xouid — три отдельных `main` в `src/cmd/`
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
- Установка в `/usr/local/bin/{app}` + `/var/log/{app}/`

## Тестирование

```bash
go test ./... -v
```

- Единственный тест: `config-app_test.go` (Load/Save круг)
- Test fixtures: `config-app_test_load.yaml`, `config-app_test_save.yaml`

## Gotchas

- **`go vet файл.go` даёт `undefined: Service`** в provider.go — всегда использовать `go vet ./...`
- **Версионирование:** только через `-version` перед Wire init — после Wire подключена БД/лог
- **Wire cleanup:** порядок cleanup обратный порядку создания — важно для закрытия подключений
- **gobp прогресс-бар:** `\r` конфликтует с выводом ошибок компилятора — печатать `\n` перед строками ошибок
- **xouid templates:** ожидает файлы `xouid_package.go.tpl` и `xouid_query.go.tpl` в пути `-template-path`
- **bin/.gitkeep:** директории бинарников не удалять (биндинги в deployment Makefile)
