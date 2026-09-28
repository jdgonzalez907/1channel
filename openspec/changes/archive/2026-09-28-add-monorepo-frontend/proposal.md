# Proposal

## Why

Hoy el repositorio es solo backend Go: un módulo en la raíz, un único workflow de CI y una sola imagen `ghcr.io/<repo>:<sha>` que se despliega en el VPS. Incorporar un SPA en ese molde obliga a reconstruir y redesplegar el backend cada vez que cambia el front (y al revés), y ata el despliegue de uno a la versión del otro. El objetivo es que front y back vivan en un mismo repositorio pero **evolucionen, se construyan y se desplieguen de forma independiente** en el mismo VPS.

## What Changes

- **Reorganización a monorepo**: el módulo Go (`go.mod`, `go.sum`, `cmd/`, `internal/`, `db/`, `docs/`, `sqlc.yaml`, `Dockerfile`, `.dockerignore`, compose, Makefile y `.env`/`.env.example` de desarrollo) se mueve a `backend/`. **BREAKING** para rutas del repo y comandos documentados; no cambia el comportamiento del dominio ni de la API.
- **Configuración y artefactos por proyecto**: cada componente tiene su propio `.env`/`.env.example` (el de producción vive en el VPS, fuera del repo), su `.dockerignore` y su `.gitignore`. `docs/endpoints.md` se mueve a `backend/docs/`.
- **Nuevo `frontend/`**: shell mínimo **Vue 3 + TypeScript + Vite + pnpm** (lo indispensable para producir `dist/`) con su propio `Dockerfile` multi-stage (`node` build → `nginx:alpine`). La UI real queda fuera de alcance.
- **Dos workflows de CI con filtro de `paths`**: `backend.yml` (solo `backend/**`) y `frontend.yml` (solo `frontend/**`). Cada uno construye y publica **su** imagen. Un push que solo toca un lado no reconstruye el otro.
- **Dos imágenes en ghcr**: `.../api:<sha>` y `.../web:<sha>`, cada una con su `:latest` de componente como puntero a su último build. El despliegue es por sha **por componente** (sin semver).
- **Nuevo `docker-compose.prod.yml`** en la raíz: solo `web` (nginx, HTTP en `:80`, único publicado) y `api` (interno en `:8080`, sin publicar), ambos con **healthcheck**. **Sin TLS en los contenedores**: el borde es HTTP y Cloudflare Tunnel termina HTTPS. La base de datos queda **fuera** del compose y se configura por envs (`POSTGRES_URL` o `POSTGRES_*`). Sin servicio de migraciones.
- **Borde nginx**: el contenedor `web` sirve la SPA en `/` con fallback de history y proxea `/v1/*` a `http://api:8080` (mismo origen, sin CORS); el `proxy_pass` va **sin barra final** para conservar el prefijo. Deja el borde listo para futuros `/webhooks/*` y `/ws`, que también entrarían por nginx hacia el `api` interno.
- **CI**: se adapta el workflow existente (`.github/workflows/ci.yml`) al monorepo y se agrega uno nuevo para el front.
- **Migraciones**: en dev siguen por `make` (sin cambios); en producción se ejecutan a mano contra la base externa, fuera del ciclo de despliegue.
- **Documentación**: `AGENTS.md` actualizado a la nueva estructura y flujo.

## Capabilities

### New Capabilities

- `delivery`: cómo se construyen, sirven y despliegan de forma independiente el SPA (`web`) y la API (`api`) — hosting de la SPA con fallback de history, enrutado de `/v1` al API por el mismo origen, builds y releases independientes por componente, y un compose de producción con base de datos externa configurada por entorno.

### Modified Capabilities

- Ninguna. Este cambio no altera requisitos de dominio ni del contrato HTTP existente.

## Impact

- **Estructura**: se mueven `cmd/`, `internal/`, `db/`, `docs/`, `go.mod`, `go.sum`, `sqlc.yaml`, `Dockerfile`, `.dockerignore`, `docker-compose.yml`, `Makefile` y `.env`/`.env.example` a `backend/`. Nuevos: `frontend/` (con `Dockerfile`, `.dockerignore`, `nginx.conf`, `.gitignore`, `.env.example`), `docker-compose.prod.yml`, `.github/workflows/backend.yml` y `.github/workflows/frontend.yml`.
- **CI/CD**: `.github/workflows/ci.yml` se adapta a `backend.yml` y se agrega `frontend.yml`; cambian los nombres de imagen en ghcr (el ajuste del despliegue en el VPS lo hace el usuario).
- **Entorno**: nuevas variables de producción para `api` (base de datos externa por envs; `POSTGRES_URL` con su `sslmode`). El front, al ser mismo origen, no necesita URL de API configurable.
- **Sin cambios** en dominio, API HTTP (`/v1`), SQLC, migraciones ni esquema.
- **No-goals**: la UI real del SPA, autenticación o sesión en el front, webhooks entrantes, WebSocket, TLS/certificados en los contenedores (los termina Cloudflare Tunnel) y runners de migración automáticos en el despliegue.
