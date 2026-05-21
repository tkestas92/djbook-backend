.PHONY: generate build run docker-up docker-down tidy

generate:
	go run github.com/99designs/gqlgen generate

build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

tidy:
	go mod tidy

docker-up:
	docker compose up --build

docker-down:
	docker compose down -v

# Copy .env.example to .env if it doesn't exist
.env:
	cp .env.example .env
	@echo "Created .env from .env.example — update values before running."
