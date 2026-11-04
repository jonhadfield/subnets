GO        ?= go
BIN       ?= subnets
MODULE    := github.com/jonhadfield/subnets

.PHONY: build test install clean

build:
	$(GO) build -o $(BIN) ./cmd/subnets

test:
	$(GO) test ./...

install:
	$(GO) install ./cmd/subnets

clean:
	rm -f $(BIN)
