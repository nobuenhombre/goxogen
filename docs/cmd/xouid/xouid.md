# xouid (cmd) — точка входа генератора XOID

## Назначение

Пакет `package main` — исполняемая точка входа приложения **xouid** (генератор Go-кода из SQL-запросов PostgreSQL с валидацией через EXPLAIN). Собирает приложение через Google Wire (цепочка из 4 провайдеров, включая пул подключений PostgreSQL), выполняет быструю проверку версии и делегирует выполнение `DomainService.Run()`.

## Состав пакета (4 файла)

| Файл | Назначение |
|------|------------|
| `main.go` | Sequence: panic recovery → `-version` check → `initializeApp()` → `app.Run()` |
| `app.go` | Интерфейс `IApp`, структура `App{dom: DomainService}` и Wire-provider `newApp()` |
| `wire.go` | `//go:build wireinject` — декларация `wire.Build()` без логики |
| `wire_gen.go` | **Автосгенерирован** (`Wire. DO NOT EDIT.`) — реализация `initializeApp()` |
| `AGENTS.md` | Контекст агента для этого каталога |

## Публичный контракт

```go
// IApp — верхнеуровневый оркестратор приложения xouid.
type IApp interface {
	Run() error
}

// App — верхнеуровневый оркестратор xouid.
type App struct {
	dom domainapp.DomainService
}

// newApp — Wire-provider верхнеуровневого приложения xouid.
func newApp(dom domainapp.DomainService) (IApp, func(), error)
```

`Run()` транслирует вызов напрямую в `a.dom.Run()` — пайплайн генерации XOID-кода; дополнительной логики в App нет.

## Проводка в приложении

```
main()
 ├─ defer recover() → log.Printf("PANIC: %v\nStack trace: %s", r, debug.Stack())
 ├─ перебор os.Args[1:]: "-version" / "--version" → fmt.Println(version.Version); os.Exit(0)
 ├─ initializeApp()  (Wire, 4 провайдера)
 │    ├─ cli.ProvideCLI()              → CLI-конфигурация (+cleanup)
 │    ├─ postgres.ProvideDB(service)   → пул pgx (pgxdb.DBQuery) (+cleanup2)
 │    ├─ domainapp.ProvideDomain(service, dbQuery) → DomainService (+cleanup3)
 │    └─ newApp(domainService)         → IApp (+cleanup4)
 ├─ defer cleanup()   — порядок cleanup: App → Domain → PostgreSQL pool → CLI
 └─ app.Run() → dom.Run()   — запуск SQL-to-Go генерации
```

Ошибки: при ошибке `initializeApp()` — `log.Fatalf("Initialization error: %v")`; при ошибке `Run()` — `log.Fatalf("Application error: %v")`.

## Wire-граф

```
cli.ProviderSet
    → postgres.ProviderSet     (зависит от cli.Service → GetDSN())
    → domainapp.ProviderSet    (зависит от cli.Service + pgxdb.DBQuery)
    → newApp(IApp)
```

Cleanup в обратном порядке: `cleanup4` (App) → `cleanup3` (Domain) → `cleanup2` (PostgreSQL pool — закрывается последним среди реальных ресурсов) → `cleanup` (CLI).

## Факты о коде и примечания

- Единственный из трёх main-пакетов, у которого Wire-граф состоит из 4 провайдеров: xouid открывает реальное подключение к PostgreSQL (пул pgx) до создания DomainService.
- Проверка версии — перебор `os.Args[1:]` со сравнением строк, до любого Wire-кода и до открытия БД. Срабатывает на оба написания флага (`-version`, `--version`).
- `newApp()` возвращает cleanup, который только пишет `log.Println("[wire-cleanup] Xouid App cleanup")` — реальные ресурсы закрываются cleanup'ами ниже по графу.
- Ошибка на любом шаге Wire-цепочки корректно каскадно вызывает cleanup ранее созданных зависимостей.

## Кросс-ссылки

- [CLI-конфигурация xouid](../../internal/app/xouid/cli/cli.md)
- [PostgreSQL-пул xouid](../../internal/app/xouid/postgres/postgres.md)
- [Domain-слой xouid (генерация)](../../internal/app/xouid/domain/domain.md)
- [Версия xouid](../../internal/app/xouid/version/version.md)