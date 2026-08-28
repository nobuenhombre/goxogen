# domain — бизнес-логика gobp (прогресс сборки)

## Назначение

Пакет `domainapp` — оркестратор gobp: запуск `go build` с прогресс-баром. Сначала dry-run (`go build -n`) подсчитывает число шагов сборки, затем реальная сборка (`go build -x`) отображает прогресс по мере выполнения команд (mkdir, компиляция, линковка) и собирает строки ошибок компилятора.

## Состав пакета (2 файла)

| Файл | Назначение |
|------|------------|
| `domain.go` | `DomainService` + `AppDomain`, `New()`, `countSteps()`, `isBuildCommandLine()`, `Run()` |
| `provider.go` | Wire `ProviderSet` + `ProvideDomain(cli.Service)` |
| `AGENTS.md` | Контекст агента для этого каталога |

## Публичный контракт

```go
type DomainService interface {
	Run() error
}

type AppDomain struct {
	cliConfig cli.Service
}

func New(cliConfig cli.Service) (DomainService, error)
func countSteps(args []string) (int, error)
func isBuildCommandLine(line string) bool
func (d *AppDomain) Run() error
```

## Алгоритм Run()

```
Run()
 ├─ binary := GetBinary(); out := GetOut()
 ├─ commonArgs := ["build"] (+"-a" если full-rebuild) + ["-o", out, binary]
 │    (без -ldflags — линковка через GOFLAGS envvar)
 ├─ Phase 1: totalSteps = countSteps(commonArgs)
 │    └─ "go build -n" (строки из stderr), фильтр isBuildCommandLine
 ├─ если totalSteps == 0:
 │    └─ «Binary is up to date» → обычный "go build" → "[OK]"/"[FAILED]" → return
 ├─ StartLine("Building project", binary) + ProgressState{Total: totalSteps}
 ├─ Phase 2: "go build -x" (вставка -x после "build")
 │    └─ стрим stderr построчно (scanner.Buffer 64KB/1MB):
 │         ├─ isBuildCommandLine(line) → step++ (зажим step ≤ totalSteps),
 │         │    state.Current/Elapsed/Remaining, печать ProgressBar("")
 │         └─ line содержит ".go:" → errorCount++, errorLines += line
 ├─ err := cmd.Wait()
 ├─ при ошибке: "\n" + ErrorLine(errorCount, elapsed) + errorLines построчно
 │    → return ge.Pin(fmt.Errorf("build failed with %d error(s)", errorCount))
 └─ успех: финальный ProgressBar + FinishLine(elapsed)
```

`New()` хранит `cli.Service` (интерфейс, без type assertion).

## Фильтр командных строк (isBuildCommandLine)

Строка считается «шагом сборки», если начинается с:
- `mkdir` — создание каталогов;
- `/` — абсолютный путь: компилятор gcc/ar/compile/link и т.п.;
- `cd ` — смена каталога;
- `cat ` — конкатенация файлов.

Один и тот же фильтр используется и в `countSteps()` (подсчёт), и в прогресс-цикле (отрисовка) — иначе оценки рассинхронизируются.

## Подсчёт шагов (countSteps)

- Клонирует аргументы, вставляет `-n` сразу после `"build"` (аргумент 0).
- `go build -n` пишет команды в **stderr** (не stdout) — читается `StderrPipe`.
- Ошибочный dry-run (`cmd.Wait() != nil`) — ошибка `failed to count build steps`.
- `scanner.Buffer(make([]byte, 64*1024), 1024*1024)` обязателен: длинные gcc/link-строки (>64KB) иначе убивают сканер и `go build -n` виснет на переполненном pipe.

## Факты о коде и примечания

- **Расхождение dry-run vs реальный цикл**: `go build -x` может выдать на 1–5 строк больше, чем `-n` (CGO-диагностика gcc/clang в stderr, начинающаяся с `/`). Симптомы до фиксов: 101% и отрицательный ETA. Защита двойная: `ProgressBar()` зажимает percent ≤ 100 и filled ≤ barLength; здесь `step` зажимается до `totalSteps`.
- Ошибки компилятора собираются по подстроке `.go:` в строке — это может ложно захватывать и не-ошибочные строки (например, пути файлов в выводе `go build -x`), что завышает `errorCount`. Впрочем, реальные ошибки печатаются построчно после `ErrorLine`.
- Прогресс-бар печатается с `\r`; перед строками ошибок печатается `\n`, чтобы коллизия `\r` не съедала вывод.
- `-verbose` из CLI не используется в domain: вывод либо прогресс-бар, либо (при полном отсутствии шагов) короткая строка up-to-date.
- Версия в коде — `v0.9.0` (вложенный AGENTS.md устарел: v0.7.0).
- Единственный рендер: `ProgressState` и render-функции берутся из общего пакета `progress-bar`.

## Кросс-ссылки

- [CLI-конфигурация gobp](../cli/cli.md)
- [Версия gobp](../version/version.md)
- [Пакет progress-bar (ProgressState, render-функции)](../../../pkg/progress-bar/progress-bar.md)
- [Точка входа gobp](../../../../cmd/gobp/gobp.md)