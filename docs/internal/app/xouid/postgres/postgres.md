# postgres — PostgreSQL-пул xouid

## Назначение

Пакет `postgres` — создание пула соединений PostgreSQL (`pgx/v5/pgxpool`) для выполнения EXPLAIN-запросов во время генерации. Обёрнут в интерфейс `pgxdb.DBQuery` из `suikat/pkg/db/connectors/postgres-pgx-db`.

## Состав пакета (2 файла)

| Файл | Назначение |
|------|------------|
| `db.go` | `NewDB(dataSourceName)` — парсинг DSN, создание `pgxpool`, cleanup закрытия |
| `provider.go` | Wire `ProviderSet` + `ProvideDB(cli.Service)` |
| `AGENTS.md` | Контекст агента для этого каталога |

## Публичный контракт

```go
// NewDB создаёт пул подключений PostgreSQL.
func NewDB(dataSourceName string) (pgxdb.DBQuery, func(), error)
```

Реализация:
1. `pgxpool.ParseConfig(dataSourceName)` — ошибка → `ge.Pin(err)` (cleanup no-op);
2. `pgxpool.NewWithConfig(context.Background(), config)` — ошибка → `ge.Pin(err)`;
3. cleanup: `log.Println("[wire-cleanup] PostgreSQL DB closing")` + `connectPool.Close()`;
4. результат: `&pgxdb.Conn{Pool: connectPool}` как `pgxdb.DBQuery`.

## Проводка в приложении

```
main.main()
 └─ initializeApp() (Wire)
      └─ postgres.ProvideDB(cliConfig cli.Service)
           ├─ dsn := cliConfig.GetDSN()          // флаг -dsn
           └─ NewDB(dsn) → (pgxdb.DBQuery, cleanup, nil)
                → domainapp.ProvideDomain(cliConfig, dbQuery)
```

Wire-граф: `cli.ProviderSet → postgres.ProviderSet → domainapp.ProviderSet → newApp`. Cleanup закрывает пул последним (обратный порядок: App → Domain → **PostgreSQL pool** → CLI).

## Факты о коде и примечания

- Если `NewDB` падает на этапе ParseConfig — Wire получает no-op cleanup (`log.Println("[wire-cleanup] DB cleanup (no-op)")`), пул не создан.
- Сам пакет не выполняет запросы — только держит пул; запросы EXPLAIN идут через `pgxdb.DBQuery.Query(ctx, explainSQL)` (domain).
- Подключение реально открывается при первом запросе пула (лениво): `NewWithConfig` сам по себе соединение не устанавливает.
- `pgxdb.Conn` — структура из suikat-пакета с полем `Pool *pgxpool.Pool`; сюда передаётся как значение с указателем на пул.

## Кросс-ссылки

- [Domain-слой xouid (использование пула для EXPLAIN)](../domain/domain.md)
- [CLI-конфигурация xouid (флаг -dsn)](../cli/cli.md)
- [Точка входа xouid](../../../../cmd/xouid/xouid.md)