# Карта документации: xouid

## Общее

| Раздел | Файл |
|--------|------|
| Бизнес-модель | [xouid-business-model.md](xouid-business-model.md) |
| Карта goxogen | [goxogen-index.md](goxogen-index.md) |
| Карта gobp | [gobp-index.md](gobp-index.md) |

## Приложение — точка входа

| Раздел | Файл | Описание |
|--------|------|----------|
| Точка входа (`src/cmd/xouid`) | [cmd/xouid/xouid.md](cmd/xouid/xouid.md) | main.go → panic recovery → `-version` → Wire (4 провайдера, включая пул PostgreSQL) → `DomainService.Run()` |
| CLI-конфигурация | [internal/app/xouid/cli/cli.md](internal/app/xouid/cli/cli.md) | 10 флагов: `-out`, `-dsn`, `-template-path`, `-package`, `-schema`, `-query-type`, `-query-func`, `-query`, `-verbose`, `-version` |
| PostgreSQL-пул | [internal/app/xouid/postgres/postgres.md](internal/app/xouid/postgres/postgres.md) | Пул pgx (pgxpool) для EXPLAIN-запросов; обёрнут в `pgxdb.DBQuery` |
| Domain — SQL-to-Go генерация | [internal/app/xouid/domain/domain.md](internal/app/xouid/domain/domain.md) | 6-шаговый пайплайн: проверка SQL → проверка шаблонов → парсинг `%%param type%%` → EXPLAIN → рендер функции → запись файла |
| Версия | [internal/app/xouid/version/version.md](internal/app/xouid/version/version.md) | `v0.4.0` (const, источник истины — `version.go`) |

## Смежные приложения

| Приложение | Файл | Описание |
|------------|------|----------|
| goxogen | [goxogen-index.md](goxogen-index.md) | Вызывает `/usr/local/bin/xouid` на шаге 1 пайплайна (`queries/uid`); поставляет шаблоны xouid (встроены) |
| gobp | [gobp-index.md](gobp-index.md) | Сборка Go-приложений с прогресс-баром (deployment `build-app-progress`) |