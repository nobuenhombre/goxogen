# goxogen — Scaffolder для генерации кода

[![Go Version](https://img.shields.io/badge/Go-1.26.1-blue)](https://go.dev/dl/)
[![Wire DI](https://img.shields.io/badge/DI-Google_Wire-green)](https://github.com/google/wire)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**goxogen** — это монорепозиторий с тремя CLI-приложениями для генерации Go-кода, визуализации процесса сборки и генерации PostgreSQL-запросов.

## Приложения

| Приложение | Версия | Описание |
|------------|--------|----------|
| **goxogen** | v0.1.0 | Генератор кода из YAML-шаблонов |
| **gobp** | v0.4.0 | Сборка Go с прогресс-баром |
| **xouid** | v0.1.0 | Генератор типизированных Go-функций из SQL-запросов с проверкой через PostgreSQL |

---

## Архитектура

```
goxogen/
├── AGENTS.md                 # Контекст для dev-агента
├── Makefile                  # Корневые цели сборки
├── go.mod                    # Go-модуль (Go 1.26.1)
├── src/
│   ├── cmd/
│   │   ├── goxogen/          # CLI scaffolder
│   │   ├── gobp/             # CLI сборки с прогресс-баром
│   │   └── xouid/            # CLI генератора XOID-запросов
│   └── internal/
│       └── app/
│           ├── goxogen/      # config + cli + domain + log + version
│           ├── gobp/         # cli + domain + version
│           └── xouid/        # cli + domain + postgres + version
├── service/
│   └── deployments/          # Makefile'ы сборки и установки под Linux
├── bin/                      # Скомпилированные бинарники (gitkeep'd)
├── goxogen                   # Скомпилированный goxogen
├── gobp                      # Скомпилированный gobp
└── xouid                     # Скомпилированный xouid
```

## Технологический стек

| Компонент | Технология | Версия |
|-----------|-----------|--------|
| Язык | Go | 1.26.1 |
| DI-фреймворк | Google Wire | v0.7.0 |
| Парсинг CLI | nobuenhombre/suikat/pkg/clivar | v0.0.170 |
| PostgreSQL | jackc/pgx/v5 | v5.9.2 |
| YAML | gopkg.in/yaml.v3 | v3.0.1 |
| Обёртка ошибок | suikat/pkg/ge (ge.Pin) | — |
| Файловый ввод/вывод | suikat/pkg/fico | — |
| Файловые утилиты | suikat/pkg/futi | — |

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
| `-runtype` | `init` | Тип запуска |
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

## Внедрение зависимостей (Google Wire)

Каждое приложение использует Google Wire для внедрения зависимостей. Граф DI:

```
goxogen:  cli.ProviderSet → domain.ProviderSet → App
gobp:     cli.ProviderSet → domain.ProviderSet → App
xouid:    cli.ProviderSet → postgres.ProviderSet → domain.ProviderSet → App
```

- Все `provider.go` экспортируют `var ProviderSet = wire.NewSet(ProvideXxx)`
- `wire.go` в `package main`: только `wire.Build()`, никакой логики
- `wire_gen.go` автосгенерирован — не редактировать вручную

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
- Устанавливается в `/usr/local/bin/{app}`, логи в `/var/log/{app}/`

## Разработка

### Соглашения

- **Ошибки:** оборачивать через `ge.Pin(err)` из `suikat/pkg/ge`
- **Cleanup:** все `Provide*` возвращают `(T, func(), error)` — Wire вызывает cleanup в обратном порядке
- **Алиасы:** конфликтующие имена пакетов алиасить (`domainapp`, `pgxdb`, `configapp`, `logfile`)
- **Версионирование:** только через флаг `-version` до инициализации Wire (избегает лишних подключений к БД/логам)
- **Линковка:** ldflags через переменную окружения `GOFLAGS`, не хардкодом

### Известные проблемы

- `go vet файл.go` выдаёт `undefined: Service` в provider.go — всегда использовать `go vet ./...`
- Прогресс-бар gobp (`\r`) конфликтует с выводом ошибок компилятора — печатать `\n` перед строками ошибок
- Шаблоны xouid ожидают файлы `xouid_package.go.tpl` и `xouid_query.go.tpl` в пути `-template-path`
- `bin/.gitkeep` — не удалять директории бинарников (на них ссылаются Makefile'ы деплоя)
