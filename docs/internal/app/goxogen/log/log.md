# log — лог-файл goxogen

## Назначение

Пакет `logfile` — перенаправление вывода стандартного логгера Go (`log`) в файл, если при запуске указан флаг `-log`. Используется для захвата диагностики пайплайна (включая команды xo/xouid), которая во время кодогенерации отключается от стандартного вывода (см. domain).

## Состав пакета (2 файла)

| Файл | Назначение |
|------|------------|
| `log-file.go` | Интерфейс `ILogFile`, структура `LogFile`, методы `Open`/`Close`/`Get` |
| `provider.go` | Wire `ProviderSet` + `ProvideLogFile(cli.Service)` |
| `AGENTS.md` | Контекст агента для этого каталога |

## Публичный контракт

```go
type ILogFile interface {
	Open(name string)
	Close()
	Get() *os.File
}

type LogFile struct {
	file *os.File
}
```

## Реализация

- **Open(name)**: `os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)`; при ошибке — `log.Fatalf(" -[exit]- error OpenFile log file (%v): %v", name, err)`. Затем `log.SetOutput(lf.file)` — глобальный stdlib-логгер начинает писать в файл.
- **Close()**: закрывает `*os.File`, если он был открыт; при ошибке — `log.Fatalf(...)`.
- **Get()**: возвращает дескриптор файла.

## Проводка в приложении

```
main.main()
 └─ initializeApp() (Wire)
      └─ logfile.ProvideLogFile(cliConfig cli.Service)
           ├─ logFile := &LogFile{}
           ├─ if len(cliConfig.(*cli.Config).LogFile) != 0   // флаг -log
           │    ├─ logFile.Open(path)
           │    └─ cleanup = logFile.Close
           └─ (ILogFile, cleanup, nil)
```

Если `-log` пуст — cleanup остаётся заглушкой (`log.Println("Log File cleanup")`), файл не открывается, вывод логгера не перенаправляется.

## Факты о коде и примечания

- Файл открывается в режиме APPEND (не перезаписывается между запусками).
- `Open()` не возвращает ошибку: при неудаче — `log.Fatalf` (аварийный выход из процесса).
- Важно для пайплайна: во время `Run()` domain временно переключает `log.SetOutput(io.Discard)`, чтобы строки `log.Printf` не ломали прогресс-бар. Единственный способ сохранить эти логи — флаг `-log`: тогда `log.SetOutput` в пайплайне перенаправляет их в файл, а не в Discard.
- `Get()` — публичный метод, но в коде приложения нигде не используется (мёртвый на практике).

## Кросс-ссылки

- [CLI-конфигурация goxogen (флаг -log)](../cli/cli.md)
- [Domain-слой goxogen (переключение log.SetOutput в Run())](../domain/domain.md)
- [Точка входа goxogen](../../../../cmd/goxogen/goxogen.md)