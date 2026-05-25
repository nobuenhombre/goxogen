# cmd — Entry points приложений

## Назначение

Пакеты `package main` — точки входа для трёх приложений. Каждая содержит Wire-инъекцию и делегирует `Run()` в `domain.DomainService`.

## Структура

```
cmd/
├── goxogen/     # goxogen main — code generation scaffolder
├── gobp/        # gobp main — build progress display
└── xouid/       # xouid main — PostgreSQL XOID query generator
```

## Файлы (на каждое приложение)

| Файл | Назначение |
|------|------------|
| `main.go` | Entry point: panic recovery → версия → Wire init → Run |
| `app.go` | `IApp` interface + `App{}` + `newApp()` Wire provider |
| `wire.go` | `//go:build wireinject` — `wire.Build()` всех ProviderSet |
| `wire_gen.go` | **Автосгенерирован** через `make wire` — не редактировать |

## Wire-интеграция

- `wire.go` только импортирует ProviderSet'ы и вызывает `wire.Build()`
- `app.go` экспортирует `newApp()` как Wire provider
- `wire_gen.go` — результат `wire gen ./src/...`

## Правила изменения

- Не добавлять логику в `wire.go`
- После изменения provider.go любого пакета — запустить `make wire`
- `wire_gen.go` форматировать через `gofmt -w` (прописано в Makefile)
