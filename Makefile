VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BINARY  := linux-webui
MODULE  := github.com/virajchitnis/linux-webui

LDFLAGS := -X main.version=$(VERSION) -s -w
GOFLAGS := -tags production

.PHONY: all build dev test lint e2e e2e-docker clean

all: build

build: ui/dist
	go mod verify
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/linux-webui

ui/dist: web/node_modules
	cd web && npm run build

web/node_modules:
	cd web && npm install

dev:
	@echo "Start 'cd web && npm run dev' in another terminal, then:"
	LINUX_WEBUI_DEV=1 go run ./cmd/linux-webui --config config.dev.toml

test:
	go vet ./...
	go test -race ./...

lint:
	golangci-lint run ./...

e2e:
	cd e2e && npx playwright test

e2e-docker:
	docker build -f Dockerfile.e2e -t linux-webui-e2e .
	docker run --rm linux-webui-e2e

clean:
	rm -f $(BINARY)
	rm -rf web/dist
