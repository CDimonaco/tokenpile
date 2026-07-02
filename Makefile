BINARY := tokenpile
CMD := ./cmd/tokenpile

.PHONY: build test lint fmt generate install clean release-check

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
