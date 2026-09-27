# Tasks

## 1. Configuración local y limpieza

- [x] 1.1 Reescribir `.env` (no versionado) y `.env.example` con `HTTP_PORT`, `LOG_LEVEL`, `TZ=UTC` y `POSTGRES_*`, retirando `ONECHANNEL_SECRET`, `META_*`, `WHATSAPP_*` y `POSTGRES_MAX_*`; verificar que `set -a; . ./.env; set +a` exporta todas las variables y que `.env.example` no contiene secretos.
- [x] 1.2 Endurecer `.gitignore` a `.env`, `.env.*` y `!.env.example`; verificar con `git check-ignore -v .env.remote` que lo ignora y con `git check-ignore -v .env.example` que no lo ignora.
- [x] 1.3 Añadir `.env*` a `.dockerignore`; verificar que el build no incorpora archivos de entorno.
- [x] 1.4 Eliminar `package.json` y `package-lock.json`; verificar que no quedan referencias (`grep -rn -i -E 'npm|npx|package.json' .`) y que `go build ./...` sigue pasando.

## 2. Docker Compose (PostgreSQL 18 + migrate)

- [x] 2.1 Escribir `docker-compose.yml` con `postgres:18-alpine`, volumen nombrado en `/var/lib/postgresql` (no `.../data`), `timezone=UTC`/`log_timezone=UTC`, `TZ`/`PGTZ` en UTC, `healthcheck` con `pg_isready` y mapeo `POSTGRES_USER=${POSTGRES_USERNAME}` / `POSTGRES_DB=${POSTGRES_DATABASE}`; verificar con `docker compose config` que no hay variables sin resolver.
- [x] 2.2 Añadir el servicio `migrate` (`profiles: ["tools"]`, volumen `./db/migrations:/migrations`) con `depends_on` al healthcheck; verificar que `docker compose config --services` lo liste y que `docker compose up -d` no lo arranque.
- [x] 2.3 Levantar `docker compose up -d postgres`; verificar que `docker compose ps` lo reporta `healthy` y que `docker compose exec postgres psql -U ... -c 'SHOW timezone;'` responde `UTC`.

## 3. Makefile

- [x] 3.1 Implementar `run`, `up`, `down`, `migrate-up`, `migrate-down`, `migrate-create NAME=...` y `reset`; verificar que `make up` deja el contenedor `healthy`.
- [x] 3.2 Verificar que `make migrate-up` aplica las cuatro migraciones sobre volumen limpio (`000001..000004`) y que `make migrate-down` revierte sin error, con `schema_migrations` presente.
- [x] 3.3 Verificar que `make migrate-create NAME=prueba` crea un par de archivos en `db/migrations/` y luego eliminar ese par.

## 4. UTC en el límite de persistencia

- [x] 4.1 Configurar `pgdb.New` con un `AfterConnect` que registre `&pgtype.TimestamptzCodec{ScanLocation: time.UTC}` en el `TypeMap` de la conexión; verificar `go build ./...` y `go vet ./...` limpios.
- [x] 4.2 Agregar un test unitario en `pgdb` que codifique y decodifique un `timestamptz` con un mapa configurado como el de producción y verifique que la ubicación resultante es `time.UTC` (sin base de datos); verificar con `go test -race -count=1 ./internal/shared/infra/pgdb/...`.

## 5. Documentación e integración

- [x] 5.1 Actualizar `AGENTS.md` (existencia de `docker-compose.yml`, puerto 5432, `TZ=UTC`, targets del Makefile y variables realmente usadas, quitando referencias sin respaldo); verificar que cada comando documentado se ejecuta tal cual.
- [x] 5.2 Sobre volumen limpio, correr `make up`, `make migrate-up` y `make run`; verificar `GET /healthz` responde 200 y `POST /v1/users` da de alta un usuario.
- [x] 5.3 Verificar el escenario de UTC de la spec: con Postgres en UTC y el proceso con `TZ` distinto de UTC (invocando `go run` sin el Makefile), comprobar que un timestamp leído se expone en UTC.
