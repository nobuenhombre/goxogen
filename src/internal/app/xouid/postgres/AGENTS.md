# postgres — PostgreSQL connection for xouid

## Назначение

Создание пула соединений PostgreSQL через `pgx/v5/pgxpool` для выполнения EXPLAIN запросов.

## Файлы

| Файл | Назначение |
|------|------------|
| `db.go` | `NewDB(dsn)` — парсинг DSN, создание `pgxpool`, cleanup закрытия |
| `provider.go` | `ProviderSet` + `ProvideDB(cli.Service)` |

## Ключевые типы

- **pgxdb.DBQuery** — из `suikat/pkg/db/connectors/postgres-pgx-db` (интерфейс с `Query()`)
- **pgxdb.Conn** — структура с полем `Pool *pgxpool.Pool`

## Wire-интеграция

```
ProvideDB(cli.Service) → (pgxdb.DBQuery, func(), error)
```
Читает `cliConfig.GetDSN()` для подключения.

## Правила изменения

- Cleanup закрывает `Pool` — Wire вызывает его в обратном порядке
- При ошибке подключения cleanup возвращает no-op
