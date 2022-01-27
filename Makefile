.PHONY: run build clean build-spellfix download-sqlite-windows-x64 download-sqlite-linux-x64 download-sqlite-darwin-x64 download-sqlite-darwin-arm64

BIN_NAME=meteorae

UNAME_S := $(shell uname -s)
VERSION = $(shell git describe --tags)
GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_DIRTY = $(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
BUILD_DATE = $(shell date '+%Y-%m-%d-%H:%M:%S')

TAGS = json1,icu

clean:
	rm -f spellfix.so
	rm -f $(BIN_NAME)-linux-x64
	rm -rf bin

build:
	make clean
	make download-spellfix
	make download-web
	@echo "Building ${BIN_NAME}"
ifeq ($(UNAME_S),Darwin)
	export CGO_CFLAGS_ALLOW="-Xpreprocessor"
endif
	export GOOS=windows
	export GOARCH=amd64
	go build -tags ${TAGS} -ldflags "-X github.com/meteorae/meteorae-server/helpers.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/meteorae/meteorae-server/helpers.BuildDate=${BUILD_DATE}" -o bin/windows-x64/$(BIN_NAME)-win-x64.exe main.go
	export GOOS=linux
	export GOARCH=amd64
	go build -tags ${TAGS} -ldflags "-X github.com/meteorae/meteorae-server/helpers.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/meteorae/meteorae-server/helpers.BuildDate=${BUILD_DATE}" -o bin/linux-x64/$(BIN_NAME)-linux-x64 main.go
#	export GOOS=darwin
#	export GOARCH=amd64
#	go build -tags ${TAGS},darwin -ldflags "-X github.com/meteorae/meteorae-server/helpers.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/meteorae/meteorae-server/helpers.BuildDate=${BUILD_DATE}" -o bin/darwin-x64/$(BIN_NAME)-darwin-x64 main.go
#	export GOOS=darwin
#	export GOARCH=arm64
#	go build -tags ${TAGS},darwin -ldflags "-X github.com/meteorae/meteorae-server/helpers.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/meteorae/meteorae-server/helpers.BuildDate=${BUILD_DATE}" -o bin/darwin-arm64/$(BIN_NAME)-darwin-arm64 main.go

run-linux:
	make build
	cp bin/linux-x64/$(BIN_NAME)-linux-x64 .
	cp bin/linux-x64/spellfix.so .
	./$(BIN_NAME)-linux-x64

download-web:
	curl -L https://github.com/meteorae/web/releases/latest/download/web.zip > server/handlers/web/web.zip
	unzip -o server/handlers/web/web.zip -d server/handlers/web/client/
	rm server/handlers/web/web.zip

download-spellfix:
	@echo "Downloading spellfix extension for SQLite"
	make download-sqlite-windows-x64
	make download-sqlite-linux-x64
#	make download-sqlite-darwin-x64
#	make download-sqlite-darwin-arm64

download-sqlite-windows-x64:
	mkdir -p ./bin/windows-x64
	cd ./bin/windows-x64
	curl -L https://github.com/nalgeon/sqlean/releases/latest/download/spellfix.dll --output spellfix.dll
	mv spellfix.dll bin/windows-x64

download-sqlite-linux-x64:
	mkdir -p ./bin/linux-x64
	curl -L https://github.com/nalgeon/sqlean/releases/latest/download/spellfix.so --output spellfix.so
	mv spellfix.so bin/linux-x64

download-sqlite-darwin-x64:
	mkdir -p ./bin/darwin-x64
	curl -L https://github.com/nalgeon/sqlean/releases/latest/download/spellfix.dylib --output spellfix.dylib

download-sqlite-darwin-arm64:
	mkdir -p ./bin/darwin-arm64
	cd ./bin/darwin-arm64
	curl -L https://github.com/nalgeon/sqlean/releases/latest/download/spellfix.dylib --output spellfix.dylib
