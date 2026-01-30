.PHONY: help test test-race lint lint-fast lint-local vet build all windows-amd linux-amd linux-386 darwin-amd darwin-arm

define helpMessage
possible targets:
- test
- test-race
- lint
- lint-fast
- lint-local
- vet
- build
- all
- windows-amd
- linux-amd
- linux-386
- darwin-amd
- darwin-arm
endef
export helpMessage

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILT   ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  = -s -w \
  -X github.com/zeropsio/zaia/internal/commands.version=$(VERSION) \
  -X github.com/zeropsio/zaia/internal/commands.commit=$(COMMIT) \
  -X github.com/zeropsio/zaia/internal/commands.built=$(BUILT)

help:
	@echo "$$helpMessage"

test:
	go test -v ./... -count=1

test-race:
	go test -race ./... -count=1

lint:
	GOOS=darwin GOARCH=arm64 golangci-lint run ./... --verbose
	GOOS=linux GOARCH=amd64 golangci-lint run ./... --verbose
	GOOS=windows GOARCH=amd64 golangci-lint run ./... --verbose

lint-fast: ## Fast lint (native platform, fast mode)
	golangci-lint run ./... --fast

lint-local: ## Full lint (native platform)
	golangci-lint run ./...

vet:
	go vet ./...

build:
	go build -ldflags "$(LDFLAGS)" -o bin/zaia ./cmd/zaia

#########
# BUILD #
#########
all: windows-amd linux-amd linux-386 darwin-amd darwin-arm

windows-amd:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o builds/zaia-win-x64.exe ./cmd/zaia/main.go

linux-amd:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o builds/zaia-linux-amd64 ./cmd/zaia/main.go

linux-386:
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags "$(LDFLAGS)" -o builds/zaia-linux-i386 ./cmd/zaia/main.go

darwin-amd:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o builds/zaia-darwin-amd64 ./cmd/zaia/main.go

darwin-arm:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o builds/zaia-darwin-arm64 ./cmd/zaia/main.go
