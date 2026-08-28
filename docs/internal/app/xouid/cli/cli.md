# cli — CLI-конфигурация xouid

## Назначение

Парсинг аргументов командной строки приложения xouid через `github.com/nobuenhombre/suikat/pkg/clivar`. Результат — `cli.Config` с 9 геттерами; через Wire интерфейс `cli.Service` потребляется обоими пакетами: `postgres` (DSN) и `domain` (все остальные настройки).

## Состав пакета (2 файла)

| Файл | Назначение |
|------|------------|
| `cli.go` | Интерфейс `Service` (9 геттеров), структура `Config` с тегами `cli`, конструктор `New()` |
| `provider.go` | Wire `ProviderSet` + `ProvideCLI()` c cleanup |
| `AGENTS.md` | Контекст агента для этого каталога |

## Публичный контракт

```go
type Service interface {
	GetOut() string
	GetDSN() string
	GetTemplatePath() string
	GetPackage() string
	GetSchema() string
	GetQueryType() string
	GetQueryFunc() string
	GetQuery() string
	GetVerbose() bool
}

type Config struct {
	Out          string `cli:"out[output path]:string="`
	DSN          string `cli:"dsn[PostgreSQL DSN]:string="`
	TemplatePath string `cli:"template-path[user supplied template path]:string="`
	Package      string `cli:"package[package name used in generated Go code]:string="`
	Schema       string `cli:"schema[schema name to generate Go types for]:string=public"`
	QueryType    string `cli:"query-type[query generated Go type filename.xo.go]:string="`
	QueryFunc    string `cli:"query-func[query generated Go func name]:string="`
	Query        string `cli:"query[query file to generate Go type and func from]:string="`
	Verbose      bool   `cli:"verbose[dont view hello message]:bool=false"`
	Version      bool   `cli:"version[Show version and exit]:bool=false"`
}

func New() (Service, error)
```

Флаги:

| Поле | Флаг | Тип | По умолчанию | Описание |
|------|------|-----|--------------|----------|
| `Out` | `-out` | string | `""` | Выходной путь (префикс имени файла `*.xouid.go`) |
| `DSN` | `-dsn` | string | `""` | PostgreSQL DSN |
| `TemplatePath` | `-template-path` | string | `""` | Каталог пользовательских шаблонов |
| `Package` | `-package` | string | `""` | Имя пакета в генерируемом коде |
| `Schema` | `-schema` | string | `public` | Схема для генерируемых типов |
| `QueryType` | `-query-type` | string | `""` | Тип запроса → имя файла `<type>.xouid.go` |
| `QueryFunc` | `-query-func` | string | `""` | Имя генерируемой Go-функции |
| `Query` | `-query` | string | `""` | Путь к SQL-файлу |
| `Verbose` | `-verbose` | bool | `false` | «Не показывать hello-сообщение» (фактически — лог EXPLAIN-плана) |
| `Version` | `-version` | bool | `false` | Показать версию и выйти |

## Проводка в приложении

```
main.main()
 └─ initializeApp() (Wire)
      ├─ cli.ProvideCLI() → (Service, cleanup, nil)
      │    ├─ cleanup: log.Println("[wire-cleanup] CLI config cleanup")
      │    └─ New() → clivar.Load(&Config{})
      ├─ postgres.ProvideDB(cliConfig)    → GetDSN()
      └─ domainapp.ProvideDomain(cliConfig, dbQuery)
           → GetOut(), GetTemplatePath(), GetPackage(), GetSchema(),
             GetQueryType(), GetQueryFunc(), GetQuery(), GetVerbose()
```

## Факты о коде и примечания

- Самый большой CLI-набор в репозитории: 10 флагов (9 геттеров в интерфейсе + `Version`, читаемый напрямую в `main.go`).
- Описание флага `-verbose` — «dont view hello message», но фактическое поведение (domain): при `GetVerbose()` true логируется `EXPLAIN plan: %v` через `log.Printf`. Никакого hello-сообщения в коде нет — расхождение описания и поведения.
- Все геттеры — тривиальные возвраты полей; интерфейс используется без type assertion (в отличие от goxogen).
- `GetSchema()` используется только в данных шаблона заголовка пакета (`PackageTemplateData.Schema`).

## Кросс-ссылки

- [Domain-слой xouid](../domain/domain.md)
- [PostgreSQL-пул xouid](../postgres/postgres.md)
- [Версия xouid](../version/version.md)
- [Точка входа xouid](../../../../cmd/xouid/xouid.md)