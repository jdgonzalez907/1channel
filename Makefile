SHELL := /bin/bash

DOCKER ?= docker
COMPOSE_DEV := $(DOCKER) compose -f backend/docker-compose.yml --project-directory backend
COMPOSE_PROD := $(DOCKER) compose -f docker-compose.prod.yml

.PHONY: run run-api run-web \
	install-web \
	build build-api build-web \
	check check-api check-web \
	up down migrate-up migrate-down migrate-create reset seed \
	prod-up prod-down

# ---------------------------------------------------------------------------
# Run (dev)
# ---------------------------------------------------------------------------

run: ## Levanta backend y frontend a la vez
	$(MAKE) -j2 run-api run-web

run-api: ## Backend en dev (inyecta backend/.env)
	set -a; . ./backend/.env; set +a; go -C backend run ./cmd/api

run-web: ## Frontend en dev (Vite; proxy /v1 -> :8080)
	pnpm --dir frontend dev

# ---------------------------------------------------------------------------
# Install / build
# ---------------------------------------------------------------------------

install-web: ## Instala dependencias del frontend
	pnpm --dir frontend install

build: build-api build-web ## Compila backend y frontend

build-api: ## Compila el binario del backend
	go -C backend build ./cmd/api

build-web: ## Compila el frontend (dist/)
	pnpm --dir frontend build

# ---------------------------------------------------------------------------
# Quality gates
# ---------------------------------------------------------------------------

check: check-api check-web ## Corre los gates de backend y frontend

check-api: ## Gates del backend
	go -C backend mod verify
	gofmt -l backend
	go -C backend vet ./...
	go -C backend build ./...
	go -C backend test -race -count=1 ./...

check-web: ## Gates del frontend
	pnpm --dir frontend type-check
	pnpm --dir frontend build

# ---------------------------------------------------------------------------
# Base de datos de desarrollo
# ---------------------------------------------------------------------------

up: ## Levanta postgres 18 (UTC)
	$(COMPOSE_DEV) up -d postgres

down: ## Detiene los servicios de desarrollo
	$(COMPOSE_DEV) down

migrate-up: ## Aplica migraciones
	$(COMPOSE_DEV) run --rm migrate up

migrate-down: ## Revierte una migracion
	$(COMPOSE_DEV) run --rm migrate down 1

migrate-create: ## Crea una migracion (NAME=<nombre>)
	@test -n "$(NAME)" || (echo "usage: make migrate-create NAME=<name>"; exit 1)
	$(COMPOSE_DEV) run --rm migrate create -ext sql -dir /migrations -seq $(NAME)

reset: ## Recrea la base desde cero
	$(COMPOSE_DEV) down -v
	$(COMPOSE_DEV) up -d postgres
	$(COMPOSE_DEV) run --rm migrate up

seed: ## Carga datos de prueba (SOLO dev; destructivo)
	$(COMPOSE_DEV) exec -T postgres sh -c 'psql -v ON_ERROR_STOP=1 -U "$$POSTGRES_USER" -d "$$POSTGRES_DB"' < backend/db/seed/dev_seed.sql

# ---------------------------------------------------------------------------
# Produccion (web + api; base externa por envs)
# ---------------------------------------------------------------------------

prod-up: ## Levanta el compose de produccion
	$(COMPOSE_PROD) up -d

prod-down: ## Detiene el compose de produccion
	$(COMPOSE_PROD) down

help: ## Muestra esta ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'
