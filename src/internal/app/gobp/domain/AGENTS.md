# domain — Business logic for gobp

## Назначение

Запуск `go build` с прогресс-баром: подсчёт шагов (dry-run), выполнение, отображение прогресса.

## Файлы

| Файл | Назначение |
|------|------------|
| `domain.go` | `DomainService` + `AppDomain` + `countSteps()` + `Run()` |
| `provider.go` | `ProviderSet` + `ProvideDomain(cli.Service)` |

## Ключевые типы

- **DomainService** — `{Run() error}`
- Render functions и **ProgressState** — из `goxogen/src/internal/pkg/progress-bar`

## Алгоритм

1. Сбор `commonArgs`: `["build", "-o", out, binary]` (+ `-a` если full-rebuild)
2. `countSteps()`: `go build -n`, фильтр командных строк (mkdir, gcc, compile, link, cd, cat)
3. Если 0 шагов — быстрая сборка (up-to-date)
4. `go build -x`, парсинг командных строк (такой же фильтр) + `.go:` ошибки
5. Рендер прогресс-бара после каждой команды
6. После завершения — финальная линия + список ошибок

## Правила изменения

- Не добавлять `-ldflags` в `commonArgs` (линковка через `GOFLAGS`)
- `\r` коллизия: при выводе ошибок печатать `\n` перед строкой
- **Scanner Buffer:** в `countSteps()` и прогресс-цикле обязательно `scanner.Buffer(make([]byte, 64*1024), 1024*1024)`. Без этого длинные строки от gcc/link (>64KB) убивают сканер, `go build -n` виснет на write из-за переполненного pipe.
- **Фильтр команд:** `isBuildCommandLine()` считает все командные строки (mkdir, /путь/компилятор, cd, cat), а не только `mkdir`. Иначе прогресс-бар показывает 100% через 1 секунду, а компиляция с CGO идёт ещё минуты.
- Тот же фильтр — и в `countSteps()` (подсчёт), и в прогресс-цикле (отрисовка)
- **Расхождение dry-run vs реальный цикл:** `go build -x` может выдать на 1–5 строк больше, чем `go build -n`. Причина — CGO: реальный gcc/clang пишет диагностику (Assembler messages, Warning'и) в stderr, которая тоже начинается с `/` и проходит фильтр `isBuildCommandLine()`. Вторичные проявления: прогресс-бар мог показать 101% и ETA становился отрицательным. Фикс: `ProgressBar()` в `progress-bar` ограничивает percent ≤ 100 и filled ≤ barLength; `domain.go` зажимает `step` до `totalSteps` при инкременте.
- ANSI-константы и render-функции в `goxogen/src/internal/pkg/progress-bar`, не дублировать
