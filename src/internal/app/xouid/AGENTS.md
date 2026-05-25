# xouid — PostgreSQL XOID Query Generator

## Назначение

CLI-утилита для генерации Go-функций из SQL-запросов (UPDATE/INSERT/DELETE) с проверкой через `EXPLAIN` в PostgreSQL.

## Pipeline

1. **CheckQuery** — проверка, что запрос UPDATE/INSERT/DELETE
2. **CheckTemplatesExists** — проверка наличия шаблонов `xouid_package.go.tpl` и `xouid_query.go.tpl`
3. **GetQueryParams** — парсинг `%%param type%%` из SQL
4. **CheckExplainSQLInPostgresql** — EXPLAIN в реальном PostgreSQL
5. **CreateFuncQuery** — рендеринг шаблона запроса
6. **WritePackageFile** — запись `.xouid.go` файла

## Файлы

| Файл | Назначение |
|------|------------|
| `cli/cli.go` | CLI флаги: `-out`, `-dsn`, `-template-path`, `-package`, `-schema`, `-query-type`, `-query-func`, `-query`, `-verbose`, `-version` |
| `cli/provider.go` | `cli.ProviderSet` + `ProvideCLI()` |
| `domain/types.go` | SQL/Template константы, `QueryParam`, `QueryTemplateData`, error-типы |
| `domain/domain-xouid.go` | `DomainService{}` — полный pipeline в 6 шагов |
| `domain/provider.go` | `domainapp.ProviderSet` + `ProvideDomain(cli.Service, pgxdb.DBQuery)` |
| `postgres/db.go` | `NewDB(dsn)` — создание `pgxpool` |
| `postgres/provider.go` | `postgres.ProviderSet` + `ProvideDB(cli.Service)` |
| `version/version.go` | `const Version = "v0.1.0"` |

## Ключевые типы

- **cli.Config** — CLI-флаги с геттерами (`GetOut()`, `GetDSN()`, `GetTemplatePath()`, etc.)
- **cli.Service** — интерфейс на 9 геттеров
- **QueryParam** — `{Name, Type}` + `GetDescriptor()` (возвращает `%%name type%%`)
- **QueryTemplateData** — данные для шаблона функции
- **PackageTemplateData** — данные для шаблона заголовка пакета
- **pgxdb.DBQuery** — из `suikat/pkg/db/connectors/postgres-pgx-db`

## Wire-интеграция

```
cli.ProviderSet
    → postgres.ProviderSet     (depends on cli.Service → dsn)
    → domainapp.ProviderSet    (depends on cli.Service + pgxdb.DBQuery)
    → newApp
```

## Правила изменения

- `CheckQuery()` разрешает только UPDATE/INSERT/DELETE — при добавлении нового типа синхронизировать с `UnknownSQLConstructionError`
- `GetQueryParams()` ожидает строго `%%name type%%` — два слова через пробел
- При добавлении нового типа (в switch `CreateSQLQueryExplain`) — добавить пример в case
- `WritePackageFile()` добавляет к существующему файлу (не перезаписывает)
- Шаблоны: `xouid_package.go.tpl` (TemplateNameNewPackage) и `xouid_query.go.tpl` (TemplateNameQuery)
