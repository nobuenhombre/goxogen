# internal — Внутренние пакеты

## Назначение

Пакеты, недоступные для импорта извне модуля (`internal`-ограничение Go). Содержат всю бизнес-логику трёх приложений.

## Структура

```
internal/
└── app/
    ├── goxogen/   # scaffolder: config + cli + domain + log + version
    ├── gobp/      # build tool: cli + domain + version
    └── xouid/     # SQL generator: cli + domain + postgres + version
```

## Принципы

- Каждое приложение изолировано в своей директории `app/{name}/`
- Все пакеты экспортируют `var ProviderSet = wire.NewSet(...)` из `provider.go`
- Нет перекрёстных импортов между приложениями
