# domain — бизнес-логика goxogen (XO-пайплайн)

## Назначение

Пакет `domainapp` — оркестратор бизнес-логики goxogen: полный пайплайн генерации Go-кода из PostgreSQL-схемы (модели), SQL-запросов (one/many) и XOID-запросов (uid). Запускается как единственный режим приложения (`DomainService.Run()`), сопровождается прогресс-баром с окном вывода xo-субпроцессов.

## Состав пакета (14 файлов Go + 16 шаблонов)

| Файл | Назначение |
|------|------------|
| `domain-app.go` | Интерфейс `DomainService`, структура `AppDomain`, конструктор `New()`, `countPipelineSteps()` |
| `steps.go` | `Run()` — последовательность шагов пайплайна; `deleteGlob()` |
| `step-1-run-xo.go` | `runXO()` + запуск xo/xouid субпроцессов и стриминг их вывода |
| `step-1b-remove-crud-blocks.go` | `removeCRUDBlocks()` — вырезание блоков `// @crud` … `// @end-crud` (readonly-режим) |
| `step-2-replace-interface-to-any.go` | `replaceInterfaceToAny()` — замена `interface{}` → `any`, `Timestamp0WithoutTimeZone` → `sql.NullTime` |
| `step-3-glue-xo-xouid.go` | `glueXoXouid()` — слияние `.xo.go` + `.xouid.go` → `.xo-xouid.go` |
| `step-4-extract-repo.go` | `extractRepo()`/`extractRepoFile()` — извлечение `@repo-start`/`@repo-end` блоков в `*-repo.xo.go` |
| `step-5-remove-xo-xouid.go` | `removeXoXouid()` — удаление временных `.xo-xouid.go` |
| `step-6-clean-xo-xouid-source-blocks.go` | `cleanXoXouidSourceBlocks()` — удаление repo-маркеров из `.xo.go`/`.xouid.go` |
| `step-7-dedup-functions.go` | `dedupFunctions()` — дедупликация Go-функций (дубли из одинаковых индексов) |
| `step-7-dedup-functions_test.go` | Юнит-тест дедупликации (`removeDuplicateFuncs`) |
| `step-8-generate-db-repo.go` | `generateDbRepo()` — генерация агрегатного `a-db-repo.go` через `a-db-repo.go.tpl` |
| `step-9-generate-provider.go` | `generateProvider()` — генерация Wire `provider.go` через `provider.go.tpl` |
| `step-10-format-code.go` | `goFormatCode()` — чистка CRUD-маркеров, `go fmt`, `goimports`, `go vet` |
| `xo-config.go` | `XOConfig` (YAML-структура), `LoadXOConfig()`, connection strings, `resolveDbName()` |
| `templates.go` | Embedded FS шаблонов (`//go:embed templates/*.tpl`), `TemplatesDir()` |
| `templates/` | 16 `.tpl` файлов (см. таблицу ниже) |
| `AGENTS.md` | Контекст агента для этого каталога |

## Публичный контракт

```go
type DomainService interface {
	Run() error
}

type AppDomain struct {
	Cli *cli.Config
}

func New(cliConfig cli.Service) DomainService   // type assertion → *cli.Config
func (d *AppDomain) Run() error
func (d *AppDomain) countPipelineSteps(cfg *XOConfig) (int, error)
func LoadXOConfig(path string) (*XOConfig, error)   // xo-config.go
func TemplatesDir() (string, error)                 // templates.go
```

## YAML-конфигурация пайплайна (XOConfig)

Структура (фактическая, из `xo-config.go`):

| YAML-путь | Поле | Примечание |
|-----------|------|------------|
| `config.db.host` | string | |
| `config.db.port` | int | |
| `config.db.name` | string | источник `Db{Name}` при отсутствии `db_name` |
| `config.db.user` | string | |
| `config.db.pass` | string | |
| `config.db.sslmode` | string | |
| `config.db.pool_max_conns` | int | только для DSN xouid |
| `config.db.backups.path` | string, `omitempty` | указатель на подструктуру |
| `config.codegen.path` | string | выходной каталог (`-o`) |
| `config.codegen.package` | string | имя go-пакета |
| `config.codegen.queries` | string | каталог SQL-запросов |
| `config.codegen.ignore_fields` | string, `omitempty` | `--ignore-fields` |
| `config.codegen.db_name` | string, `omitempty` | переопределяет `db.name` для `Db{Name}Repo` |
| `config.codegen.db_is_readonly` | bool, `omitempty` | включает шаг removeCRUDBlocks |

Connection strings:
- **xo** (`XoConnectionString`): `pgsql://user:pass@host:port/name?sslmode=...` (схема `pgsql://`, пароль в URL).
- **xouid** (`XouidConnectionString`): `postgres://user:pass@host:port/name?sslmode=...` + `&pool_max_conns=N` если > 0.

`resolveDbName()`: `codegen.db_name` (если задан), иначе `db.name`; оба приводятся к PascalCase (разделители `_`, `-`, `.`, пробел; первый символ каждого слова в верхний регистр). `itoa` — самописное int→string (без strconv).

