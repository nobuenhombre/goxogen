# goxogen — Scaffolder (code generation)

## Назначение

Генератор Go-кода по шаблонам. CLI-утилита, которая читает YAML-конфиг, парсит CLI аргументы и выполняет кодогенерацию.

## Файлы

| Файл | Назначение |
|------|------------|
| `cli/cli.go` | CLI флаги: `-runtype`, `-config`, `-log`, `-version`. Загружаются через `clivar.Load()` |
| `cli/provider.go` | `cli.ProviderSet` + `ProvideCLI()` |
| `config/config-app.go` | YAML config loader/saver через `yaml.v3` |
| `config/provider.go` | `configapp.ProviderSet` + `ProvideConfigApp()` |
| `domain/domain-app.go` | `DomainService{}.Run()` — бизнес-логика XO-генерации |
| `domain/provider.go` | `domainapp.ProviderSet` + `ProvideDomain()` |
| `domain/templates.go` | Embedded XO-шаблоны через `//go:embed templates/*.tpl` |
| `domain/templates/` | 16 `.tpl` файлов XO-шаблонов (Postgres, MSSQL, MySQL, Oracle, xo_db, xouid) |
| `log/log-file.go` | `ILogFile` — перенаправление `log` в файл |
| `log/provider.go` | `logfile.ProviderSet` + `ProvideLogFile()` |
| `version/version.go` | `const Version = "v0.4.0"` |

## Ключевые типы

- **cli.Config** — структура с тегами `cli:\"flag[desc]:type=default\"`
- **configapp.Config** — структура с YAML-полями (пока пустая)
- **domainapp.DomainService** — интерфейс `{ Run() error }`
- **logfile.LogFile** — файловый логгер (перенаправляет `log.SetOutput`)

## Wire-интеграция

```
cli.ProviderSet
    → configapp.ProviderSet    (depends on cli.Service)
    → logfile.ProviderSet      (depends on cli.Service)
    → domainapp.ProviderSet    (depends on cli.Service)
    → newApp                   (depends on domainapp.DomainService)
```

## Правила изменения

- `DomainService.Run()` — точка расширения: сюда добавлять новые команды
- При добавлении новой секции в config — обновить `Config{}` и `config-app_test.go`
- `version/version.go` — единственный источник версии
