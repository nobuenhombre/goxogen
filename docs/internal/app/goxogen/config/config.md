# config — YAML-конфигурация goxogen

## Назначение

Пакет `configapp` — загрузка и сохранение YAML-конфигурации приложения goxogen через `gopkg.in/yaml.v3` и файловый ввод-вывод `suikat/pkg/fico`. Текущая структура `Config` — пустая заглушка без полей: этот пакет является каркасом (scaffold) для будущих секций конфига.

## Состав пакета (5 файлов)

| Файл | Назначение |
|------|------------|
| `config-app.go` | Интерфейс `Service`, пустая структура `Config`, `Load()` / `Save()` / `Get()` |
| `provider.go` | Wire `ProviderSet` + `ProvideConfigApp(cli.Service)` |
| `config-app_test.go` | Тест Load → Save → Read (круг) |
| `config-app_test_load.yaml` | Входная фикстура теста |
| `config-app_test_save.yaml` | Файл, записанный тестом (остаётся после прогона) |
| `AGENTS.md` | Контекст агента для этого каталога |

## Публичный контракт

```go
type Service interface {
	Load(fileName string) error
	Save(fileName string) error
	Get() *Config
}

type Config struct {
	// Add your config sections here
}

func New(fileName string) (Service, error)          // загружает конфиг из файла
func (c *Config) Load(fileName string) error        // читает YAML-файл и unmarshal в Config
func (c *Config) Save(fileName string) error        // marshal Config → пишет YAML-файл
func (c *Config) Get() *Config                      // возвращает сам Config (не копию)
```

## Реализация

- **Load**: `fico.TxtFile(fileName).Read()` → `yaml.Unmarshal` в `c`. Ошибки оборачиваются `ge.Pin(err)`.
- **Save**: `yaml.Marshal(c)` → `txtFile.Write(string(configData))`. Ошибки оборачиваются `ge.Pin(err)`.
- **Get**: возвращает `c` (указатель на тот же экземпляр — мутации через `Get()` влияют на исходный конфиг).
- **New**: создаёт `&Config{}`, вызывает `Load(fileName)`.

## Проводка в приложении

```
main.main()
 └─ initializeApp() (Wire)
      └─ configapp.ProvideConfigApp(cliConfig cli.Service)
           ├─ cleanup: log.Println("App config cleanup")
           └─ path := cliConfig.(*cli.Config).Config   // значение флага -config
                → New(path) → Load(path) → configapp.Service
```

Провайдер принимает `cli.Service`, делает type assertion к `*cli.Config` и передаёт поле `Config` (путь из флага `-config`) в `New()`.

## Факты о коде и примечания

- **Структура Config пустая** — пакет реально ничего не конфигурирует: YAML-файл читается и разбирается, но полей в `Config` нет, поэтому `Load` фактически валидирует только синтаксис YAML. Вся реальная конфигурация пайплайна (БД, пути, `db_name`) живёт в `domain/xo-config.go`, а не здесь.
- `Get()` возвращает указатель на исходный экземпляр — «геттер» без защиты от мутаций (AGENTS.md фиксирует это как особенность).
- Мелкая асимметрия: конструкторы других пакетов возвращают интерфейс; `Config` здесь же реализует `Service` через pointer-ресиверы — `&Config{}` валиден как `Service`.
- Cleanup в `ProvideConfigApp` только логирует; реального ресурса (дескриптора файла) пакет не держит — файл читается целиком в память и закрывается внутри `fico`.
- Единственный тест в репозитории: `config-app_test.go` (круг Load → Save → Read).

## Кросс-ссылки

- [CLI-конфигурация goxogen](../cli/cli.md)
- [Domain-слой goxogen (реальная конфигурация пайплайна — xo-config.go)](../domain/domain.md)
- [Лог-файл goxogen](../log/log.md)
- [Точка входа goxogen](../../../../cmd/goxogen/goxogen.md)