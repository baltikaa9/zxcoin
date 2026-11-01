all: build

test:
	go test ./...

race:
	go test -race ./...

lint: test
	golangci-lint run --max-issues-per-linter=0 --max-same-issues=0

build: lint
	go build -o zxcoin.o ./cmd/zxcoin

migrate:
	migrate -source file://migrations -database 'sqlite://zxcoin.db' up
