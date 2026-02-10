.PHONY: build run test clean dev db-up db-down

build:
	go build -o bin/taran ./cmd/taran

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
