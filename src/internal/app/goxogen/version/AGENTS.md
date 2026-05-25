# version — goxogen version constant

## Назначение

Единственный источник версии для goxogen.

## Файлы

| Файл | Назначение |
|------|------------|
| `version.go` | `const Version = "v0.1.0"` |

## Правила изменения

- Обновлять при релизе
- Формат: SemVer (`vMAJOR.MINOR.PATCH`)
- Нет Wire-интеграции (читается напрямую в `main.go`)