## Пайплайн: Run() — 10 шагов (+1 readonly)

Нумерация в коде (`steps.go`) соответствует `countPipelineSteps()` (10 базовых + 1 при `DbIsReadonly`):

| # (код) | Метод | Описание |
|---------|-------|----------|
| 1 | `runXO` | Удаляет `*.xo.go`, `*.xouid.go`; `xo basic`; `xo queries one`; `xo queries many`; `xouid queries uid`; удаляет `sp_*.xo.go` |
| 1b | `removeCRUDBlocks` | Только при `db_is_readonly: true`: вырезает строки между `// @crud` и `// @end-crud` во всех `*.xo.go` |
| 2 | `replaceInterfaceToAny` | `interface{}` → `any`; `Timestamp0WithoutTimeZone` → `sql.NullTime` (строковые ReplaceAll во всех `*.go`) |
| 3 | `glueXoXouid` | Для каждого `X.xo.go` создаёт `X.xo-xouid.go` = содержимое `.xo.go` + `\n` + содержимое `X.xouid.go` (если существует) |
| 4 | `extractRepo` | Из блоков `// @repo-start`/`// @repo-end` в `.xo-xouid.go` генерирует `*-repo.xo.go`: интерфейс `IXxxRepository` (методы дедуплицируются) + реализации (ресивер нормализуется к имени репозитория) |
| 5 | `removeXoXouid` | Удаляет все `.xo-xouid.go` |
| 6 | `cleanXoXouidSourceBlocks` | Удаляет блоки `@repo-start`/`@repo-end` из `.xo.go` и `.xouid.go` (кроме `*-repo.xo.go`) |
| 6b | `dedupFunctions` | Двухпроходный разбор: удаляет дубли функций/методов (ключ: имя для top-level, `*Receiver.Method` для методов), `seen` — глобальный по файлам пакета; дедупликация также внутри `*-repo.xo.go` |
| 7 | `generateDbRepo` | Сканирует `*-repo.xo.go` → имена `IXxxRepository` → генерирует `a-db-repo.go` (шаблон `a-db-repo.go.tpl`): структура `Db{DbName}Repo` + `NewDb{DbName}Repository(config *pgxdb.Config, log types.SQLLoggerFunc)` + `Close()`. Если repo-файлов/имён нет — пропуск |
| 8 | `generateProvider` | Если существует `a-db-repo.go` → генерирует `provider.go` (шаблон `provider.go.tpl`): `ProviderSet` + `Provider{DbName}(config *pgxdb.Config, log types.SQLLoggerFunc)` с cleanup (`dbRepo.Close()`) |
| 9 | `goFormatCode` | Удаляет строки с `@crud`/`@end-crud`; затем `go fmt <dir>`, `goimports -w` (все `*.go`, при сбое — только warning), `go vet` (при сбое — ошибка) |

Каждый шаг обёрнут: `pt.Increment("<название>")` → при ошибке `pt.AddError` + `pt.Fail()` + возврат `ge.Pin(fmt.Errorf(...))`.

## Детали шага 1 (runXO)

- **xo basic**: `xo <dsn> -o <out> --template-path <templates> --package <pkg> -v [--ignore-fields ...]`. stdout+stderr стримятся построчно в «окно xo» под прогресс-баром через фильтр `isXoObjectName` (отбрасывает SQL-фрагменты, заголовки `SQL:`/`PARAMS:`/`[]`, строки `[public]`, `[public r]`; снимает суффикс ` true]`/` false]`).
- **xo queries one/many**: для каждого `queries/<subdir>/*.sql` (сорт. по имени, `--query-mode --query-trim --query-strip --query-interpolate --query-type <Type> --query-func <Func> -v`), имя файла вида `TypeName[-FuncName].sql`; если целевой `<type>.xo.go` уже существует — добавляется `--append`. SQL подаётся через stdin. `one` дополнительно получает `--query-only-one`.
- **xouid uid**: для каждого `queries/uid/*.sql` вызывается бинарник `/usr/local/bin/xouid` с флагами `-out`, `-dsn`, `-template-path`, `-package`, `-query-type`, `-query-func`, `-query`, `-verbose=false`; вывод через `CombinedOutput()` (не стримится).
- **Стриминг**: `runXoStreamCmd` объединяет stdout+stderr через `io.Pipe` + 2 goroutine, читает построчно (`scanner.Buffer(64KB, 1MB)`), строки «имён объектов» уходят в `pt.PushXoLine` (кольцевой буфер 5 строк). После каждого блока (basic / подкаталог) — `pt.ClearXoOutput()`.

## Прогресс-бар (совместно с progress-bar)

- Заголовок пайплайна: `PipelineTitle = "Generating code"`, project name фиксирован — `"xoxgen"`.
- Pre-count: `countPipelineSteps()` — 10 (+1 readonly). Число `pt.Increment()` в Run() совпадает с ним.
- Перед пайплайном печатаются Connection string / Output / Package / Queries цветом `ColorProject`.
- На время пайплайна `log.SetOutput(io.Discard)` (восстановление через defer) — чтобы `log.Printf` не сдвигал прогресс-бар; логи сохраняются только при `-log` (см. пакет log).
- `pt.Finish()` + `os.Stdout.Sync()` перед возвратом.

