ifneq (,$(wildcard .env))
include .env
export $(shell sed 's/=.*//' .env)
endif

MIGRATIONS_DIR ?= db/migrations

.PHONY: migrate-create migrate-up migrate-down migrate-down-all migrate-force migrate-version test-integration test docker-build podman-build

migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

migrate-up:
	@test -n "$(DATABASE_DSN)" || (echo "DATABASE_DSN is required"; exit 1)
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_DSN)" up

migrate-down:
	@test -n "$(DATABASE_DSN)" || (echo "DATABASE_DSN is required"; exit 1)
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_DSN)" down 1

migrate-down-all:
	@test -n "$(DATABASE_DSN)" || (echo "DATABASE_DSN is required"; exit 1)
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_DSN)" down

migrate-force:
	@test -n "$(DATABASE_DSN)" || (echo "DATABASE_DSN is required"; exit 1)
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_DSN)" force $(v)

migrate-version:
	@test -n "$(DATABASE_DSN)" || (echo "DATABASE_DSN is required"; exit 1)
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_DSN)" version

test-integration:
	dotenv run go test -v ./...

test:
	dotenv run go test -v ./...

docker-build:
	docker build -t knowledger:latest .

podman-build:
	podman build -t knowledger:latest .