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
| `domain/templates/` | 18 `.tpl` файлов — XO-шаблоны + `a-db-repo.go.tpl` для шага 7 + `provider.go.tpl` для шага 8 |
| `domain/step-7-generate-db-repo.go` | Шаг 7: генерация `a-db-repo.go` через embedded шаблон `a-db-repo.go.tpl` (`text/template`) |
| `domain/step-8-generate-provider.go` | Шаг 8: генерация `provider.go` через embedded шаблон `provider.go.tpl` (`text/template`) |
| `domain/xo-gen.go` | 9-шаговый пайплайн XO-генерации (runXO, replaceInterfaceToAny, glueXoXouid, extractRepo, removeXoXouid, cleanXoXouidSourceBlocks, generateDbRepo, generateProvider, goFormatCode) |
| `domain/xo-config.go` | Конфигурация XO-пайплайна, загрузка YAML, resolveDbName |
| `log/log-file.go` | `ILogFile` — перенаправление `log` в файл |
| `log/provider.go` | `logfile.ProviderSet` + `ProvideLogFile()` |
| `version/version.go` | `const Version = "v0.16.0"` |

## Ключевые типы

- **cli.Config** — структура с тегами `cli:"flag[desc]:type=default"`
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
