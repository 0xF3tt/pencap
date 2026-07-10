BINARY   := pencap
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -s -w -X main.version=$(VERSION)
DIST     := dist

PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64

# Fixed garble seed (8 bytes, base64) so obfuscated builds are cached and
# rebuilds after the first are fast. Rotate per release for a fresh symbol
# mapping, or set GARBLE_SEED=random for a new mapping every build (slow).
GARBLE_SEED ?= EN6+Q8qVrm4=

.PHONY: help build test vet fmt lint audit install clean release tools

help:
	@grep -E '^## ' Makefile | sed 's/^## //'

## build: compile a stripped, trimmed, static binary for the host platform
build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) .

## test: run unit tests
test:
	go test ./...

## vet: go vet
vet:
	go vet ./...

## fmt: check formatting (fails if gofmt would change anything)
fmt:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "not gofmt-ed:"; echo "$$unformatted"; exit 1; \
	fi

## lint: full static-analysis gate (fmt + vet + staticcheck + golangci-lint)
lint: fmt vet
	staticcheck ./...
	golangci-lint run ./...

## audit: security scan gate (gosec + govulncheck)
audit:
	gosec -quiet ./...
	govulncheck ./...

## install: build and install to GOPATH/bin
install:
	CGO_ENABLED=0 go install -trimpath -ldflags "$(LDFLAGS)" .

## clean: remove build artifacts
clean:
	rm -rf $(BINARY) $(DIST)

## tools: install the dev/CI tooling (garble, staticcheck, gosec, govulncheck)
## golangci-lint isn't reliably `go install`-able; use its installer script
## or brew: https://golangci-lint.run/welcome/install/
tools:
	go install mvdan.cc/garble@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

## release: run the full check gate, then cross-compile obfuscated binaries
## (via garble) for each platform in PLATFORMS, so a dropped/recovered
## binary doesn't hand a client's blue team readable symbol names for the
## engagement tooling.
release: test lint audit
	@command -v garble >/dev/null || (echo "garble not found: make tools" && exit 1)
	@mkdir -p $(DIST)
	@for p in $(PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; \
		out=$(DIST)/$(BINARY)-$(VERSION)-$$os-$$arch; \
		[ "$$os" = "windows" ] && out=$$out.exe; \
		echo "building $$out"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch garble -tiny -literals -seed=$(GARBLE_SEED) build -trimpath -ldflags "$(LDFLAGS)" -o $$out . || exit 1; \
	done
	@cd $(DIST) && (sha256sum * > SHA256SUMS 2>/dev/null || shasum -a 256 * > SHA256SUMS)
	@cat $(DIST)/SHA256SUMS
