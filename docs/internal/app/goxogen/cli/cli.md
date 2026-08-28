# cli — CLI-конфигурация goxogen

## Назначение

Парсинг аргументов командной строки приложения goxogen через `github.com/nobuenhombre/suikat/pkg/clivar`. Результат — `cli.Config`, который через Wire передаётся остальным пакетам приложения (config, log, domain).

## Состав пакета (2 файла)

| Файл | Назначение |
|------|------------|
| `cli.go` | Пустой интерфейс `Service`, структура `Config` с тегами `cli`, конструктор `New()` |
| `provider.go` | Wire `ProviderSet` + `ProvideCLI()` c cleanup |
| `AGENTS.md` | Контекст агента для этого каталога |

## Публичный контракт

```go
// Service — маркерный интерфейс для Wire (реализация — *cli.Config).
type Service interface {
}

// Config — структура CLI-конфигурации (теги clivar).
type Config struct {
	Config  string `cli:"config[Path to YAML config]:string=config.yaml"`
	LogFile string `cli:"log[Path to log file]:string="`
	Version bool   `cli:"version[Show version and exit]:bool=false"`
}

// New загружает значения флагов из аргументов командной строки.
func New() (Service, error)
```

Флаги, декларируемые тегами `cli`:

| Поле | Флаг | Тип | Значение по умолчанию | Описание |
|------|------|-----|----------------------|----------|
| `Config` | `-config` | string | `config.yaml` | Путь к YAML-конфигу |
| `LogFile` | `-log` | string | `""` | Путь к лог-файлу |
| `Version` | `-version` | bool | `false` | Показать версию и выйти |

Загрузка выполняется `clivar.Load(cfg)`; ошибка оборачивается `ge.Pin(err)`.

## Проводка в приложении

```
main.main()
 └─ initializeApp() (Wire)
      └─ cli.ProvideCLI() → (Service, cleanup, nil)
           ├─ cleanup: log.Println("[wire-cleanup] CLI config cleanup")
           └─ New() → clivar.Load(&Config{})
                → возвращается как cli.Service (реализация — *cli.Config)
```

Потребители `cli.Service` выполняют type assertion `cliConfig.(*cli.Config).Config` — так `config.ProvideConfigApp` получает путь к YAML-файлу (`-config`), а `log.ProvideLogFile` — путь к лог-файлу (`-log`).

## Факты о коде и примечания

- Интерфейс `Service` пустой — он служит маркером для Wire; фактическая реализация `*cli.Config`, и в пакетах-потребителях выполняется жёсткий type assertion без проверки (`cliConfig.(*cli.Config)` — panic при неверном типе).
- Флага `-runtype` в коде НЕТ (корневой AGENTS.md упоминает его — устаревшая информация). Реальный набор флагов: `-config`, `-log`, `-version`. У приложения один режим работы — пайплайн кодогенерации.
- default `-config` = `config.yaml` задаётся в теге clivar, а не в коде.
- `ProvideCLI()` возвращает cleanup даже при ошибке `New()` — Wire вызовет его при каскадном откате.

## Кросс-ссылки

- [Конфиг YAML goxogen](../config/config.md)
- [Лог-файл goxogen](../log/log.md)
- [Domain-слой goxogen (пайплайн)](../domain/domain.md)
- [Версия goxogen](../version/version.md)
- [Точка входа goxogen](../../../../cmd/goxogen/goxogen.md)