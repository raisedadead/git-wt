# git-wt Makefile
# For local development, ensure ~/go/bin is in PATH before /usr/local/bin

BINARY := git-wt
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/raisedadead/git-wt/internal/commands.version=$(VERSION)"

# Default install location (Go-idiomatic)
PREFIX ?= $(shell go env GOPATH)
BINDIR := $(PREFIX)/bin

.PHONY: all build install install-global uninstall clean test lint

all: build

# Build to ./bin/
build:
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/git-wt
	@echo "Built bin/$(BINARY) ($(VERSION))"

# Install to ~/go/bin (default) - for development
install: build
	@mkdir -p $(BINDIR)
	cp bin/$(BINARY) $(BINDIR)/$(BINARY)
	@echo "Installed to $(BINDIR)/$(BINARY)"

# Install to /usr/local/bin - overwrites homebrew version
install-global: build
	cp bin/$(BINARY) /usr/local/bin/$(BINARY)
	@echo "Installed to /usr/local/bin/$(BINARY)"

uninstall:
	rm -f $(BINDIR)/$(BINARY)
	@echo "Removed $(BINDIR)/$(BINARY)"

clean:
	rm -rf bin/
	go clean

test:
	go test -v ./...

lint:
	go vet ./...
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping"

# Quick rebuild and show version
dev: build
	@./bin/$(BINARY) --version
