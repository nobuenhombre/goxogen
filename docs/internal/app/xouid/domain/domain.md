# domain — бизнес-логика xouid (SQL-to-Go генерация)

## Назначение

Пакет `domainapp` — оркестратор xouid: 6-шаговый пайплайн генерации Go-функции из одного SQL-файла (UPDATE/INSERT/DELETE) с валидацией запроса через `EXPLAIN` в реальном PostgreSQL и рендерингом из пользовательских шаблонов.

## Состав пакета (3 файла)

| Файл | Назначение |
|------|------------|
| `domain-xouid.go` | `DomainService` + `AppDomain`, `New()`, все шаги пайплайна |
| `types.go` | Константы SQL/шаблонов, `QueryParam`, `QueryTemplateData`, `PackageTemplateData`, error-типы |
| `provider.go` | Wire `ProviderSet` + `ProvideDomain(cli.Service, pgxdb.DBQuery)` |
| `AGENTS.md` | Контекст агента для этого каталога |

## Публичный контракт

```go
type DomainService interface {
	Run() error
}

type AppDomain struct {
	cliConfig cli.Service
	db        pgxdb.DBQuery
	queryText string
}

func New(cliConfig cli.Service, db pgxdb.DBQuery) (DomainService, error)
func (d *AppDomain) Run() error

// Шаги (все публичные методы AppDomain):
func (d *AppDomain) CheckQuery() error
func (d *AppDomain) CheckTemplatesExists() error
func (d *AppDomain) GetQueryParams() (*[]QueryParam, error)
func (d *AppDomain) CreateSQLQueryNormal(qp *[]QueryParam) string
func (d *AppDomain) CreateSQLQueryExplain(qp *[]QueryParam) (string, error)
func (d *AppDomain) CheckExplainSQLInPostgresql(qp *[]QueryParam) (*[]string, error)
func (d *AppDomain) CreateFuncQuery(qp *[]QueryParam) (string, error)
func (d *AppDomain) CreateNewPackage() (string, error)
func (d *AppDomain) WritePackageFile(queryStr string) error
```

## Пайплайн Run() — 6 шагов

| # | Шаг | Описание |
|---|-----|----------|
| 1 | `CheckQuery` | `strings.ToLower(queryText)`, `strings.Contains` хотя бы одного из `update`/`insert`/`delete`. Нет ни одного — `UnknownSQLConstructionError` |
| 2 | `CheckTemplatesExists` | В `-template-path` должны существовать файлы `xouid_package.go.tpl` и `xouid_query.go.tpl` (иначе `TemplateNotFoundError`) |
| 3 | `GetQueryParams` | Регэксп `(%%(\s|\S)*?%%)` находит все дескрипторы `%%name type%%`; каждый — ровно 2 слова через пробел после снятия `%%` (иначе `IncorrectQueryParamError`) |
| 4 | `CheckExplainSQLInPostgresql` | `EXPLAIN <sql>` с конкретными примерами значений вместо параметров → `db.Query(ctx)` → план как `[]string`; при `-verbose` план логируется |
| 5 | `CreateFuncQuery` | `template.ParseFiles(xouid_query.go.tpl)` + `ExecuteTemplate("xouidquery", QueryTemplateData)` с SQL, где параметры заменены на `$1, $2, …` |
| 6 | `WritePackageFile` | Файл `<out><queryType>.xouid.go` (тип — lowercase): если не существует — сначала рендерится заголовок пакета (`xouidpackage`), затем дописывается строка функции (append) |

`New()` при создании читает SQL-файл `-query` целиком в `queryText`.

## Параметры запроса

Формат дескриптора: `%%name type%%` (2 части, разделённые пробелом).

| Поле QueryParam | Описание |
|-----------------|----------|
| `Name` | Имя параметра (go-идентификатор) |
| `Type` | Go-тип параметра |
| `GetDescriptor()` | Возвращает `%%name type%%` (исходный вид) |

Замена для выполнения в PostgreSQL (`CreateSQLQueryNormal`): каждый дескриптор → `$N` (1-based порядок появления).

EXPLAIN-примеры по типам (`CreateSQLQueryExplain`):

| Go-тип | Пример в EXPLAIN |
|--------|------------------|
| `int`, `int32`, `int64` | `1` |
| `float`, `float32`, `float64` | `1.25` |
| `string` | `'hello'` |
| `bool` | `true` |
| `uuid.UUID` | `'<новый uuid>'::uuid` (генерируется `uuid.New()`) |
| `time.Time` | `'2020-02-03 15:45:10'` (без приведения типа) |
| `[]int`, `[]int32`, `[]int64` | `'{1,2,3}'::int[]` |
| любой другой | `UnknownQueryParamTypeError` |

## Типы данных (types.go)

- `SQLUpdate = "update"`, `SQLInsert = "insert"`, `SQLDelete = "delete"` — whitelist `CheckQuery`.
- `TemplateNewPackage = "xouid_package.go.tpl"` (template name `xouidpackage`), `TemplateQuery = "xouid_query.go.tpl"` (template name `xouidquery`).
- `QueryTemplateData{Type, Name, QueryParams, SqlQuery}` — данные шаблона функции.
- `PackageTemplateData{Package, Schema}` — данные заголовка пакета.
- Error-типы: `UnknownSQLConstructionError`, `TemplateNotFoundError`, `IncorrectQueryParamError`, `UnknownQueryParamTypeError` (все — `fmt.Errorf`-стиль, оборачиваются `ge.Pin`).
- `QueryExplainResult{Plan []string}` — объявлен, но **нигде не используется** (результат EXPLAIN — `*[]string`).

## Проводка в приложении

```
cmd/xouid/main.go
 └─ initializeApp() (Wire)
      ├─ postgres.ProvideDB(cliConfig)          → pgxdb.DBQuery (пул)
      └─ domainapp.ProvideDomain(cliConfig, db) → (DomainService, cleanup, nil)
           └─ New(cliConfig, db)
                ├─ queryFile.Read()  // -query
                └─ &AppDomain{cliConfig, db, queryText}
                     → app.Run() → dom.Run() → 6-шаговый пайплайн → запись .xouid.go
```

## Факты о коде и примечания

- `CheckQuery` использует `strings.Contains`, а не «первые слова»: запрос пройдёт валидацию, если слова `update`/`insert`/`delete` встречаются **в любом месте** текста (например, в комментарии или в SELECT-подзапросе). «SELECT … FROM … WHERE deleted_at IS NULL» пройдёт как update-free — но и `SELECT ... WHERE status = 'update'` пройдёт ошибочно.
- `GetQueryParams`: split по одному пробелу без нормализации — `%%name  type%%` (двойной пробел) даст 3 части и ошибку.
- `time.Time` в EXPLAIN подставляется как строка `'2020-02-03 15:45:10'` **без** `::timestamp` — точность типа зависит от контекста выражения в SQL.
- `WritePackageFile` не перезаписывает существующий файл: заголовок пакета пишется только при создании, функция всегда дописывается (append).
- Имя выходного файла: `-out` + `strings.ToLower(-query-type)` + `.xouid.go` — `-out` используется как префикс (ожидается либо путь с разделителем, либо `./out/`).
- `Verbose` логирует план через глобальный логгер (`log.Printf`), а не в stdout.
- В отличие от goxogen, шаблоны НЕ встроены: ожидаются на диске в `-template-path`.

## Кросс-ссылки

- [CLI-конфигурация xouid](../cli/cli.md)
- [PostgreSQL-пул xouid](../postgres/postgres.md)
- [Версия xouid](../version/version.md)
- [Точка входа xouid](../../../../cmd/xouid/xouid.md)