BINARY := tokenpile
CMD := ./cmd/tokenpile
GOGRAPH := $(shell go env GOPATH)/bin/gograph

.PHONY: build test lint fmt generate install uninstall clean release-check check status tools map pack

build:
	go build -o $(BINARY) $(CMD)

test:
	go test -race ./...

lint:
	golangci-lint run --timeout 5m

fmt:
	gofmt -w .

generate:
	go generate ./...

install:
	go install $(CMD)

uninstall:
	rm -f $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)/$(BINARY)

clean:
	rm -f $(BINARY)

release-check:
	goreleaser check

tools:
	go install github.com/ozgurcd/gograph/cmd/gograph@latest

map:
	$(GOGRAPH) build .

pack:
	npx --yes repomix@latest --compress --ignore ".gograph/**" --output .gograph/pack.md

spec-check:
	@command -v openspec >/dev/null 2>&1 || { \
		echo "WARNING: openspec not on PATH, skipping spec validation"; \
		exit 0; \
	}
	openspec validate --specs

check:
	$(MAKE) fmt
	$(MAKE) lint
	$(MAKE) test
	$(MAKE) spec-check
	$(MAKE) map

status:
	@echo "=== Branch ==="
	@git branch --show-current
	@echo ""
	@echo "=== Uncommitted ==="
	@git status --short || true
	@echo ""
	@echo "=== Recent commits ==="
	@git log --oneline -10
	@echo ""
	@echo "=== OpenSpec changes ==="
	@openspec list 2>/dev/null || echo "(openspec not available)"
	@echo ""
	@echo "=== CI (last run) ==="
	@gh run list --limit 1 2>/dev/null || echo "(gh not available)"
