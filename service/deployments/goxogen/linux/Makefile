#======================================================
# goxogen
#======================================================

APP_NAME=goxogen
APP_LOG_NAME=goxogen

PROJECT_ROOT_PATH=../../../..
INSTALL_PATH=/usr/local/bin

BUILD_PLATFORM=linux
BIN_PATH=bin/$(APP_NAME)/linux

APP_BINARY=$(BIN_PATH)/$(APP_NAME)
APP_INSTALL=$(INSTALL_PATH)/$(APP_NAME)

#=========================================================================
#
#=========================================================================
.PHONY: help build

help: Makefile
	@echo "Выберите опцию сборки "$(BINARY_NAME)":"
	@sed -n 's/^##//p' $< | column -s ':' |  sed -e 's/^/ /'

## all: Скомпилировать приложение и установить
all: uninstall-app build-app install-app

## build-app: Скомпилировать приложение
build-app:
	cd $(PROJECT_ROOT_PATH)/ && \
	go mod tidy && \
	CGO_ENABLED=0 GOOS=$(BUILD_PLATFORM) GOARCH=amd64 go build -ldflags="-s -w" -o $(APP_BINARY) -v ./src/cmd/$(APP_NAME) && \
	chmod +x $(APP_BINARY) && \
	ls -lh $(APP_BINARY) && \
	$(APP_BINARY) --version;

## build-app-progress: Скомпилировать с прогресс-баром
build-app-progress:
	cd $(PROJECT_ROOT_PATH)/ && \
	CGO_ENABLED=0 GOOS=$(BUILD_PLATFORM) GOARCH=amd64 \
	GOFLAGS="-ldflags=-s -w" \
	gobp --full-rebuild -binary ./src/cmd/$(APP_NAME) -out $(APP_BINARY) && \
	chmod +x $(APP_BINARY) && \
	ls -lh $(APP_BINARY) && \
	$(APP_BINARY) --version;

## install-app: Установить приложение
install-app:
	sudo ln -sf $(shell pwd)/$(PROJECT_ROOT_PATH)/$(APP_BINARY) $(APP_INSTALL);

## uninstall-app: Удалить приложение
uninstall-app:
	sudo rm -f $(APP_INSTALL);
