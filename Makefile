BINARY := tokenpile
CMD := ./cmd/tokenpile

.PHONY: build test lint fmt generate install clean release-check check status

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

clean:
	rm -f $(BINARY)

release-check:
	goreleaser check

check:
	$(MAKE) fmt
	$(MAKE) lint
	$(MAKE) test

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