## Проводка в приложении

```
cmd/goxogen/main.go
 └─ initializeApp() (Wire)
      └─ domainapp.ProvideDomain(cli.Service) → (DomainService, cleanup, nil)
           └─ New(cliConfig) → &AppDomain{Cli: cliConfig.(*cli.Config)}
                → app.Run() → dom.Run()
                     ├─ LoadXOConfig(d.Cli.Config)  // файл из -config
                     ├─ countPipelineSteps(cfg)
                     ├─ NewProgressTracker("Generating code", total)
                     ├─ TemplatesDir() → temp dir с 16 .tpl
                     └─ 10 (или 11) шагов пайплайна
```

## Факты о коде и примечания

- **Нумерация файлов versus шагов сдвинута**: `step-8-generate-db-repo.go` реализует шаг 7, `step-9-generate-provider.go` — шаг 8, `step-10-format-code.go` — шаг 9. Файл `step-7-dedup-functions.go` — это шаг 6b.
- **Несостыковка шаблонов**: комментарий AGENTS.md обещает `pgxdb.NewDB` для конструктора DbRepo, но шаблон `a-db-repo.go.tpl` вызывает `pgxdb.New(config, log)`. Также в шаблоне лишняя строка `"github.com/nobuenhombre/suikat/pkg/db/types"` импортируется и используется (`types.SQLLoggerFunc`).
- `xo basic` — единственный вызов без `--append`: модели генерируются начисто; queries дописываются в уже существующие файлы (`--append` по `os.Stat`), что создаёт риск дублей — их чинит шаг 6b `dedupFunctions`.
- Пароль БД попадает в URL connection string и логируется (`log.Printf("[xo] Running: ...")`).
- Ошибки `.xo-go`-субпроцессов содержат полный захваченный вывод команды (`\n%s`).
- Ошибки xo/xouid на шаге 1 оборачиваются, но пайплайн продолжает следующие вызовы только на успехе каждого предыдущего (sequential fail-fast).
- `extractRepo` при ошибке отдельного файла логирует warning и **продолжает** (skip), а не падает.
- `goimports` отсутствие в PATH не роняет шаг 9 — только «Warning: goimports failed (may not be installed)»; `go vet` при ошибке — фатален для шага.
- `dedupFunctions` удаляет дубли **глобально по всем файлам пакета**: если в любом файле функция с ключом уже встречалась, во всех последующих файлах она вырезается — корректно для Go-пакета, но «первый встреченный» определяется порядком сортировки имён файлов.
- Тест `step-7-dedup-functions_test.go` покрывает `removeDuplicateFuncs`/`extractFuncKey`.
- Проект в прогресс-баре назван `xoxgen` (опечатка в коде `NewProgressTracker`), тогда как продукт — goxogen.

## Шаблоны (16 .tpl, embedded)

| Шаблон | Назначение |
|--------|------------|
| `postgres.type.go.tpl` | Генерация типов PostgreSQL (xo basic) |
| `postgres.enum.go.tpl` | Enum-генерация |
| `postgres.foreignkey.go.tpl` | Внешние ключи |
| `postgres.index.go.tpl` | Индексы |
| `postgres.proc.go.tpl` | Хранимые процедуры |
| `postgres.query.go.tpl` | Функции запросов |
| `postgres.querytype.go.tpl` | Типы запросов |
| `mssql.type.go.tpl` | MSSQL-типы |
| `mysql.type.go.tpl` | MySQL-типы |
| `oracle.type.go.tpl` | Oracle-типы |
| `xo_db.go.tpl` | Обёртка подключения xo |
| `xo_package.go.tpl` | Заголовок пакета xo |
| `xouid_package.go.tpl` | Заголовок пакета xouid |
| `xouid_query.go.tpl` | Функция запроса xouid |
| `a-db-repo.go.tpl` | Агрегатный DbRepo (шаг 7, `text/template`) |
| `provider.go.tpl` | Wire-провайдер (шаг 8, `text/template`) |

Все 16 извлекаются в `os.MkdirTemp("", "goxogen-templates-*")` при старте пайплайна (`TemplatesDir()`); каталог живёт до конца процесса. Шаблоны 7-8 (`a-db-repo.go.tpl`, `provider.go.tpl`) читаются напрямую из embedded FS через `text/template`.

## Кросс-ссылки

- [CLI-конфигурация goxogen](../cli/cli.md)
- [Конфиг YAML goxogen (каркас — XOConfig здесь, в domain)](../config/config.md)
- [Лог-файл goxogen](../log/log.md)
- [Версия goxogen](../version/version.md)
- [Пакет progress-bar (тракер/окно xo)](../../../pkg/progress-bar/progress-bar.md)
- [Точка входа goxogen](../../../../cmd/goxogen/goxogen.md)