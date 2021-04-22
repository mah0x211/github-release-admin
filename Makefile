#
# Makefile
# Created by Masatoshi Fukunaga on 21/4/4
#
LINT_OPT=--issues-exit-code=0 \
		--enable-all \
		--tests=false \
		--disable=funlen \
		--disable=gochecknoinits \
		--disable=gochecknoglobals \
		--disable=gocognit \
		--disable=goconst \
		--disable=godox \
		--disable=goerr113 \
		--disable=gomnd \
		--disable=lll \
		--disable=maligned \
		--disable=prealloc \
		--disable=wsl \
		--exclude=ifElseChain

.EXPORT_ALL_VARIABLES:

.PHONY: all test build lint coverage clean dist

all: test build

test:
	go test -timeout 1m -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html coverage.out -o coverage.out.html

lint:
	golangci-lint run $(LINT_OPT) ./...

coverage: test
	go tool cover -func=coverage.out

build:
	go build -o build/github-release-create cmd/create/main.go
	go build -o build/github-release-delete cmd/delete/main.go
	go build -o build/github-release-download cmd/download/main.go
	go build -o build/github-release-list cmd/list/main.go

dist: build
	tar -C build/ -zcvf build/github-release-create.tar.gz github-release-create
	tar -C build/ -zcvf build/github-release-delete.tar.gz github-release-delete
	tar -C build/ -zcvf build/github-release-download.tar.gz github-release-download
	tar -C build/ -zcvf build/github-release-list.tar.gz github-release-list

clean:
	go clean
