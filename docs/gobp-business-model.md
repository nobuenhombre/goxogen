# Бизнес-модель: gobp — сборка Go-приложений с прогрессом

## Зачем это всё

`go build` большого приложения (особенно с CGO) молчит минуты: компиляция, линковка, копирование — без какой-либо обратной связи. Непонятно, идёт ли сборка, сколько шагов осталось, сколько времени займёт. Ошибки компилятора при этом перемешиваются с выводом и теряются.

gobp оборачивает `go build` и показывает **прогресс с ETA** по фактическим шагам сборки (mkdir, компиляция gcc/compile, линковка), собирает строки ошибок компилятора и выводит их сводкой.

## Замысел (Vision)

Одна команда `gobp -binary ./src/cmd/myapp -out ./build/app` заменяет немой `go build` на:

1. **Подсчёт шагов** — dry-run (`go build -n`) определяет число шагов сборки;
2. **Прогресс** — реальная сборка (`go build -x`) отображает шаги в баре с временем и ETA;
3. **Ошибки** — агрегированные строки ошибок (по подстроке `.go:`) выводятся сводкой под баром.

Флаг `-full-rebuild` форсирует полную пересборку (`go build -a`); линковка настраивается через envvar `GOFLAGS="-ldflags=..."` (не через флаги).

## Предметная область

### 1. Входы

| Вход | Источник | Примечание |
|------|----------|------------|
| Go-пакет/бинарник | флаг `-binary` (default `.`) | путь к `./src/cmd/...` |
| Выходной путь | флаг `-out` (default `./build/app`) | |
| Полная пересборка | флаг `-full-rebuild` | добавляет `-a` к `go build` |
| ldflags | envvar `GOFLAGS` | не хардкодится в аргументах |

### 2. Процесс (Run())

```
Run()
 ├─ binary := GetBinary(); out := GetOut()
 ├─ commonArgs := ["build"] (+"-a" если full-rebuild) + ["-o", out, binary]
 ├─ Phase 1: totalSteps = countSteps(commonArgs)      // "go build -n", stderr
 │    └─ фильтр isBuildCommandLine (mkdir, /, cd , cat )
 ├─ если totalSteps == 0 → «Binary is up to date» → обычный go build → [OK]/[FAILED]
 ├─ Phase 2: "go build -x" — стрим stderr построчно (scanner 64KB/1MB):
 │    ├─ isBuildCommandLine(line) → step++ (зажим ≤ totalSteps), печать ProgressBar
 │    └─ line содержит ".go:" → errorCount++, errorLines += line
 ├─ cmd.Wait()
 ├─ ошибка: "\n" + ErrorLine + все errorLines → "build failed with N error(s)"
 └─ успех: финальный ProgressBar + FinishLine
```

Фильтр `isBuildCommandLine` одинаков в подсчёте и в цикле отрисовки — иначе числа рассинхронизируются.

### 3. Выходы

| Выход | Описание |
|-------|----------|
| Собранный бинарник | по пути `-out` |
| Прогресс-бар | `[▮▮▮░░░] N% (cur/total) ⏱ mm:ss ⌛ mm:ss` (с `\r`, без переноса строки) |
| Сводка | `✅ Done in N seconds!` или `✖ N Errors detected` + строки ошибок построчно |

### 4. Место в экосистеме

- **Deployment-Makefile'ы** (`service/deployments/{app}/linux/Makefile`, цель `build-app-progress`): сборка goxogen/gobp/xouid для установки в `/usr/local/bin` идёт через `gobp --full-rebuild -binary ./src/cmd/{app} -out bin/{app}/linux/{app}` с `GOFLAGS="-ldflags=-s -w"`.
- **progress-bar** — общий пакет: `ProgressState` + `StartLine`/`FinishLine`/`ErrorLine`; gobp не использует `ProgressTracker`/окно xo.

## Ключевые принципы

1. **Честный прогресс** — процент зажимается до 100, заполнение — до длины бара (dry-run может недо-подсчитать из-за CGO-диагностики).
2. **Ошибки не теряются** — строки с `.go:` агрегируются и печатаются отдельно под баром, с `\n` перед выводом (коллизия с `\r`).
3. **Робастность сканера** — буфер 64KB/1MB обязателен: длинные gcc/link-строки (>64KB) иначе убивают `bufio.Scanner`.
4. **Минимальный интерфейс** — 4 рабочих флага; линковка — через `GOFLAGS`, а не аргументы.

## Целевая аудитория

| Роль | Что получает |
|------|-------------|
| Разработчик | Видимость хода сборки тяжёлого (CGO) приложения, ETA, сводку ошибок |
| Deployment-процесс | `build-app-progress`: сборка с прогресс-баром для всех трёх приложений репозитория |

## Архитектурные драйверы

| Потребность | Решение |
|------------|---------|
| Немой `go build` без обратной связи | Dry-run подсчёт (`go build -n`) + реальный прогресс (`go build -x`) |
| Расхождение dry-run vs real (CGO добавляет строки) | Зажимы: percent ≤ 100, filled ≤ barLength, `step` ≤ `totalSteps` |
| Длинные строки gcc/link >64KB | `scanner.Buffer(make([]byte, 64*1024), 1024*1024)` |
| `\r` бара съедает строки ошибок | Печать `\n` перед блоками ошибок |
| Ошибки компилятора вперемешку с выводом | Агрегация по `.go:` + сводка `ErrorLine` |

## Факты о коде и примечания

- **`-verbose` мёртв**: геттер `GetVerbose()` есть в `cli.Service`, но не вызывается в domain — флаг не влияет на поведение.
- **Завышение errorCount**: ошибки собираются по подстроке `.go:` — пути файлов в выводе `go build -x` тоже могут попасть в счётчик (сами ошибки при этом печатаются корректно).
- **Версия**: в коде `v0.9.0`, AGENTS.md/README указывают `v0.7.0` — устаревшие данные.
- При `totalSteps == 0` (всё уже собрано) gobp не рисует бар — печатает короткую строку «Binary is up to date».
- В комментариях/логах приложение именуется «Gobp» (capital B), имя пакета/бинарника — `gobp`.

## Кросс-ссылки

- [Карта документации gobp](gobp-index.md)
- [Domain — прогресс сборки (детали)](internal/app/gobp/domain/domain.md)
- [CLI-конфигурация gobp](internal/app/gobp/cli/cli.md)
- [progress-bar](internal/pkg/progress-bar/progress-bar.md)
- [goxogen — бизнес-модель](goxogen-business-model.md)