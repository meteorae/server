.PHONY: run build clean download-external build-spellfix

BIN_NAME=meteorae

UNAME_S := $(shell uname -s)
VERSION=$(shell git describe --tags)
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
BUILD_DATE=$(shell date '+%Y-%m-%d-%H:%M:%S')
TAGS=json1,icu

clean:
	rm -rf bin
	rm -f spellfix.c

build:
	make clean
	make download-external
	make build-spellfix
	@echo "Building ${BIN_NAME}"
	go build -tags ${TAGS} -ldflags "-X github.com/meteorae/meteorae-server/helpers.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/meteorae/meteorae-server/helpers.BuildDate=${BUILD_DATE}" -o bin/$(BIN_NAME) main.go

run:
	make build
	./bin/$(BIN_NAME)

download-external:
	@echo "Downloading external dependencies"
	curl -L https://github.com/sqlite/sqlite/raw/master/ext/misc/spellfix.c --output spellfix.c

build-spellfix:
	@echo "Building spellfix extension for SQLite"
ifeq ($(OS),Windows_NT)
	gcc -shared spellfix.c -o libgo-sqlite3-spellfix.dll
else
ifeq ($(UNAME_S),Linux)
	gcc -fPIC -shared spellfix.c -o libgo-sqlite3-spellfix.so
endif
ifeq ($(UNAME_S),Darwin)
	gcc -fPIC -dynamiclib spellfix.c -o libgo-sqlite3-spellfix.dylib
endif
endif

