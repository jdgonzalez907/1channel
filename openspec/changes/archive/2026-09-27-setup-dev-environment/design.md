# Design

## Context

Ver `proposal.md` - Why para la motivación. Restricciones observadas en el código y el entorno:

- La app lee el entorno solo con `os.Getenv` (`cmd/api/config.go`); no carga `.env`, y `go.mod` no trae ningún dotenv.
- El `Dockerfile` es de producción (multi-stage, binario estático, usuario no root, sin hot reload). Su imagen final ya instala `ca-certificates` y `tzdata`.
- `pgdb.New` construye el pool con `pgxpool.ParseConfig(dsn)` y no toca el `TypeMap`.
- En pgx v5.11.0, `TimestamptzCodec` decodifica el formato binario con `time.Unix(...)` (ubicación `time.Local`) y solo aplica `.In(loc)` si `ScanLocation != nil`. Verificado en `pgtype/timestamptz.go`.
- El puerto 5432 quedó libre en el host; existía otro contenedor Postgres usándolo. El usuario del shell no pertenece al grupo `docker`.

## Goals / Non-Goals

**Goals:**

- Un `docker compose up` reproducible que levante PostgreSQL 18 en UTC, sin depender de `zimaboard2`.
- Aplicar migraciones sin instalar `migrate` en la máquina, de forma explícita (no automática).
- Iterar la app en segundos (`go run`), manteniendo intacto el `Dockerfile` de producción.
- Que la zona horaria del host no afecte lo almacenado ni lo leído.

**Non-Goals:**

- Llevar la app dentro de compose (implica rebuild por cambio o `Dockerfile.dev` + hot reload).
- Túnel/ingress/TLS, webhook de Meta y sender real de WhatsApp.
- Migrar datos desde `zimaboard2` o gestionar sus credenciales.
- Cablear `POSTGRES_MAX_*` ni cambios de API/dominio.

## Decisions

### D1 - Compose solo con datos; la app corre en el host

`docker-compose.yml` levanta solo `postgres` (y `migrate` bajo demanda). La app se ejecuta con `go run ./cmd/api`.

- Razón: el `Dockerfile` actual no tiene hot reload; reconstruir la imagen en cada cambio vuelve el ciclo editar→probar mucho más lento. Correr la app en Docker no es más lento en ejecución (mismo kernel), pero sí en iteración.
- Alternativas descartadas: app en compose con la imagen de producción (rebuild lento); `Dockerfile.dev` + `air` (config extra que hoy nada aprovecha, porque no hay webhook ni integración entrante).

### D2 - PostgreSQL 18 con volumen en la ruta correcta

Imagen `postgres:18-alpine`. El volumen nombrado se monta en `/var/lib/postgresql`, **no** en `/var/lib/postgresql/data`.

- Razón: en 18+ `PGDATA` es `/var/lib/postgresql/18/docker` y el `VOLUME` de la imagen es `/var/lib/postgresql`. Montar en `.../data` deja los datos fuera del volumen y se pierden al recrear el contenedor.
- Servidor en UTC con `command: postgres -c timezone=UTC -c log_timezone=UTC` y `TZ`/`PGTZ` en UTC.
- Puerto publicado `5432:5432` (estándar; el host está libre). Es configurable vía `POSTGRES_PORT`.
- `healthcheck` con `pg_isready` para que `migrate` espere a que la base acepte conexiones.
- Mapeo de nombres: la imagen usa `POSTGRES_USER`/`POSTGRES_DB`; la app usa `POSTGRES_USERNAME`/`POSTGRES_DATABASE`. El compose los mapea desde `.env`.

### D3 - Migraciones bajo demanda, no automáticas

Servicio `migrate` (imagen `migrate/migrate`) con `profiles: ["tools"]`, de modo que `docker compose up` no lo arranca. Se invoca con `docker compose run --rm migrate ...` y se envuelve en targets del Makefile.

- Razón: `up` no debe mutar el esquema por sorpresa; y `migrate` no está instalado en el host. El volumen `./db/migrations` se monta de escritura para que `migrate create` funcione.
- Alternativas descartadas: `docker-entrypoint-initdb.d` (no versiona ni rastrea `schema_migrations`); migrar en el arranque de la app (acopla esquema a runtime, fuera del alcance).

### D4 - `.env` local como fuente de configuración; sin `.env.remote`

