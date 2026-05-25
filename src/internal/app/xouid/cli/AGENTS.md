# cli — CLI configuration for xouid

## Назначение

Парсинг CLI-аргументов для xouid через `suikat/pkg/clivar`.

## Файлы

| Файл | Назначение |
|------|------------|
| `cli.go` | `Service` interface (9 геттеров) + `Config` struct + геттеры |
| `provider.go` | `ProviderSet` + `ProvideCLI()` |

## Ключевые типы

- **Service** — `{GetOut(), GetDSN(), GetTemplatePath(), GetPackage(), GetSchema(), GetQueryType(), GetQueryFunc(), GetQuery(), GetVerbose()}`
- **Config** — флаги: `-out`, `-dsn`, `-template-path`, `-package`, `-schema`, `-query-type`, `-query-func`, `-query`, `-verbose`, `-version`

## Wire-интеграция

```
ProvideCLI() → (Service, func(), error)
```
`Service` используется и `postgres.ProviderSet`, и `domainapp.ProviderSet`.

## Правила изменения

- При добавлении флага — добавить поле + геттер + обновить `Service` interface
- Все геттеры — простые return (no logic)
