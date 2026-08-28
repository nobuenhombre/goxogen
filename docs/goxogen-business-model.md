# Бизнес-модель: goxogen — конвейер генерации Go-кода из PostgreSQL

## Зачем это всё

goxogen — конвейер генерации Go-кода (code generation scaffolder) для проектов, использующих связку **xo + xouid**. Ручная регенерация кода из схемы — многошаговый и хрупкий процесс: нужно удалить старые `.xo.go`, запустить `xo basic`, `xo queries one/many`, `xo queries uid`, склеить файлы, вырезать repo-блоки, убрать дубли функций, сгенерировать агрегатный репозиторий и Wire-провайдер, отформатировать. Любое из действий, сделанное вручную и не в том порядке, ломает сборку.

goxogen сводит весь процесс к **одной команде** `goxogen -config config.yaml` с прогресс-баром.

## Замысел (Vision)

Одна команда должна превращать **PostgreSQL-схему + SQL-запросы** в готовый к использованию Go-пакет: модели, функции запросов, интерфейсы репозиториев, агрегатный `Db{DbName}Repo`, Wire-провайдер — отформатированный (`go fmt`), с импортами (`goimports`) и проверенный (`go vet`).

Пайплайн — **единственный режим работы** приложения (флага `-runtype` в коде нет). Есть один вспомогательный режим — readonly-генерация при `db_is_readonly: true` (добавляет шаг 1b).

## Предметная область

### 1. Входы конвейера

| Вход | Источник | Примечание |
|------|----------|------------|
| Схема PostgreSQL | YAML-конфиг `config.db.*` | `pgsql://user:pass@host:port/name?sslmode=...` для xo |
| SQL-запросы | `config.codegen.queries` | Подкаталоги `one/`, `many/`, `uid/`; имена `TypeName[-FuncName].sql` |
| Параметры генерации | `config.codegen.*` | `path`, `package`, `ignore_fields`, `db_name`, `db_is_readonly` |
| Шаблоны | **Встроены в бинарник** (`//go:embed templates/*.tpl`) | 16 файлов; извлекаются в temp-каталог при старте |

### 2. Выходы конвейера (артефакты)

| Артефакт | Что это | Кто создаёт |
|----------|---------|-------------|
| `*.xo.go`, `*.xouid.go` | Модели + функции запросов | xo (basic/one/many) + xouid (uid) |
| `*-repo.xo.go` | Репозитории из `@repo-start`/`@repo-end` блоков: интерфейс `IXxxRepository` + реализации | шаг 4 (extractRepo) |
| `a-db-repo.go` | Агрегатный `Db{DbName}Repo` + `NewDb{DbName}Repository(config, log)` + `Close()` | шаг 7 (generateDbRepo) |
| `provider.go` | Wire `ProviderSet` + `Provider{DbName}` с cleanup | шаг 8 (generateProvider) |

### 3. Производственный процесс (пайплайн)

Нумерация — фактическая из кода (`steps.go` / `countPipelineSteps`): 10 базовых шагов + 1 при readonly.

| # | Метод | Описание |
|---|-------|----------|
| 1 | `runXO` | Удаляет старые `*.xo.go`/`*.xouid.go`; `xo basic` (модели); `xo queries one`; `xo queries many`; `xouid queries uid` (субпроцесс `/usr/local/bin/xouid`); удаляет `sp_*.xo.go` (процедуры) |
| 1b | `removeCRUDBlocks` | Только `db_is_readonly: true`: вырезает строки между `// @crud` и `// @end-crud` |
| 2 | `replaceInterfaceToAny` | `interface{}` → `any`; `Timestamp0WithoutTimeZone` → `sql.NullTime` |
| 3 | `glueXoXouid` | Склейка `.xo.go` + `.xouid.go` → временный `.xo-xouid.go` |
| 4 | `extractRepo` | Из `@repo-start`/`@repo-end` блоков генерирует `*-repo.xo.go` |
| 5 | `removeXoXouid` | Удаляет временные `.xo-xouid.go` |
| 6 | `cleanXoXouidSourceBlocks` | Убирает repo-маркеры из `.xo.go`/`.xouid.go` |
| 6b | `dedupFunctions` | Удаляет дубли Go-функций, сгенерированных из разных индексов на одни и те же колонки |
| 7 | `generateDbRepo` | Сканирует `*-repo.xo.go` → генерирует `a-db-repo.go` (шаблон `a-db-repo.go.tpl`) |
| 8 | `generateProvider` | Генерирует `provider.go` (шаблон `provider.go.tpl`) — только если есть `a-db-repo.go` |
| 9 | `goFormatCode` | Удаляет `@crud`-маркеры; `go fmt`; `goimports -w`; `go vet` |

