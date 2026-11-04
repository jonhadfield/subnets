GO        ?= go
BIN       ?= subnets
MODULE    := github.com/jonhadfield/subnets
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS   := -X $(MODULE)/internal/cli.Version=$(VERSION)

.PHONY: build test install clean

build:
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/subnets

test:
	$(GO) test ./...

install:
	$(GO) install -ldflags "$(LDFLAGS)" ./cmd/subnets

clean:
	rm -f $(BIN)
