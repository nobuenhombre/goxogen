# version — версия gobp

## Назначение

Единственный источник версии приложения gobp. Используется в `cmd/gobp/main.go` для флага `-version`/`--version` (вывод версии до инициализации Wire).

## Состав пакета (1 файл)

| Файл | Назначение |
|------|------------|
| `version.go` | Константа `Version = "v0.9.0"` |
| `AGENTS.md` | Контекст агента для этого каталога |

## Публичный контракт

```go
// Version is the current application version following SemVer format.
const Version = "v0.9.0"
```

## Проводка в приложении

```
cmd/gobp/main.go
 └─ перебор os.Args[1:]: arg == "-version" || arg == "--version"
      → fmt.Println(version.Version); os.Exit(0)
```

Wire-интеграции нет: пакет импортируется напрямую в `main.go` до вызова `initializeApp()`.

## Факты о коде и примечания

- **Текущее значение в коде — `v0.9.0`**. Вложенный и корневой `AGENTS.md`, а также README указывают `v0.7.0` — устаревшие данные; источником истины является `version.go`.
- Формат SemVer: `vMAJOR.MINOR.PATCH`.
- Версия — `const`, не переменная: не переопределяется через ldflags `-X`; бампится правкой исходника.

## Кросс-ссылки

- [Точка входа gobp](../../../../cmd/gobp/gobp.md)
- [CLI-конфигурация gobp (флаг -version)](../cli/cli.md)