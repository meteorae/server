.PHONY: run build clean build-linux build-windows build-darwin-intel build-darwin-apple run-linux run-windows run-darwin-intel run-darwin-apple

BIN_NAME=meteorae

UNAME_S := $(shell uname -s)
UNAME_P := $(shell uname -p)
VERSION = $(shell git describe --tags)
GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_DIRTY = $(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
BUILD_DATE = $(shell date '+%Y-%m-%d-%H:%M:%S')

TAGS = json1,icu

clean:
	rm -f $(BIN_NAME)-linux-x64
	rm -f $(BIN_NAME)-win-x64.exe
	rm -f $(BIN_NAME)-darwin-intel
	rm -f $(BIN_NAME)-darwin-apple
	rm -rf bin

build:
	make clean
	@echo "Building ${BIN_NAME}"
ifeq ($(UNAME_S),Darwin)
	export CGO_CFLAGS_ALLOW="-Xpreprocessor"
endif
	make build-linux
	make build-windows
	make build-darwin-intel
	make build-darwin-apple

build-windows:
	export GOOS=windows
	export GOARCH=amd64
	go build -tags ${TAGS} -ldflags "-X github.com/meteorae/meteorae-server/helpers.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/meteorae/meteorae-server/helpers.BuildDate=${BUILD_DATE}" -o bin/windows-x64/$(BIN_NAME)-win-x64.exe main.go

build-linux:
	export GOOS=linux
	export GOARCH=amd64
	go build -tags ${TAGS} -ldflags "-X github.com/meteorae/meteorae-server/helpers.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/meteorae/meteorae-server/helpers.BuildDate=${BUILD_DATE}" -o bin/linux-x64/$(BIN_NAME)-linux-x64 main.go

build-darwin-intel:
	export GOOS=darwin
	export GOARCH=amd64
	go build -tags ${TAGS} -ldflags "-X github.com/meteorae/meteorae-server/helpers.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/meteorae/meteorae-server/helpers.BuildDate=${BUILD_DATE}" -o bin/darwin-x64/$(BIN_NAME)-darwin-intel main.go

build-darwin-apple:
	export GOOS=darwin
	export GOARCH=arm64
	go build -tags ${TAGS} -ldflags "-X github.com/meteorae/meteorae-server/helpers.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/meteorae/meteorae-server/helpers.BuildDate=${BUILD_DATE}" -o bin/darwin-arm64/$(BIN_NAME)-darwin-apple main.go

run:
ifeq ($(OS),Windows_NT)
	make run-windows
else
ifeq ($(UNAME_S),Linux)
	make run-linux
endif
ifeq ($(UNAME_S),Darwin)
ifeq ($(UNAME_P),x86_64)
	make run-darwin-intel
else
	make run-darwin-apple
endif
endif
endif

run-linux:
	@echo "Building ${BIN_NAME}"
	make build-linux
	cp bin/linux-x64/$(BIN_NAME)-linux-x64 .
	./$(BIN_NAME)-linux-x64

run-windows:
	@echo "Building ${BIN_NAME}"
	make build-windows
	cp bin/windows-x64/$(BIN_NAME)-win-x64.exe .
	./$(BIN_NAME)-win-x64.exe

run-darwin-intel:
	@echo "Building ${BIN_NAME}"
	make build-darwin-intel
	cp bin/darwin-x64/$(BIN_NAME)-darwin-x64 .
	./$(BIN_NAME)-darwin-x64

run-darwin-apple:
	@echo "Building ${BIN_NAME}"
	make build-darwin-apple
	cp bin/darwin-arm64/$(BIN_NAME)-darwin-arm64 .
	./$(BIN_NAME)-darwin-arm64

download-web:
	curl -L https://github.com/meteorae/web/releases/latest/download/web.zip > server/handlers/web/web.zip
	unzip -o server/handlers/web/web.zip -d server/handlers/web/client/
	rm server/handlers/web/web.zip
