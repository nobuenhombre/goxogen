export GOPROXY := https://proxy.golang.org,direct
export GOSUMDB := off

include configs/_make_/config/project.mk
include configs/_make_/config/go-build.mk
include configs/_make_/lib/go-build/cache-ram-drive.mk
include configs/_make_/lib/go-build/progress-bar.mk

#===========================================
# команды makefile -
# если команда совпадет с названием каталога
#===========================================
.PHONY: help deps run build all lint test fmt wire build-app-progress-gobp

help: Makefile
	@echo "Выберите опцию сборки:"
	@sed -n 's/^##//p' $< | column -s ':' |  sed -e 's/^/ /'

## deps: Инициализация модулей, скачать все необходимые программе модули
deps:
	rm -f go.mod
	rm -f go.sum
	go mod init $(PROJECT_NAME)
	go get -u ./...

## wire: Генерация wire_gen.go через Google Wire
wire:
	wire ./src/...
	gofmt -w src/cmd/*/wire_gen.go
