# goxogen (cmd) — точка входа scaffolder'а

## Назначение

Пакет `package main` — исполняемая точка входа приложения **goxogen** (code generation scaffolder). Собирает приложение через Google Wire, выполняет быструю проверку версии до инициализации зависимостей и делегирует выполнение `DomainService.Run()`.

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
// IApp — верхнеуровневый оркестратор приложения.
type IApp interface {
	Run() error
}

// App — верхнеуровневый оркестратор.
type App struct {
	dom domainapp.DomainService
}

// newApp — Wire-provider верхнеуровневого приложения.
func newApp(dom domainapp.DomainService) (IApp, func(), error)
```

`Run()` транслирует вызов напрямую в `a.dom.Run()` — никакой дополнительной логики в App нет.

## Проводка в приложении

```
main()
 ├─ defer recover() → log.Printf("PANIC: %v\nStack trace: %s", r, debug.Stack())
 ├─ перебор os.Args[1:]: "-version" / "--version" → fmt.Println(version.Version); os.Exit(0)
 ├─ initializeApp()  (Wire)
 │    ├─ cli.ProvideCLI()        → CLI-конфигурация (+cleanup)
 │    ├─ domainapp.ProvideDomain(service) → DomainService (+cleanup)
 │    └─ newApp(domainService)   → IApp (+cleanup)
 ├─ defer cleanup()              — порядок cleanup: App → Domain → CLI
 └─ app.Run() → dom.Run()        — запуск пайплайна кодогенерации
```

Ошибки: при ошибке `initializeApp()` — `log.Fatalf("Initialization error: %v")`; при ошибке `Run()` — `log.Fatalf("Application error: %v")`.

## Wire-граф

```
cli.ProviderSet → domainapp.ProviderSet → newApp(IApp)
```

Wire cleanup вызывается в обратном порядке создания: `cleanup3` (App) → `cleanup2` (Domain) → `cleanup` (CLI).

## Факты о коде и примечания

- Проверка версии — перебор `os.Args[1:]` со сравнением строк, до любого Wire-кода и до открытия БД/логов. Срабатывает на оба написания флага (`-version`, `--version`).
- `newApp()` возвращает не-пустой cleanup, который ничего не освобождает — только пишет `log.Println("[wire-cleanup] App cleanup")`. Реальные ресурсы (при их наличии) закрываются cleanup'ами ниже по графу.
- `wire_gen.go` сгенерирован с Go-версией сигнатуры `provideDomain` как `domainapp.ProvideDomain` — в `wire.go` пакет импортирован под алиасом `domainapp`, в сгенерированном файле алиас не используется, но импорт `"goxogen/src/internal/app/goxogen/domain"` присутствует.

## Кросс-ссылки

- [CLI-конфигурация goxogen](../../internal/app/goxogen/cli/cli.md)
- [Domain-слой goxogen (пайплайн)](../../internal/app/goxogen/domain/domain.md)
- [Версия goxogen](../../internal/app/goxogen/version/version.md)
- [Конфиг YAML goxogen](../../internal/app/goxogen/config/config.md)