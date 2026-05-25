# log — Log file management for goxogen

## Назначение

Перенаправление вывода `log` в файл, если указан флаг `-log`.

## Файлы

| Файл | Назначение |
|------|------------|
| `log-file.go` | `ILogFile` interface + `LogFile{}` — Open/Close/Get |
| `provider.go` | `ProviderSet` + `ProvideLogFile(cli.Service)` |

## Ключевые типы

- **ILogFile** — `{Open(name), Close(), Get() *os.File}`
- **LogFile** — обёртка над `*os.File`
- `Open()` вызывает `log.SetOutput(file)` — глобальный stdlib логгер

## Wire-интеграция

```
ProvideLogFile(cli.Service) → (ILogFile, func(), error)
```
Cleanup меняется: если файл открыт — cleanup = `logFile.Close()`