`.env` (gitignored) pasa a apuntar a la base local. `.env.example` (versionado) queda con placeholders. `.gitignore` pasa a `.env`, `.env.*`, `!.env.example`.

- Razón de seguridad: el patrón actual `.env` no cubre `.env.*`; un `.env.remote` habría sido commiteable con secretos. En vez de crear ese archivo, las credenciales remotas se abandonan y no se guardan en disco.
- No se crea `.env.remote`. Las credenciales de `zimaboard2` se descartan del flujo diario.
- Contenido: solo variables que el código lee hoy (`HTTP_PORT`, `LOG_LEVEL`, `POSTGRES_USERNAME/PASSWORD/HOST/PORT/DATABASE`) más `TZ=UTC`. Se eliminan `ONECHANNEL_SECRET`, `META_*`, `WHATSAPP_*` y `POSTGRES_MAX_*` (sin referencias en Go).

### D5 - UTC en tres capas

1. Servidor PostgreSQL en UTC (D2).
2. `TZ=UTC` inyectado por `make run` al proceso de la app.
3. `ScanLocation: time.UTC` en el `TypeMap` de pgx vía `AfterConnect` al crear el pool en `pgdb.New`.

- Razón: (1) y (2) son entorno; (3) es robusto aunque alguien corra `go run` sin el Makefile y su `time.Local` sea otra zona. Las escrituras ya usan `time.Now().UTC()` en los handlers.
- Alternativa considerada y descartada como única: solo `TZ=UTC` (frágil ante ejecución sin Makefile); solo `timezone=UTC` en el DSN (afecta la sesión, no la ubicación de decodificación binaria).

### D6 - Limpieza de archivos sueltos

Se eliminan `package.json` (`{}`) y `package-lock.json` (`packages: {}`). No hay referencias a npm/npx/node en el repo; el CLI de OpenSpec es un binario pnpm global.

### D7 - Makefile como interfaz del flujo

Targets: `run`, `up`, `down`, `migrate-up`, `migrate-down`, `migrate-create NAME=...`, `reset`. `run` hace `set -a; . ./.env; set +a` y luego `go run ./cmd/api`, que es lo que hace efectivo `TZ=UTC`.

- Alternativas descartadas: documentar el `source` a mano (propenso a olvidos); loader dotenv dentro de la app (mete lógica de entorno en runtime); `direnv` (dependencia externa).

## Risks / Trade-offs

- [Ejecutar `go run` sin `make run` pierde `TZ=UTC`] → mitigado por `ScanLocation: time.UTC` en pgx; la lectura sigue siendo UTC.
- [Montar el volumen en `/var/lib/postgresql/data` en PG18] → usar `/var/lib/postgresql`; está fijado en D2.
- [Password con caracteres especiales rompe el DSN del `migrate`] → en local los valores son `dev/dev`; si cambian, URL-encodear. `POSTGRES_URL` puede usarse como override.
- [Otro proyecto ocupa 5432] → `POSTGRES_PORT` es configurable; por defecto 5432.
- [El shell no está en el grupo `docker`] → operar con `sudo` o agregar el usuario al grupo; no afecta el diseño de los artefactos.
- [Repuntar `.env` es BREAKING para el flujo de desarrollo] → documentado en `AGENTS.md` y centralizado en el Makefile.
- [Sin test automatizado de la lectura en UTC: CI no levanta Postgres] → el requisito queda como escenario verificable; un test de integración con base real queda fuera de alcance.
- [Datos previos en `zimaboard2` quedan inaccesibles para el flujo diario] → intencional; si se necesitan, se exportan aparte.

## Migration Plan

1. Añadir `docker-compose.yml`, `Makefile`, `.env.example` (local) y reescribir `.env`.
2. Endurecer `.gitignore` y `.dockerignore`.
3. Aplicar el cambio de UTC en `pgdb.New`.
4. Eliminar `package.json` y `package-lock.json`.
5. `docker compose up -d postgres` y `make migrate-up` sobre volumen vacío.
6. `make run` y verificar `/healthz` y un alta de usuario.
7. Rollback: `make reset` (borra volumen) y revertir los archivos de este cambio; no hay migración de datos que deshacer.

## Open Questions

- Fijar el tag exacto de la imagen `migrate/migrate` (hoy sin versión); no cambia el enfoque ni las tareas.
