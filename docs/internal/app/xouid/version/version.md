# version — версия xouid

## Назначение

Единственный источник версии приложения xouid. Используется в `cmd/xouid/main.go` для флага `-version`/`--version` (вывод версии до инициализации Wire).

## Состав пакета (1 файл)

| Файл | Назначение |
|------|------------|
| `version.go` | Константа `Version = "v0.4.0"` |
| `AGENTS.md` | Контекст агента для этого каталога |

## Публичный контракт

```go
// Version is the current application version following SemVer format.
const Version = "v0.4.0"
```

## Проводка в приложении

```
cmd/xouid/main.go
 └─ перебор os.Args[1:]: arg == "-version" || arg == "--version"
      → fmt.Println(version.Version); os.Exit(0)
```

Wire-интеграции нет: пакет импортируется напрямую в `main.go` до вызова `initializeApp()`.

## Факты о коде и примечания

- **Текущее значение в коде — `v0.4.0`**. Вложенный и корневой `AGENTS.md` и README указывают `v0.1.0` — устаревшие данные; источником истины является `version.go`.
- Формат SemVer: `vMAJOR.MINOR.PATCH`.
- Версия — `const`, не переменная: не переопределяется через ldflags `-X`; бампится правкой исходника.

## Кросс-ссылки

- [Точка входа xouid](../../../../cmd/xouid/xouid.md)
- [CLI-конфигурация xouid (флаг -version)](../cli/cli.md)