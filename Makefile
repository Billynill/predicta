.PHONY: run build docker-up docker-down docker-logs probe

run:
	go run ./cmd/api

build:
	go build -o bin/predicta ./cmd/api

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f api

probe:
	go run ./cmd/jira-probe
