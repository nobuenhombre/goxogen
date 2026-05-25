# domain — Business logic for xouid

## Назначение

6-шаговый pipeline: валидация SQL → проверка шаблонов → парсинг параметров → EXPLAIN → рендер → запись.

## Файлы

| Файл | Назначение |
|------|------------|
| `domain-xouid.go` | `DomainService` + `AppDomain` + полный pipeline |
| `types.go` | SQL/Template константы, `QueryParam`, `QueryTemplateData`, error-типы |
| `provider.go` | `ProviderSet` + `ProvideDomain(cli.Service, pgxdb.DBQuery)` |

## Ключевые типы

- **DomainService** — `{Run() error}`
- **QueryParam** — `{Name, Type}` + `GetDescriptor()` → `%%name type%%`
- **QueryTemplateData** — данные для `xouidquery` шаблона
- **PackageTemplateData** — данные для `xouidpackage` шаблона
- Error types: `UnknownSQLConstructionError`, `TemplateNotFoundError`, `IncorrectQueryParamError`, `UnknownQueryParamTypeError`

## SQL → Explain type mapping

| Go type | EXPLAIN example |
|---------|-----------------|
| int, int32, int64 | `1` |
| float, float32, float64 | `1.25` |
| string | `'hello'` |
| bool | `true` |
| uuid.UUID | `'<new>'::uuid` |
| time.Time | `'2020-02-03 15:45:10'` |
| []int, []int32, []int64 | `'{1,2,3}'::int[]` |

## Правила изменения

- `CheckQuery()` — разрешены только UPDATE/INSERT/DELETE
- `%%param type%%` — строго два слова через пробел
- При добавлении типа параметра — обновить switch в `CreateSQLQueryExplain`
- `WritePackageFile()` — append к `.xouid.go`, не overwrite
- Шаблоны ожидаются в `-template-path` как файлы `xouid_package.go.tpl` и `xouid_query.go.tpl`
