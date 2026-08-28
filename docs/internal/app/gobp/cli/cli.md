# cli — CLI-конфигурация gobp

## Назначение

Парсинг аргументов командной строки приложения gobp через `github.com/nobuenhombre/suikat/pkg/clivar`. Результат — `cli.Config` с геттерами; через Wire интерфейс `cli.Service` передаётся в domain.

## Состав пакета (2 файла)

| Файл | Назначение |
|------|------------|
| `cli.go` | Интерфейс `Service` (4 геттера), структура `Config` с тегами `cli`, конструктор `New()` |
| `provider.go` | Wire `ProviderSet` + `ProvideCLI()` c cleanup |
| `AGENTS.md` | Контекст агента для этого каталога |

## Публичный контракт

```go
type Service interface {
	GetBinary() string
	GetOut() string
	GetVerbose() bool
	GetFullRebuild() bool
}

type Config struct {
	Binary      string `cli:"binary[Go binary to build, e.g. ./src/cmd/myapp]:string=."`
	Out         string `cli:"out[output binary path]:string=./build/app"`
	Verbose     bool   `cli:"verbose[show full build output]:bool=false"`
	FullRebuild bool   `cli:"full-rebuild[force full rebuild with go build -a]:bool=false"`
	Version     bool   `cli:"version[Show version and exit]:bool=false"`
}

func New() (Service, error)
```

Флаги:

| Поле | Флаг | Тип | По умолчанию | Описание |
|------|------|-----|--------------|----------|
| `Binary` | `-binary` | string | `.` | Go-бинарник/пакет для сборки |
| `Out` | `-out` | string | `./build/app` | Путь выходного бинарника |
| `Verbose` | `-verbose` | bool | `false` | Показывать полный вывод сборки |
| `FullRebuild` | `-full-rebuild` | bool | `false` | Форсировать полную пересборку (`go build -a`) |
| `Version` | `-version` | bool | `false` | Показать версию и выйти |

В отличие от goxogen-клi, `Service` — содержательный интерфейс с геттерами (`GetBinary`, `GetOut`, `GetVerbose`, `GetFullRebuild`); каждый геттер просто возвращает соответствующее поле. Поле `Verbose` в текущем domain-коде не используется (см. факты).

## Проводка в приложении

```
main.main()
 └─ initializeApp() (Wire)
      └─ cli.ProvideCLI() → (Service, cleanup, nil)
           ├─ cleanup: log.Println("[wire-cleanup] CLI config cleanup (gobp)")
           └─ New() → clivar.Load(&Config{})
                → domain.ProvideDomain(cliConfig) // через интерфейс, без type assertion
```

Domain обращается к конфигурации только через методы интерфейса: `GetBinary()`, `GetOut()`, `GetFullRebuild()`.

## Факты о коде и примечания

- `Config` содержит поле `Version`, но интерфейс `Service` геттера для него **не** имеет — проверка версии выполняется в `main.go` напрямую (файл `version.go`), и в domain это поле не передаётся.
- Геттер `GetVerbose()` существует в интерфейсе, но не вызывается ни в `domain.go`, ни где-либо ещё — на практике флаг `-verbose` не влияет на поведение (мёртвое поле/метод на текущий момент).
- Интерфейс с геттерами (Getter pattern) — в отличие от goxogen, где `cli.Service` пустой и потребители делают type assertion к `*cli.Config`.

## Кросс-ссылки

- [Domain-слой gobp](../domain/domain.md)
- [Версия gobp](../version/version.md)
- [Точка входа gobp](../../../../cmd/gobp/gobp.md)