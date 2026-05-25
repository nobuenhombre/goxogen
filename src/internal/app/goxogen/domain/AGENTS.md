# domain — Business logic for goxogen

## Назначение

Оркестратор бизнес-логики goxogen. Пока содержит только заглушку `Run()`.

## Файлы

| Файл | Назначение |
|------|------------|
| `domain-app.go` | `DomainService` interface + `AppDomain` + `New(cliConfig)` |
| `provider.go` | `ProviderSet` + `ProvideDomain(cli.Service)` |

## Ключевые типы

- **DomainService** — `{Run() error}`
- **AppDomain** — хранит `Cli *cli.Config`

## Wire-интеграция

```
ProvideDomain(cli.Service) → (DomainService, func(), error)
```

## Правила изменения

- `Run()` — точка для добавления новых команд (через `-runtype`)
- `New()` принимает `cli.Service`, делает type assertion к `*cli.Config`
