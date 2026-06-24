export GO111MODULE=on


BIN := aur-cli

GO ?= go
GOFLAGS := -v
EXTRA_GOFLAGS ?=
LDFLAGS := $(LDFLAGS) -X main.version=dev -X main.builtBy=makefile

.PHONY: default
default: build

.PHONY: clean
clean:
	$(GO) clean $(GOFLAGS) -i ./...
	rm -rf $(BIN)

.PHONY: lint
lint:
	golangci-lint run

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: test
test:
	$(GO) test $(GOFLAGS) -race ./... -coverprofile=.coverage.out
	go tool cover -func=.coverage.out

.PHONY: build
build: $(BIN)

$(BIN): $(SOURCES)
	$(GO) build $(GOFLAGS) -ldflags '-s -w $(LDFLAGS)' $(EXTRA_GOFLAGS) -o $@ ./cmd/$(BIN)
