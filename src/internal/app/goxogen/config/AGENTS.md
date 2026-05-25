# config — YAML configuration for goxogen

## Назначение

Загрузка и сохранение YAML-конфигурации приложения через `gopkg.in/yaml.v3` + `suikat/pkg/fico`.

## Файлы

| Файл | Назначение |
|------|------------|
| `config-app.go` | `Service` interface + `Config` struct + `Load()`/`Save()`/`Get()` |
| `provider.go` | `ProviderSet` + `ProvideConfigApp(cli.Service)` |
| `config-app_test.go` | Тест Load → Save → Read |
| `config-app_test_load.yaml` | Фикстура для теста |
| `config-app_test_save.yaml` | Выходной файл теста (остаётся после теста) |

## Ключевые типы

- **Service** — `{Load(fileName), Save(fileName), Get() *Config}`
- **Config** — пустая структура (пока без полей)

## Wire-интеграция

```
ProvideConfigApp(cli.Service) → (configapp.Service, func(), error)
```
Принимает `cli.Service`, type-assert к `*cli.Config`, читает `Config.Config` (поле `-config`).

## Правила изменения

- При добавлении секций в Config — обновить тест `config-app_test.yaml`
- `Get()` возвращает сам Config (не pointer на копию) — мутации возможны
