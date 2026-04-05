.PHONY: build build-linux run test clean dev db-up db-down

build:
	go build -o bin/taran ./cmd/taran

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/taran-linux-amd64 ./cmd/taran

run: build
	./bin/taran

test:
	go test ./...

clean:
	rm -rf bin/

dev:
	go run ./cmd/taran

db-up:
	docker compose up -d

db-down:
	docker compose down
