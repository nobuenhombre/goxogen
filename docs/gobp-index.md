# Карта документации: gobp

## Общее

| Раздел | Файл |
|--------|------|
| Бизнес-модель | [gobp-business-model.md](gobp-business-model.md) |
| Карта goxogen | [goxogen-index.md](goxogen-index.md) |
| Карта xouid | [xouid-index.md](xouid-index.md) |

## Приложение — точка входа

| Раздел | Файл | Описание |
|--------|------|----------|
| Точка входа (`src/cmd/gobp`) | [cmd/gobp/gobp.md](cmd/gobp/gobp.md) | main.go → panic recovery → `-version` → Wire → `DomainService.Run()` |
| CLI-конфигурация | [internal/app/gobp/cli/cli.md](internal/app/gobp/cli/cli.md) | Флаги `-binary` / `-out` / `-verbose` / `-full-rebuild` / `-version`; интерфейс с геттерами |
| Domain — прогресс сборки | [internal/app/gobp/domain/domain.md](internal/app/gobp/domain/domain.md) | Dry-run подсчёт шагов (`go build -n`) → реальная сборка (`go build -x`) с прогресс-баром |
| Версия | [internal/app/gobp/version/version.md](internal/app/gobp/version/version.md) | `v0.9.0` (const, источник истины — `version.go`) |

## Общие пакеты

| Пакет | Файл | Описание |
|-------|------|----------|
| progress-bar | [internal/pkg/progress-bar/progress-bar.md](internal/pkg/progress-bar/progress-bar.md) | Общий ANSI-прогресс-бар; gobp использует `ProgressState` + render-функции напрямую |

## Смежные приложения

| Приложение | Файл | Описание |
|------------|------|----------|
| goxogen | [goxogen-index.md](goxogen-index.md) | XO-пайплайн кодогенерации; deployment `build-app-progress` собирает его через gobp |
| xouid | [xouid-index.md](xouid-index.md) | SQL-to-Go генератор UPDATE/INSERT/DELETE |