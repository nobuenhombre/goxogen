# version — версия goxogen

## Назначение

Единственный источник версии приложения goxogen. Используется в `cmd/goxogen/main.go` для флага `-version`/`--version` (вывод версии до инициализации Wire).

## Состав пакета (1 файл)

| Файл | Назначение |
|------|------------|
| `version.go` | Константа `Version = "v0.44.0"` |
| `AGENTS.md` | Контекст агента для этого каталога |

## Публичный контракт

```go
// Version is the current application version following SemVer format.
const Version = "v0.44.0"
```

## Проводка в приложении

```
cmd/goxogen/main.go
 └─ перебор os.Args[1:]: arg == "-version" || arg == "--version"
      → fmt.Println(version.Version); os.Exit(0)
```

Wire-интеграции нет: пакет импортируется напрямую в `main.go` до вызова `initializeApp()`.

## Факты о коде и примечания

- **Текущее значение в коде — `v0.44.0`**. Корневой и вложенный `AGENTS.md` и README указывают `v0.16.0` — устаревшие данные; источником истины является `version.go`.
- Формат SemVer: `vMAJOR.MINOR.PATCH`.
- Версия — `const`, не переменная: не может быть переопределена через ldflags `-X` (в отличие от схемы `var Version`). Бампить версию можно только правкой исходника.

## Кросс-ссылки

- [Точка входа goxogen](../../../../cmd/goxogen/goxogen.md)
- [CLI-конфигурация goxogen (флаг -version)](../cli/cli.md)