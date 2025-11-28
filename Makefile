.PHONY: tidy test run-api run-analytics docker-up docker-down

tidy:
	go mod tidy

test:
	go test ./...

run-api:
	go run ./cmd/fin-api

run-analytics:
	go run ./cmd/fin-analytics

docker-up:
	docker compose up --build

docker-down:
	docker compose down -v

