# gobp (cmd) — Entry point for build progress tool

## Назначение

Точка входа для gobp — утилита отображения прогресса сборки.

## Файлы

| Файл | Назначение |
|------|------------|
| `main.go` | panic recovery → `-version` check → `initializeApp()` → `app.Run()` |
| `app.go` | `IApp` + `App{dom: DomainService}` + `newApp()` |
| `wire.go` | `wire.Build(cli.ProviderSet, domainapp.ProviderSet, newApp)` |

## Wire-граф

```
cli.ProviderSet → domainapp.ProviderSet → newApp(IApp)
```

## Правила изменения

- Не добавлять логику в `wire.go`
- После изменения `providers` в cli/domain — `make wire`
