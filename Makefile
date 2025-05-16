.PHONY: all build run test clean swagger

all: swagger build

build:
	go build -o bin/api cmd/api/main.go

run:
	go run cmd/api/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/
	rm -rf docs/

swagger:
	swag init -g cmd/api/main.go -o docs

install-swagger:
	go install github.com/swaggo/swag/cmd/swag@latest 