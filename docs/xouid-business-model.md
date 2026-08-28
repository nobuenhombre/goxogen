# Бизнес-модель: xouid — типизированная генерация UPDATE/INSERT/DELETE из SQL

## Зачем это всё

Точечные SQL-запросы UPDATE/INSERT/DELETE в Go-проектах пишутся руками: обёртка-функция, подстановка параметров, работа с БД. Это однообразно и чревато ошибками трёх сортов:

- **ошибки в SQL** — запрос падает только на реальном вызове, а не при написании;
- **ошибки в параметрах** — опечатка в имени, неверный тип, несоответствие порядка;
- **неунифицированный стиль** — каждый разработчик пишет обёртки по-своему.

xouid генерирует типизированную Go-функцию из одного SQL-файла: дескрипторы `%%name type%%` в SQL задают параметры, запрос **валидируется через EXPLAIN в реальном PostgreSQL** до генерации, а результат рендерится из пользовательских шаблонов.

## Замысел (Vision)

Одна команда: `xouid -query update.sql -dsn "..." -template-path ./templates -package gen -query-type TaskStatus -query-func UpdateTaskStatusSetStatus` — превращает:

```
UPDATE task SET status = %%status string%% WHERE id = %%id int%%
```

в типизированную Go-функцию в файле `taskstatus.xouid.go` — но только после того, как запрос успешно прошёл `EXPLAIN` в реальной БД.

## Предметная область

### 1. Входы

| Вход | Флаг | Примечание |
|------|------|------------|
| SQL-файл | `-query` | читается целиком в `New()` |
| Схема БД | `-dsn` | PostgreSQL DSN для EXPLAIN-валидации |
| Шаблоны | `-template-path` | **обязательны на диске**: `xouid_package.go.tpl` (`xouidpackage`) + `xouid_query.go.tpl` (`xouidquery`) |
| Имя пакета/схемы | `-package`, `-schema` | для заголовка пакета (schema — default `public`) |
| Имя типа/функции | `-query-type`, `-query-func` | тип → имя файла `<type>.xouid.go`, func → имя функции |
| Выход | `-out` | префикс пути выходного файла |

### 2. Производственный процесс (пайплайн, 6 шагов)

| # | Шаг | Описание |
|---|-----|----------|
| 1 | `CheckQuery` | lowercase + `strings.Contains` хотя бы одного из `update`/`insert`/`delete`; иначе `UnknownSQLConstructionError` |
| 2 | `CheckTemplatesExists` | в `-template-path` должны лежать оба `.tpl`; иначе `TemplateNotFoundError` |
| 3 | `GetQueryParams` | регэксп `(%%(\s|\S)*?%%)` → дескрипторы `%%name type%%` (ровно 2 слова); каждый → `QueryParam{Name, Type}` |
| 4 | `CheckExplainSQLInPostgresql` | `EXPLAIN <sql>` с примерами значений вместо параметров → `db.Query(ctx)` → план `[]string`; при `-verbose` — `log.Printf` плана |
| 5 | `CreateFuncQuery` | `template.ParseFiles(xouid_query.go.tpl)` + `ExecuteTemplate("xouidquery", …)`; параметры → `$1, $2, …` |
| 6 | `WritePackageFile` | файл `<out><querytype>.xouid.go`: при создании — заголовок пакета (`xouidpackage`), затем **append** строки функции |

### 3. Выходы

| Выход | Описание |
|-------|----------|
| `<type>.xouid.go` | Заголовок пакета (только при создании файла) + типизированная функция запроса (дописывается) |
| EXPLAIN-план | логируется при `-verbose` |

### 4. Место в экосистеме

- **goxogen-пайплайн** — основной потребитель: шаг 1 (`runXO`) вызывает `/usr/local/bin/xouid` для каждого `queries/uid/*.sql` с флагами `-out`, `-dsn`, `-template-path`, `-package`, `-query-type`, `-query-func`, `-query`. Шаблоны поставляет goxogen (встроенные `xouid_package.go.tpl`/`xouid_query.go.tpl` извлекаются в temp-каталог).
- **Стенд-алон режим** — из командной строки разработчиком для единичных запросов.
- Результат xouid попадает в общий конвейер goxogen: `.xouid.go` склеивается с `.xo.go` (шаг 3), из блоков извлекаются репозитории (шаг 4).

## Ключевые принципы

1. **Проверяй до генерации** — EXPLAIN-валидация в реальной БД до рендера функции.
2. **Типизированные параметры** — `%%name type%%`-дескрипторы с whitelist типов (int, int32, int64, float*, string, bool, uuid.UUID, time.Time, `[]int*`).
3. **Append-only запись** — существующий файл не перезаписывается: заголовок пишется один раз, функции дописываются.
4. **Пользовательские шаблоны** — в отличие от goxogen, шаблоны НЕ встроены: ожидаются на диске (`-template-path`).

## Целевая аудитория

| Роль | Что получает |
|------|-------------|
| Пайплайн goxogen | Генерацию `queries/uid` без ручного переключения инструментов |
| Go-разработчик | Типизированную обёртку под конкретный SQL + подтверждение, что запрос валиден |

## Архитектурные драйверы

| Потребность | Решение |
|------------|---------|
| Ошибки в SQL обнаруживаются поздно | EXPLAIN в реальном PostgreSQL до генерации |
| Опечатки в именах/типах параметров | `%%name type%%`-дескрипторы + `IncorrectQueryParamError`/`UnknownQueryParamTypeError` |
| Поддержка частых типов | EXPLAIN-примеры по типам: `int`→`1`, `string`→`'hello'`, `uuid.UUID`→генерация uuid, `[]int`→`'{1,2,3}'::int[]` |
| Многостраничный файл запросов | Append-only: заголовок при создании, функция всегда дописывается |
| Проверка плана глазами | `-verbose` логирует `EXPLAIN plan: %v` |

## Факты о коде и примечания

- **`strings.Contains` вместо «первого слова»**: запрос проходит валидацию, если `update`/`insert`/`delete` встречается в любом месте текста — например, в комментарии или литерале (`status = 'update'`) — ложное срабатывание.
- **Парсинг параметров**: split по одному пробелу — `%%name  type%%` (двойной пробел) даст ошибку (3 части).
- **`time.Time` в EXPLAIN** подставляется как `'2020-02-03 15:45:10'` **без** `::timestamp` — точность типа зависит от контекста SQL-выражения.
- **Имя выходного файла**: `-out` + lowercase `-query-type` + `.xouid.go`; `-out` используется как префикс (ожидается путь с разделителем).
- **`QueryExplainResult{Plan []string}`** объявлен в types.go, но нигде не используется (EXPLAIN возвращается как `*[]string`).
- **Описание флага `-verbose`** — «dont view hello message», фактически — логирование плана; hello-сообщения в коде нет.
- Версии: в коде `v0.4.0`, AGENTS.md/README указывают `v0.1.0` — устаревшие данные.

## Кросс-ссылки

- [Карта документации xouid](xouid-index.md)
- [Domain — SQL-to-Go генерация (детали)](internal/app/xouid/domain/domain.md)
- [PostgreSQL-пул xouid](internal/app/xouid/postgres/postgres.md)
- [CLI-конфигурация xouid](internal/app/xouid/cli/cli.md)
- [goxogen — бизнес-модель (вызов xouid на шаге 1)](goxogen-business-model.md)