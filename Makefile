test:
	go test ./...

build: test
	go build -o zxcoin.o
