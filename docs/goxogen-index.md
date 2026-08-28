# Карта документации: goxogen

## Общее

| Раздел | Файл |
|--------|------|
| Бизнес-модель | [goxogen-business-model.md](goxogen-business-model.md) |
| Карта gobp | [gobp-index.md](gobp-index.md) |
| Карта xouid | [xouid-index.md](xouid-index.md) |

## Приложение — точка входа

| Раздел | Файл | Описание |
|--------|------|----------|
| Точка входа (`src/cmd/goxogen`) | [cmd/goxogen/goxogen.md](cmd/goxogen/goxogen.md) | main.go → panic recovery → `-version` → Wire → `DomainService.Run()` |
| CLI-конфигурация | [internal/app/goxogen/cli/cli.md](internal/app/goxogen/cli/cli.md) | Флаги `-config` / `-log` / `-version` через clivar |
| Конфиг YAML | [internal/app/goxogen/config/config.md](internal/app/goxogen/config/config.md) | Каркас Load/Save; реальная конфигурация пайплайна — в domain (`xo-config.go`) |
| Domain — XO-пайплайн | [internal/app/goxogen/domain/domain.md](internal/app/goxogen/domain/domain.md) | 10 шагов генерации (+1 readonly), `XOConfig`, 16 embedded-шаблонов |
| Лог-файл | [internal/app/goxogen/log/log.md](internal/app/goxogen/log/log.md) | Перенаправление stdlib-логгера в файл (флаг `-log`) |
| Версия | [internal/app/goxogen/version/version.md](internal/app/goxogen/version/version.md) | `v0.44.0` (const, источник истины — `version.go`) |

## Общие пакеты

| Пакет | Файл | Описание |
|-------|------|----------|
| progress-bar | [internal/pkg/progress-bar/progress-bar.md](internal/pkg/progress-bar/progress-bar.md) | Общий ANSI-прогресс-бар + `ProgressTracker` с окном xo (используют goxogen и gobp) |

## Смежные приложения

| Приложение | Файл | Описание |
|------------|------|----------|
| gobp | [gobp-index.md](gobp-index.md) | Сборка Go-приложений с прогресс-баром (deployment `build-app-progress`) |
| xouid | [xouid-index.md](xouid-index.md) | SQL-to-Go генератор UPDATE/INSERT/DELETE; вызывается как субпроцесс на шаге 1 пайплайна (`queries/uid`) |