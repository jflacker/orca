.PHONY: build test lint fmt e2e install snapshot

build:
	go build -o orca ./cmd/orca

test:
	go test -race ./...

lint:
	golangci-lint run

fmt:
	gofumpt -w .

e2e:
	go test -tags e2e ./e2e/... -v

install:
	go install ./cmd/orca

snapshot:
	goreleaser release --snapshot --clean
