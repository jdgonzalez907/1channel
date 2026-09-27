# Proposal

## Why

El proyecto documenta `docker-compose.yml` como parte del flujo de desarrollo, pero el archivo no existe y el desarrollo local depende de una base PostgreSQL remota (`zimaboard2`) cuyas credenciales viven en un `.env` que la aplicación no lee. No hay forma reproducible de levantar PostgreSQL 18 ni de aplicar las migraciones sin instalar `migrate` en la máquina. Además, la zona horaria del host puede filtrarse en las lecturas: `pgx` v5 decodifica `timestamptz` en `time.Local` cuando `ScanLocation` es nulo, así que el mismo instante se expone con distinto offset según la región del proceso.

## What Changes

- Nuevo `docker-compose.yml` con **PostgreSQL 18** (servidor y logs en UTC) y un servicio `migrate` bajo demanda mediante `profiles: ["tools"]`, para que `docker compose up` no aplique el esquema por sorpresa.
- Nuevo `Makefile` con `run`, `up`, `down`, `migrate-up`, `migrate-down`, `migrate-create` y `reset`. El target `run` inyecta `.env` (incluido `TZ=UTC`) y arranca la app con `go run`.
- **BREAKING** (solo entorno de desarrollo): el `.env` deja de apuntar a la base remota y pasa a ser la configuración local. Se reescriben `.env` y `.env.example` con las variables que el código realmente lee hoy; las credenciales remotas se abandonan.
- `.gitignore` endurecido a `.env`, `.env.*` y `!.env.example`, para que ningún `.env.*` futuro (p. ej. `.env.remote`) pueda commitearse con secretos. Sin `.env.remote`.
- Eliminación de `package.json` y `package-lock.json`: archivos sueltos, sin referencias en el repositorio.
- UTC de extremo a extremo: `pgdb` fuerza `ScanLocation: time.UTC` en el type map de pgx y el servidor PostgreSQL corre con `timezone=UTC` / `log_timezone=UTC`.
- `AGENTS.md` alineado con la realidad (compose existente, puerto, TZ, Makefile y variables realmente usadas).

## Capabilities

### New Capabilities

- Ninguna. Este cambio es, en su mayor parte, infraestructura de desarrollo y no introduce comportamiento de dominio nuevo.

### Modified Capabilities

- `persistence`: se añade el requisito de que los timestamps se almacenen y recuperen como instantes UTC con independencia de la zona horaria del host; antes la zona de lectura dependía de `time.Local` del proceso.

## Impact

- Nuevos archivos: `docker-compose.yml`, `Makefile`.
- Modificados: `.env` (no versionado), `.env.example`, `.gitignore`, `.dockerignore`, `internal/shared/infra/pgdb/pgdb.go`, `AGENTS.md`.
- Eliminados: `package.json`, `package-lock.json`.
- Sin cambios en la API HTTP, en el dominio, en SQLC ni en las migraciones existentes.
- Riesgo de datos: el volumen local es nuevo; no se migran datos de `zimaboard2`, que queda fuera del flujo diario.
- Fuera de alcance (no-goals): túnel/ingress/TLS, webhook de Meta, sender real de WhatsApp, CI con PostgreSQL, pool `POSTGRES_MAX_*` (hoy no se lee desde el código).
