# domain — Business logic for gobp

## Назначение

Запуск `go build` с прогресс-баром: подсчёт шагов (dry-run), выполнение, отображение прогресса.

## Файлы

| Файл | Назначение |
|------|------------|
| `domain.go` | `DomainService` + `AppDomain` + `countSteps()` + `Run()` |
| `types.go` | `ProgressState`, ANSI color constants, render functions |
| `provider.go` | `ProviderSet` + `ProvideDomain(cli.Service)` |

## Ключевые типы

- **DomainService** — `{Run() error}`
- **ProgressState** — `{Title, ProjectName, Current, Total, Elapsed, Remaining, Errors, StartTime}`
- Render functions: `StartLine()`, `ProgressBar()`, `FinishLine()`, `ErrorLine()`

## Алгоритм

1. Сбор `commonArgs`: `["build", "-o", out, binary]` (+ `-a` если full-rebuild)
2. `countSteps()`: `go build -n`, фильтр `mkdir`
3. Если 0 шагов — быстрая сборка (up-to-date)
4. `go build -x`, парсинг `mkdir` + `.go:` ошибки
5. Рендер прогресс-бара после каждого `mkdir`
6. После завершения — финальная линия + список ошибок

## Правила изменения

- Не добавлять `-ldflags` в `commonArgs` (линковка через `GOFLAGS`)
- `\r` коллизия: при выводе ошибок печатать `\n` перед строкой
- ANSI-константы в `types.go`, не дублировать
