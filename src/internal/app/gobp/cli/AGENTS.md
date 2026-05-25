# cli — CLI configuration for gobp

## Назначение

Парсинг CLI-аргументов для gobp через `suikat/pkg/clivar`.

## Файлы

| Файл | Назначение |
|------|------------|
| `cli.go` | `Service` interface + `Config` struct + геттеры |
| `provider.go` | `ProviderSet` + `ProvideCLI()` |

## Ключевые типы

- **Service** — `{GetBinary(), GetOut(), GetVerbose(), GetFullRebuild()}`
- **Config** — `{Binary, Out, Verbose, FullRebuild}` с тегами `cli`

## Wire-интеграция

```
ProvideCLI() → (Service, func(), error)
```

## Правила изменения

- При добавлении флага — добавить поле + геттер + обновить `Service` interface
- Геттеры — простые return полей (Getter pattern для Wire-абстракции)
