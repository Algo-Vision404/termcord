VERSION ?= $(shell go run ./internal/version/cmd 2>/dev/null || echo 0.1.0)
LDFLAGS := -s -w -X github.com/termcord/termcord/internal/version.Version=$(VERSION)

.PHONY: build install test clean

build:
	go build -ldflags "$(LDFLAGS)" -o bin/termcord ./cmd/termcord
	go build -ldflags "$(LDFLAGS)" -o bin/termcord-cli ./cmd/termcord-cli

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/termcord
	go install -ldflags "$(LDFLAGS)" ./cmd/termcord-cli

test:
	go test ./...

clean:
	rm -rf bin/
