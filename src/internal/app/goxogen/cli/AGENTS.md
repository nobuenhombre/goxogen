# cli — CLI configuration for goxogen

## Назначение

Парсинг CLI-аргументов для goxogen через `suikat/pkg/clivar`.

## Файлы

| Файл | Назначение |
|------|------------|
| `cli.go` | `Config` struct с тегами `cli`, геттер отсутствует (type assertion в provider) |
| `provider.go` | `ProviderSet` + `ProvideCLI()` |

## Ключевые типы

- **Service** — пустой интерфейс (маркер для Wire)
- **Config** — `{RunType, Config, LogFile, Version}` с тегами `cli`

## Wire-интеграция

```
ProvideCLI() → (Service, func(), error)
```

## Правила изменения

- При добавлении флага — обновить структуру `Config` и корневой `AGENTS.md`
- `cli.Service` — пустой интерфейс, type assertion в `ProvideConfigApp`: `cliConfig.(*cli.Config).Config`
