# gobp — Build Progress Pipeline

## Назначение

CLI-утилита для отображения прогресса `go build` с прогресс-баром. Запускает `go build -n` для подсчёта шагов (mkdir), затем `go build -x` с отображением прогресса.

## Файлы

| Файл | Назначение |
|------|------------|
| `cli/cli.go` | CLI флаги: `-binary`, `-out`, `-verbose`, `-full-rebuild` |
| `cli/provider.go` | `cli.ProviderSet` + `ProvideCLI()` |
| `domain/domain.go` | `DomainService{}.Run()` — подсчёт шагов + прогресс-бар |
| `domain/provider.go` | `domainapp.ProviderSet` + `ProvideDomain()` |
| `version/version.go` | `const Version = "v0.4.0"` |

## Ключевые типы

- **cli.Config** — структура с геттерами `GetBinary()`, `GetOut()`, `GetVerbose()`, `GetFullRebuild()`
- **cli.Service** — интерфейс (абстракция для Wire)
- **domainapp.AppDomain** — реализует `DomainService`, управляет прогресс-баром
- **ProgressState** — из `goxogen/src/internal/pkg/progress-bar`, не дублировать

## Алгоритм

1. `countSteps()` — `go build -n` (dry-run), считает строки с `mkdir`
2. Если шагов 0 — бинарник актуален в кеше, просто `go build`
3. Иначе `go build -x`, парсит `mkdir` строки, рендерит прогресс-бар
4. Собирает `.go:` строки как ошибки, выводит их после `ErrorLine`

## Wire-интеграция

```
cli.ProviderSet → domainapp.ProviderSet → newApp
```

## Правила изменения

- `commonArgs` не должен содержать `-ldflags` (линковка через `GOFLAGS` envvar)
- `\r` в прогресс-баре конфликтует с выводом ошибок — печатать `\n` перед error lines
- При добавлении нового флага — добавить геттер в `Service` интерфейс
