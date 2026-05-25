# goxogen (cmd) — Entry point for scaffolder

## Назначение

Точка входа для goxogen — code generation scaffolder.

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

- Не добавлять Wire-зависимости, которых нет в provider.go пакетов
- `newApp()` возвращает `(IApp, func(), error)` — cleanup для закрытия ресурсов
