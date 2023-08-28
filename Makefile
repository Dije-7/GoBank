.PHONY: build run test

build:
	@go build -o bin/GoBank

run: build
	@./bin/GoBank

run-seed: build
	@./bin/GoBank -seed

test:
	@go test -v ./...