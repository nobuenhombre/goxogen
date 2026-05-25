# src — Исходный код (все приложения)

## Назначение

Корневая директория для Go-кода трёх приложений: goxogen (scaffolder), gobp (build tool), xouid (SQL-to-Go генератор).

## Структура

```
src/
├── cmd/           # entry points (main package)
└── internal/      # внутренние пакеты (не импортируются извне)
    └── app/
        ├── goxogen/   # config + cli + domain + log + version
        ├── gobp/      # cli + domain + version
        └── xouid/     # cli + domain + postgres + version
```

## Принципы

- Каждое приложение в `src/internal/app/{app}/` имеет собственную `version/`, `cli/`, `domain/`
- Wire DI: каждый пакет экспортирует `ProviderSet` из `provider.go`
- `src/cmd/{app}/wire.go` — build-tag `wireinject`, только `wire.Build()`
- Никакого общего кода между приложениями (каждый живёт изолированно)
