# deployments — Сборка и установка приложений

## Назначение

Makefile'ы для компиляции и установки бинарников `goxogen`, `gobp` и `xouid` на Linux.

## Структура

```
deployments/
├── goxogen/linux/Makefile
├── gobp/linux/Makefile
└── xouid/linux/Makefile
```

## Команды

```bash
make all                    # uninstall → build → install
make build-app              # go build с CGO_ENABLED=0
make build-app-progress     # сборка через gobp с прогресс-баром
make install-app            # установка в /usr/local/bin + /var/log/{app}/
make uninstall-app          # удаление из /usr/local/bin
```

## Ключевые параметры

- `APP_NAME` — имя приложения (goxogen / gobp / xouid)
- `PROJECT_ROOT_PATH=../../..` — путь к корню проекта
- `INSTALL_PATH=/usr/local/bin`
- `APP_BINARY=bin/{app}/linux/{app}`
