# goxogen — Scaffolder для генерации кода

[![Go Version](https://img.shields.io/badge/Go-1.26.1-blue)](https://go.dev/dl/)
[![Wire DI](https://img.shields.io/badge/DI-Google_Wire-green)](https://github.com/google/wire)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

**goxogen** — это монорепозиторий с тремя CLI-приложениями для генерации Go-кода, визуализации процесса сборки и генерации PostgreSQL-запросов.

## Приложения

| Приложение | Версия | Описание |
|------------|--------|----------|
| **goxogen** | v0.1.0 | Генератор кода из YAML-шаблонов с полным пайплайном XO-генерации |
| **gobp** | v0.6.0 | Сборка Go с прогресс-баром — подсчёт шагов, ETA, агрегация ошибок |
| **xouid** | v0.1.0 | Генератор типизированных Go-функций из SQL-запросов (UPDATE/INSERT/DELETE) с EXPLAIN-валидацией через PostgreSQL |

---

## Архитектура

```
goxogen/
├── AGENTS.md                 # Контекст для dev-агента (30+ вложенных AGENTS.md)
├── Makefile                  # Корневые цели сборки (deps, wire)
├── go.mod                    # Go-модуль (Go 1.26.1)
├── src/
│   ├── cmd/
│   │   ├── goxogen/          # CLI scaffolder (runtype: init, xo)
│   │   ├── gobp/             # CLI сборки с прогресс-баром
│   │   └── xouid/            # CLI генератора XOID-запросов
│   └── internal/
│       └── app/
│           ├── goxogen/      # Config, CLI, domain, log, version
│           ├── gobp/         # CLI, domain, version
│           └── xouid/        # CLI, domain, postgres, version
├── service/
│   └── deployments/          # Makefile'ы сборки и установки под Linux
├── bin/                      # Скомпилированные бинарники (gitkeep'd)
├── goxogen                   # Скомпилированный goxogen
├── gobp                      # Скомпилированный gobp
├── xouid                     # Скомпилированный xouid
├── README.md                 # Документация на английском
├── README.RU.md              # Документация на русском
└── LICENSE                   # Apache 2.0
```

Каждый пакет содержит собственный AGENTS.md с описанием назначения, ключевых типов, Wire-интеграции и правил модификации.

## Технологический стек

| Компонент | Технология | Версия |
|-----------|-----------|--------|
| Язык | Go | 1.26.1 |
| DI-фреймворк | Google Wire | v0.7.0 |
| Парсинг CLI | nobuenhombre/suikat/pkg/clivar | v0.0.170 |
| PostgreSQL | jackc/pgx/v5 | v5.9.2 |
| YAML | gopkg.in/yaml.v3 | v3.0.1 |
| UUID | github.com/google/uuid | v1.6.0 |
| Обёртка ошибок | suikat/pkg/ge (ge.Pin) | — |
| Файловый ввод/вывод | suikat/pkg/fico | — |
| Файловые утилиты | suikat/pkg/futi | — |
| Абстракция БД | suikat/pkg/db/connectors/postgres-pgx-db | — |

## Возможности

- **Scaffolding кода** — генерация структур Go-проектов из YAML-конфигов через `goxogen -runtype=init`
- **Пайплайн XO-генерации** — 7-шаговый процесс: схема БД → модели → функции запросов (one/many/UID) → извлечение репозиториев → форматирование
- **Прогресс-бар сборки** — `gobp` оборачивает `go build` с подсчётом шагов через dry-run, ANSI-прогресс-баром, отображением времени/ETA и сбором ошибок
- **Генерация PostgreSQL-запросов** — `xouid` валидирует SQL через EXPLAIN, парсит `%%param type%%` дескрипторы и генерирует типизированные Go-функции через шаблоны
- **Google Wire DI** — чистое внедрение зависимостей во всех трёх приложениях с правильным порядком cleanup

## Начало работы

### Требования

- Go 1.26.1+
- Google Wire CLI (`go install github.com/google/wire/cmd/wire@latest`)

### Сборка

```bash
make deps          # Переинициализация go.mod, загрузка зависимостей
make wire          # Генерация wire_gen.go для всех приложений

# Сборка отдельных приложений
go build ./src/cmd/goxogen
go build ./src/cmd/gobp
go build ./src/cmd/xouid
```

### Тестирование

```bash
go test ./... -v
```

## Флаги CLI

### goxogen

| Флаг | По умолчанию | Описание |
|------|-------------|----------|
| `-config` | `config.yaml` | Путь к YAML-конфигу |
| `-log` | `""` | Путь к лог-файлу |
| `-version` | `false` | Показать версию и выйти |

### gobp

| Флаг | По умолчанию | Описание |
|------|-------------|----------|
| `-binary` | `.` | Go-бинарник для сборки (напр. `./src/cmd/myapp`) |
| `-out` | `./build/app` | Путь к выходному бинарнику |
| `-verbose` | `false` | Показать полный вывод сборки |
| `-full-rebuild` | `false` | Полная пересборка (`go build -a`) |
| `-version` | `false` | Показать версию и выйти |

### xouid

| Флаг | По умолчанию | Описание |
|------|-------------|----------|
| `-out` | `""` | Путь к выходному файлу |
| `-dsn` | `""` | PostgreSQL DSN |
| `-template-path` | `""` | Путь к директории с шаблонами |
| `-package` | `""` | Имя пакета в генерируемом Go-коде |
| `-schema` | `public` | Имя схемы |
| `-query-type` | `""` | Тип запроса (генерируется как `{type}.xouid.go`) |
| `-query-func` | `""` | Имя генерируемой Go-функции |
| `-query` | `""` | Путь к файлу с SQL-запросом |
| `-verbose` | `false` | Не показывать приветствие |
| `-version` | `false` | Показать версию и выйти |

## Пайплайн XO-генерации кода

При запуске goxogen с `-runtype=xo -config=config.yaml` выполняется 7-шаговый пайплайн генерации Go-кода из схемы PostgreSQL:

### Структура YAML-конфига

```yaml
config:
  db:
    host: localhost
    port: 5432
    name: dbname
    user: user
    pass: pass
    sslmode: disable
    pool_max_conns: 10
    backups:
      path: /path/to/backups
  codegen:
    path: ./gen           # Директория вывода
    package: gen           # Имя Go-пакета
    queries: ./queries     # Директория с SQL-файлами
    ignore_fields: created_at,updated_at
```

### Шаги пайплайна

| Шаг | Действие | Описание |
|-----|----------|----------|
| 1 | **runXO** | Удаляет старые `.xo.go`/`.xouid.go`, запускает xo для моделей + запросов (one/many/uid), удаляет stored procedures |
| 2 | **replaceInterfaceToAny** | Заменяет `interface{}` на `any` во всех сгенерированных файлах |
| 3 | **glueXoXouid** | Склеивает `.xo.go` + `.xouid.go` → временные `.xo-xouid.go` |
| 4 | **extractRepo** | Извлекает блоки `@repo-start`/`@repo-end` → `*-repo.xo.go` (файлы репозиториев) |
| 5 | **removeXoXouid** | Удаляет временные `.xo-xouid.go` файлы |
| 6 | **cleanXoXouidSourceBlocks** | Удаляет маркеры `@repo-start`/`@repo-end` из `.xo.go` и `.xouid.go` |
| 7 | **goFormatCode** | Запускает `go fmt`, `goimports -w`, `go vet` для чистого вывода |

SQL-файлы именуются по шаблону `TypeName-FuncName.sql` и организованы в поддиректориях:
- `queries/one/` — запросы на одну строку
- `queries/many/` — запросы на множество строк
- `queries/uid/` — UPDATE/INSERT/DELETE (обрабатываются xouid)

## Внедрение зависимостей (Google Wire)

Каждое приложение использует Google Wire для внедрения зависимостей. Граф DI:

```
goxogen:  cli ProviderSet → domain ProviderSet → App
gobp:     cli ProviderSet → domain ProviderSet → App
xouid:    cli ProviderSet → postgres ProviderSet → domain ProviderSet → App
```

- Все `provider.go` экспортируют `var ProviderSet = wire.NewSet(ProvideXxx)`
- `wire.go` в `package main`: только `wire.Build()`, никакой логики
- `wire_gen.go` автосгенерирован — не редактировать вручную
- Все `Provide*` возвращают `(T, func(), error)` для правильного порядка cleanup

## Деплой

Makefile'ы деплоя для каждого приложения в `service/deployments/{app}/linux/`:

```bash
# Сборка + установка одного приложения
cd service/deployments/goxogen/linux && make all

# Сборка с прогресс-баром (требуется скомпилированный gobp)
cd service/deployments/goxogen/linux && make build-app-progress
```

Параметры сборки:
- `CGO_ENABLED=0`, `GOOS=linux`, `GOARCH=amd64`
- ldflags `-s -w` для уменьшения размера бинарника
- Устанавливается через симлинк в `/usr/local/bin/{app}`

## Разработка

### Соглашения

- **Ошибки:** оборачивать через `ge.Pin(err)` из `suikat/pkg/ge`
- **Cleanup:** все `Provide*` возвращают `(T, func(), error)` — Wire вызывает cleanup в обратном порядке
- **Алиасы:** конфликтующие имена пакетов алиасить (`domainapp`, `pgxdb`, `configapp`, `logfile`)
- **Версионирование:** только через флаг `-version` до инициализации Wire (избегает лишних подключений к БД/логам)
- **Линковка:** ldflags через переменную окружения `GOFLAGS`, не хардкодом
- **Вложенные AGENTS.md:** перед модификацией любого пакета ознакомьтесь с соответствующим AGENTS.md

### Известные проблемы

- `go vet файл.go` выдаёт `undefined: Service` в provider.go — всегда использовать `go vet ./...`
- Прогресс-бар gobp (`\r`) конфликтует с выводом ошибок компилятора — печатать `\n` перед строками ошибок
- Буфер сканера gobp: обязателен размер не менее 1MB (`scanner.Buffer(make([]byte, 64*1024), 1024*1024)`) для длинных строк CGO
- Dry-run gobp (`go build -n`) может насчитать меньше шагов, чем реальная сборка (`go build -x`) из-за диагностики CGO — прогресс-бар ограничен 100%
- Шаблоны xouid: ожидаются файлы `xouid_package.go.tpl` (имя шаблона: `xouidpackage`) и `xouid_query.go.tpl` (имя шаблона: `xouidquery`) в `-template-path`
- Параметры SQL в xouid: формат `%%имяПараметра тип%%` (поддерживаемые типы: int, int32, int64, float, float32, float64, string, bool, uuid.UUID, time.Time, массивы)
- `bin/.gitkeep`: не удалять директории бинарников (на них ссылаются Makefile'ы деплоя)

## Лицензия

Apache 2.0 — см. [LICENSE](LICENSE)