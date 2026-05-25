# xouid (cmd) — Entry point for XOID SQL-to-Go generator

## Назначение

Точка входа для xouid — генератор Go-кода из SQL (PostgreSQL EXPLAIN + templates).

## Файлы

| Файл | Назначение |
|------|------------|
| `main.go` | panic recovery → `-version` check → `initializeApp()` → `app.Run()` |
| `app.go` | `IApp` + `App{dom: DomainService}` + `newApp()` |
| `wire.go` | `wire.Build(cli.ProviderSet, postgres.ProviderSet, domainapp.ProviderSet, newApp)` |

## Wire-граф

```
cli.ProviderSet
    → postgres.ProviderSet     (depends on cli.Service → GetDSN())
    → domainapp.ProviderSet    (depends on cli.Service + pgxdb.DBQuery)
    → newApp(IApp)
```

## Правила изменения

- `postgres.ProviderSet` должен быть перед `domainapp.ProviderSet` в `wire.Build()`
- Cleanup: PostgreSQL pool закрывается последним (обратный порядок)
