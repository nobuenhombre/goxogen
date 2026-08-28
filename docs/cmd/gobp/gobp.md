# gobp (cmd) — точка входа build progress утилиты

## Назначение

Пакет `package main` — исполняемая точка входа приложения **gobp** (утилита отображения прогресса сборки Go-приложений). Собирает приложение через Google Wire, выполняет быструю проверку версии до инициализации зависимостей и делегирует выполнение `DomainService.Run()`.

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
// IApp — верхнеуровневый оркестратор приложения gobp.
type IApp interface {
	Run() error
}

// App — верхнеуровневый оркестратор gobp.
type App struct {
	dom domainapp.DomainService
}

// newApp — Wire-provider верхнеуровневого приложения gobp.
func newApp(dom domainapp.DomainService) (IApp, func(), error)
```

`Run()` транслирует вызов напрямую в `a.dom.Run()` — пайплайн отображения прогресса сборки; дополнительной логики в App нет.

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
 └─ app.Run() → dom.Run()        — запуск пайплайна сборки с прогресс-баром
```

Ошибки: при ошибке `initializeApp()` — `log.Fatalf("Initialization error: %v")`; при ошибке `Run()` — `log.Fatalf("Application error: %v")`.

## Wire-граф

```
cli.ProviderSet → domainapp.ProviderSet → newApp(IApp)
```

Wire cleanup вызывается в обратном порядке создания: `cleanup3` (App) → `cleanup2` (Domain) → `cleanup` (CLI).

## Факты о коде и примечания

- Идентичен по структуре main-пакету goxogen (идентичные main.go, применяется та же 3-шаговая схема: panic recovery → version check → Wire).
- Проверка версии — перебор `os.Args[1:]` со сравнением строк, до любого Wire-кода. Срабатывает на оба написания флага (`-version`, `--version`).
- `newApp()` возвращает cleanup, который только пишет `log.Println("[wire-cleanup] Gobp App cleanup")` — реальных ресурсов на уровне App нет.
- В комментариях и логах приложение именуется "Gobp" (capital B), в то время как имя пакета/бинарника — `gobp`.

## Кросс-ссылки

- [CLI-конфигурация gobp](../../internal/app/gobp/cli/cli.md)
- [Domain-слой gobp (прогресс сборки)](../../internal/app/gobp/domain/domain.md)
- [Версия gobp](../../internal/app/gobp/version/version.md)
- [Общий прогресс-бар](../../internal/pkg/progress-bar/progress-bar.md)