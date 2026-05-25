# service — Инфраструктура развёртывания

## Назначение

Содержит Makefile'ы для сборки и установки каждого приложения на Linux.

## Структура

```
service/
└── deployments/
    ├── goxogen/linux/Makefile
    ├── gobp/linux/Makefile
    └── xouid/linux/Makefile
```

## Файлы

| Файл | Назначение |
|------|------------|
| `{app}/linux/Makefile` | Build + install + uninstall для одного приложения |

## Правила изменения

- Все три Makefile идентичны по структуре (отличаются только `APP_NAME`)
- При изменении одного — синхронизировать остальные
- CGO_ENABLED=0, GOOS=linux, GOARCH=amd64
- `build-app-progress` зависит от собранного бинарника `gobp` в корне проекта
