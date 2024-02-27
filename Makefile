NAME := quail-view
BUILD_VERSION ?= 0.0.2

SHELL := /bin/bash

# build program for local OS and windows
build:
	@echo "build: building to bin/${NAME}..."
	go build -o bin/${NAME} main.go

# run program
run:
	@echo "run: running..."
	#source .env && go run main.go $$EQ_PATH/dbx.eqg
	#source .env && go run . $$EQ_PATH/it13968.eqg
	#source .env && go run . $$EQ_PATH/luc.eqg
	#source .env && go run . $$EQ_PATH/crushbone.s3d
	source .env && go run . $$EQ_PATH/gequip5.s3d
	#source .env && go run . $$EQ_PATH/wrm.eqg
	#source .env && go run . $$EQ_PATH/steamfontmts.eqg
	#source .env && go run . $$EQ_PATH/global5_chr.s3d

view-%:
	@echo "view-$*: running..."
	source .env && go run . $$EQ_PATH/$*

# bundle program with windows icon
bundle:
	@echo "if go-winres is not found, run go install github.com/tc-hib/go-winres@latest"
	@echo "bundle: setting program icon"
	go-winres simply --icon program.png

# run tests that aren't flagged for SINGLE_TEST
.PHONY: test
test:
	@echo "test: running tests..."
	@mkdir -p test
	@rm -rf test/*
	@go test ./...

# build all supported os's
build-all: build-darwin build-windows build-linux build-windows-addon

build-darwin:
	@echo "build-darwin: ${BUILD_VERSION}"
	@mkdir -p bin
	@CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -trimpath -buildmode=pie -ldflags="-X main.Version=${BUILD_VERSION} -s -w" -o bin/${NAME}-darwin main.go

build-linux:
	@echo "build-linux: ${BUILD_VERSION}"
	@mkdir -p bin
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-X main.Version=${BUILD_VERSION} -s -w" -o bin/${NAME}-linux main.go

build-windows:
	@echo "build-windows: ${BUILD_VERSION}"
	@mkdir -p bin
	@CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -trimpath -buildmode=pie -ldflags="-X main.Version=${BUILD_VERSION} -s -w" -o bin/${NAME}.exe main.go

# run pprof and dump 4 snapshots of heap
profile-heap:
	@echo "profile-heap: running pprof watcher for 2 minutes with snapshots 0 to 3..."
	@-mkdir -p bin
	curl http://localhost:6060/debug/pprof/heap > bin/heap.0.pprof
	sleep 30
	curl http://localhost:6060/debug/pprof/heap > bin/heap.1.pprof
	sleep 30
	curl http://localhost:6060/debug/pprof/heap > bin/heap.2.pprof
	sleep 30
	curl http://localhost:6060/debug/pprof/heap > bin/heap.3.pprof

# peek at a heap
profile-heap-%:
	@echo "profile-heap-$*: use top20, svg, or list *word* for pprof commands, ctrl+c when done"
	go tool pprof bin/heap.$*.pprof

# run a trace on program
profile-trace:
	@echo "profile-trace: getting trace data, this can show memory leaks and other issues..."
	curl http://localhost:6060/debug/pprof/trace > bin/trace.out
	go tool trace bin/trace.out

# run sanitization against golang
sanitize:
	@echo "sanitize: checking for errors"
	rm -rf vendor/
	go vet -tags ci ./...
	test -z $(goimports -e -d . | tee /dev/stderr)
	-go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	gocyclo -over 99 .
	golint -set_exit_status $(go list -tags ci ./...)
	staticcheck -go 1.14 ./...
	go test -tags ci -covermode=atomic -coverprofile=coverage.out ./...
    coverage=`go tool cover -func coverage.out | grep total | tr -s '\t' | cut -f 3 | grep -o '[^%]*'`

# CICD triggers this
set-version-%:
	@echo "VERSION=${BUILD_VERSION}.$*" >> $$GITHUB_ENV
	@export BUILD_VERSION=${BUILD_VERSION}.$*
