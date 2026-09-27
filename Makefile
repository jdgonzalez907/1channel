SHELL := /bin/bash

DOCKER ?= docker

.PHONY: run up down migrate-up migrate-down migrate-create reset

run:
	set -a; . ./.env; set +a; go run ./cmd/api

up:
	$(DOCKER) compose up -d postgres

down:
	$(DOCKER) compose down

migrate-up:
	$(DOCKER) compose run --rm migrate up

migrate-down:
	$(DOCKER) compose run --rm migrate down 1

migrate-create:
	@test -n "$(NAME)" || (echo "usage: make migrate-create NAME=<name>"; exit 1)
	$(DOCKER) compose run --rm migrate create -ext sql -dir /migrations -seq $(NAME)

reset:
	$(DOCKER) compose down -v
	$(DOCKER) compose up -d postgres
	$(DOCKER) compose run --rm migrate up
