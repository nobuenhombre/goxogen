# progress-bar — общий ANSI-прогресс-бар

## Назначение

Пакет `progressbar` — разделяемая библиотека отрисовки прогресса в терминале, используемая обоими приложениями: **goxogen** (пайплайн кодогенерации с окном вывода xo-субпроцессов) и **gobp** (прогресс сборки `go build -x`). Содержит ANSI-константы, render-функции (однострочный бар и финальные сообщения) и жизненный цикл `ProgressTracker`.

## Состав пакета (1 файл)

| Файл | Назначение |
|------|------------|
| `progress-bar.go` | Константы ANSI/символов, `ProgressState`, render-функции, `ProgressTracker` |
| `AGENTS.md` (src) | Контекст агента для этого каталога |

## Публичный контракт

```go
// Цвета ANSI
const (
	ColorTitle, ColorProject, ColorSuccess, ColorError, ColorWarning,
	ColorTime, ColorEta, ColorReset, ColorXoLine string
)
// Управление курсором
const (AnsiCursorUp, AnsiClearLine, AnsiEraseBelow string)
// Размер окна xo
const XoWindowLines = 5
// Символы бара
const (charDone, charPending, charOk, charError, charWarning, charTime, charEta rune)

// Состояние прогресс-бара
type ProgressState struct {
	Title, ProjectName string
	Current, Total     int
	Elapsed, Remaining int
	Errors             int
	StartTime          int64
}
func (p *ProgressState) ProgressBar(currentFile string) string

// Render-функции
func StartLine(title, project string) string
func FinishLine(elapsed int) string
func ErrorLine(errors, elapsed int) string

// Жизненный цикл
type ProgressTracker struct { ... }
func NewProgressTracker(title string, total int) *ProgressTracker
func (pt *ProgressTracker) Increment(currentFile string)
func (pt *ProgressTracker) AddError(line string)
func (pt *ProgressTracker) Finish()
func (pt *ProgressTracker) Fail()
func (pt *ProgressTracker) PushXoLine(line string)
func (pt *ProgressTracker) ClearXoOutput()
```

## Рендер (ProgressBar)

Одна строка с `\r` (без перевода строки): `[████░░░░] N% (cur/total) ⏱ mm:ss ⌛ mm:ss ▶ file`.

- percent = `(Current * 100) / Total`, **зажимается до 100**; filled = `(barLength * Current) / Total`, barLength = 50, **зажимается до barLength** (защита от переполнения, когда реальных шагов больше подсчитанных — см. gobp).
- ETA = `elapsed * (Total-Current) / Current`; при `Current <= 1` показывается `--:--`.
- `currentFile` усекается до ~35 символов (добавляется префикс `...`).
- Пустые pending-сегменты рисуются красным `✖`-цветом (`charPending` = `━` c `ColorError`).

## Render-функции

| Функция | Вывод |
|---------|-------|
| `StartLine(title, project)` | `<новые строки> <title> [<project>]` |
| `FinishLine(elapsed)` | `✅ Done in N seconds!` |
| `ErrorLine(errors, elapsed)` | `✖ N Errors detected` + `⚠ Stopped after N seconds` |

## ProgressTracker и окно xo (goxogen)

Внутри: `ProgressState` + `startTime` + `errors []string` + кольцевой буфер `xoLines [5]string` (`XoWindowLines = 5`).

- **Increment(file)**: `Current++` (зажим до Total), расчёт Elapsed/Remaining, печать бара. Если окно xo активно — перерисовка всего 6-строчного блока (`renderWithXoLines`) вместо однострочного бара.
- **PushXoLine(line)**: сдвиг кольцевого буфера вверх (новые строки — снизу), перерисовка блока. Строки усекаются до 100 символов.
- **ClearXoOutput()**: стирает окно xo (курсором вверх/вниз + очистка строк), сбрасывает `xoActive`/`xoCount`, перерисовывает голый бар — экран не пустует между шагами.
- **Fail()**: при активном окне сначала чистит его, затем `\n` + `ErrorLine` + все накопленные `errors` построчно.
- **Finish()**: Current = Total, финальный бар + `FinishLine`.

Первый рендер окна чистит всё ниже бара (`AnsiEraseBelow`), последующие — курсор вверх на 5 строк (`AnsiCursorUp`) и перезапись in-place.

## Использование в приложениях

### gobp (прямое использование)

```
ProgressState{Title: "Building project", ProjectName: binary, Total: totalSteps}
→ state.ProgressBar("") после каждой командной строки go build -x
→ при ошибке: "\n" + ErrorLine(errorCount, elapsed) + строки ".go:"
→ успех: последний ProgressBar + FinishLine
```

### goxogen (через ProgressTracker)

```
NewProgressTracker("Generating code", totalSteps)   // project name жёстко "xoxgen"
→ pt.Increment("<имя шага>") на каждый шаг пайплайна
→ runXO стримит строки xo через pt.PushXoLine (фильтр isXoObjectName)
→ ClearXoOutput() между блоками xo
→ ошибка: pt.AddError + pt.Fail()
→ успех: pt.Finish()
```

## Факты о коде и примечания

- **Общие символы и константы**: ANSI-цвета и render-функции — единственный источник для обоих приложений; конвенция — не дублировать их в app-пакетах.
- `NewProgressTracker` жёстко проставляет `ProjectName: "xoxgen"` (опечатка «xoxgen» вместо «goxogen» в коде).
- `ProgressBar` возвращает строку с `\r` и печатается через `fmt.Print` (не Println): перед любым многострочным выводом (ошибки компилятора, error lines) вызывающий код обязан печатать `\n`, иначе строки съедаются.
- Зажимы percent ≤ 100 и filled ≤ barLength — защита от расхождения dry-run/real (CGO-диагностика gcc добавляет строки в `go build -x`).
- `charPending` рисуется `ColorError` — «пустая» часть бара красная, «готовая» зелёная; семантика стиля, а не ошибки.
- Строка бара не пишет в stdout логи: goxogen дополнительно переключает `log.SetOutput(io.Discard)` на время пайплайна.
- `XoWindowLines` и кольцевой буфер — специфика goxogen; gobp использует только `ProgressState` + render-функции.

## Кросс-ссылки

- [Domain-слой goxogen (ProgressTracker + окно xo)](../../app/goxogen/domain/domain.md)
- [Domain-слой gobp (ProgressState + render)](../../app/gobp/domain/domain.md)
- [Сборка gobp](../../../cmd/gobp/gobp.md)