Каждый шаг: `pt.Increment()` → при ошибке `pt.AddError()` + `pt.Fail()` + fail-fast (пайплайн останавливается).

### 4. Место в экосистеме

- **xouid** — субпроцесс шага 1: пайплайн вызывает `/usr/local/bin/xouid` (hardcoded путь) для каждого `queries/uid/*.sql` (`-out`, `-dsn`, `-template-path`, `-package`, `-query-type`, `-query-func`, `-query`). Шаблоны xouid — из встроенного набора goxogen.
- **gobp** — используется deployment-Makefile'ами (`make build-app-progress`): сборка самого goxogen идёт через `gobp --full-rebuild`.
- **progress-bar** — общий пакет: визуализация пайплайна (заголовок `Generating code`, имя проекта `xoxgen`, окно xo на 5 строк).

## Ключевые принципы

1. **Одна команда** — весь путь от схемы до готового Go-пакета, без ручных шагов.
2. **Самодостаточность** — все 16 шаблонов встроены в бинарник; внешних файлов шаблонов не требуется.
3. **Видимость процесса** — прогресс-бар + живое окно xo (кольцевой буфер 5 строк), ошибки — под баром.
4. **Чистый результат** — форматирование, импорты и `go vet` на выходе; CRUD-маркеры вырезаются.
5. **Fail-fast** — остановка на первой ошибке шага с явным сообщением.
6. **Готовность к DI** — пайплайн генерирует не только код, но и Wire-провайдер с cleanup.

## Целевая аудитория

| Роль | Что получает |
|------|-------------|
| Go-разработчик | Готовый пакет моделей/репозиториев/провайдера одной командой; readonly-режим для схем без CRUD |
| Deployment-процесс | Стабильная перегенерация кода; сборку самого бинарника — через gobp (прогресс) |

## Архитектурные драйверы

| Потребность | Решение |
|------------|---------|
| Многошаговость ручной генерации | Единый пайплайн из 10 шагов с прогресс-баром |
| Дубли функций из дублирующихся индексов | Шаг 6b `dedupFunctions` (глобальный `seen` по пакету) |
| Repo-блоки в сгенерированных файлах | Извлечение в `*-repo.xo.go` + чистка маркеров |
| Готовность к Wire в приложении-потребителе | `a-db-repo.go` + `provider.go` |
| Ошибки компиляции после генерации | `go fmt` + `goimports` + `go vet` на шаге 9 |
| Долгий вывод xo не должен ломать бар | Окно xo (5 строк) + фильтр `isXoObjectName` + `log.SetOutput(io.Discard)` |

## Факты о коде и примечания

- **Число шагов**: в коде — 10 базовых (+1 readonly), хотя README и корневой AGENTS.md говорят о «9-step» — устаревшие данные. Нумерация файлов сдвинута: `step-8-*` реализует шаг 7, `step-9-*` — шаг 8, `step-10-*` — шаг 9.
- **Hardcoded путь xouid**: `/usr/local/bin/xouid`. На машине без установленного xouid шаг `uid` падает (ошибка субпроцесса).
- **Опечатка**: в прогресс-баре имя проекта — `xoxgen` (в коде `NewProgressTracker`), продукт — goxogen.
- **Пароль в логе**: DSN с паролем пишется через `log.Printf` при запуске xo-субпроцессов.
- **Несостыковка шаблона**: AGENTS.md обещает `pgxdb.NewDB` в конструкторе DbRepo, фактический шаблон `a-db-repo.go.tpl` вызывает `pgxdb.New(config, log)`.
- Версии в коде: goxogen `v0.44.0` (README: v0.16.0 — устарело).

## Кросс-ссылки

- [Карта документации goxogen](goxogen-index.md)
- [XO-пайплайн (детали)](internal/app/goxogen/domain/domain.md)
- [Конфиг YAML (каркас)](internal/app/goxogen/config/config.md)
- [progress-bar](internal/pkg/progress-bar/progress-bar.md)
- [gobp — бизнес-модель](gobp-business-model.md)
- [xouid — бизнес-модель](xouid-business-model.md)