test:
	go test ./...

race:
	go test -race ./...

all: test
	go build -o zxcoin.o ./cmd/zxcoin

lint:
	golangci-lint run --max-issues-per-linter=0 --max-same-issues=0

migrate:
	migrate -source file://migrations -database 'sqlite://zxcoin.db' up
