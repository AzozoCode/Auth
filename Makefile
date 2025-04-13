build:
	@go build -o bin/auth

run: build
	@./bin/auth

test:
	@go test -v ./...

