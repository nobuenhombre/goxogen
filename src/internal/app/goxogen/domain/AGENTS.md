# domain — Business logic for goxogen

## Назначение

Оркестратор бизнес-логики goxogen. Запускает полный 8-шаговый XO pipeline
с прогресс-баром и отображением текущего обрабатываемого файла.

## Файлы

| Файл | Назначение |
|------|------------|
| `domain-app.go` | `DomainService` interface + `AppDomain` + `New(cliConfig)` + `countPipelineSteps()` |
| `xo-gen.go` | `Run()` — 8-шаговый pipeline генерации + все step-функции |
| `xo-config.go` | `XOConfig` — YAML-конфиг для XO, разбор connection string |
| `provider.go` | `ProviderSet` + `ProvideDomain(cli.Service)` |
| `templates.go` | Embedded XO-шаблоны + `a-db-repo.go.tpl` (`//go:embed templates/*.tpl`), экспорт `TemplatesDir()` |
| `templates/` | 17 файлов `.tpl` — 16 XO-шаблонов для генерации Go-кода (Postgres, MSSQL, MySQL, Oracle, xo_db, xouid) + `a-db-repo.go.tpl` для шага 7 |
| `step-7-generate-db-repo.go` | Шаг 7: генерация `a-db-repo.go` через embedded шаблон `a-db-repo.go.tpl` (`text/template`)

## Ключевые типы

- **DomainService** — `{Run() error}`
- **AppDomain** — хранит `Cli *cli.Config`
- **ProgressTracker** (из `goxogen/src/internal/pkg/progress-bar`) — управляет жизненным циклом прогресс-бара: `Increment(file)`, `AddError(line)`, `Finish()`, `Fail()`

## Pipeline (8 шагов)

1. **runXO** — удаление старых файлов, xo basic, xo queries (one/many/uid), удаление sp_*
2. **replaceInterfaceToAny** — замена `interface{}` на `any` во всех .go
3. **glueXoXouid** — слияние .xo.go + .xouid.go → .xo-xouid.go
4. **extractRepo** — извлечение @repo блоков в *-repo.xo.go
5. **removeXoXouid** — удаление .xo-xouid.go
6. **cleanXoXouidSourceBlocks** — очистка @repo маркеров из исходников
7. **generateDbRepo** — генерация агрегатного Db{Name}Repo (a-db-repo.go) через embedded шаблон `a-db-repo.go.tpl` (`text/template`) с конструктором, создающим подключение к БД через `pgxdb.NewDB`, и методом `Close()`
8. **goFormatCode** — go fmt, goimports, go vet (с захватом ошибок)

## Прогресс-бар

- Pre-count шагов: 12 фиксированных + количество SQL-файлов в queries/{one,many,uid}
- Каждый шаг показывает название операции в прогресс-баре
- Под шкалой прогресса — имя текущего обрабатываемого файла
- `goFormatCode` перехватывает stdout/stderr subprocess'ов, ошибки показываются после `ErrorLine` как в gobp
- При ошибке любого шага: `pt.Fail()` выводит `✖ N Errors detected / ⚠ Stopped after N seconds`

## Wire-интеграция

```
ProvideDomain(cli.Service) → (DomainService, func(), error)
```

## Правила изменения

- `Run()` — точка для добавления новых команд (через `-runtype`)
- `New()` принимает `cli.Service`, делает type assertion к `*cli.Config`
- Все step-функции, работающие с файлами, принимают `pt *ProgressTracker`
- ANSI-константы и символы в `goxogen/src/internal/pkg/progress-bar`, не дублировать
- `\r` в прогресс-баре конфликтует с выводом ошибок — `Fail()` печатает `\n` перед error lines
- `countPipelineSteps()` должен синхронизироваться с количеством `pt.Increment()` вызовов в pipeline
