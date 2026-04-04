.PHONY: up down seed build

up:
	docker compose up -d --build

down:
	docker compose down


seed:
	POSTGRES_HOST=localhost go run cmd/seeder/main.go

build:
	docker compose build
