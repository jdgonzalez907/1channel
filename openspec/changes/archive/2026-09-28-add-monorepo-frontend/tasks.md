# Tasks

## 1. Reestructurar el backend bajo `backend/`

- [x] 1.1 Crear `backend/` y mover `go.mod`, `go.sum`, `sqlc.yaml`, `cmd/`, `internal/`, `db/`, `docs/`, `Dockerfile`, `.dockerignore`, `docker-compose.yml` y `.env`/`.env.example` (el `Makefile` queda en la raíz y se reescribe en 4.3); verificar con `git status` que no queda ningún archivo de backend en la raíz salvo el `Makefile` y que el módulo sigue declarando `github.com/jdgonzalez907/1channel`.
- [x] 1.2 Verificar el compose de dev desde `backend/`: `docker compose -f backend/docker-compose.yml --project-directory backend config` no deja variables sin resolver y `docker compose ... up -d postgres` + `... run --rm migrate up` aplican el esquema.
- [x] 1.3 Verificar el backend desde su nuevo directorio: `go -C backend mod verify`, `gofmt -l backend`, `go -C backend vet ./...`, `go -C backend build ./...` y `go -C backend test -race -count=1 ./...` limpios; `sqlc generate` desde `backend/` sin diff.
- [x] 1.4 Ajustar `backend/.dockerignore` al nuevo contexto de build y verificar que `docker build -f backend/Dockerfile backend` produce la imagen del `api`.
- [x] 1.5 Crear `backend/.gitignore` (patrones Go, `.env`/`.env.*`, `!.env.example`) y reducir el `.gitignore` raíz a lo global (`.idea/`, `.vscode/`, `.DS_Store`, `*.out`, `coverage.*`); verificar con `git check-ignore -v backend/.env` y que `backend/.env.example` no queda ignorado.

## 2. Shell del frontend y su imagen `web`

- [x] 2.1 Crear `frontend/` con el template oficial de Vue (`pnpm create vue@latest`, TypeScript sin features extra), agregar el proxy de dev de `/v1` y fijar `packageManager`; verificar que `pnpm install` genera `pnpm-lock.yaml` y que `pnpm --dir frontend build` produce `frontend/dist/` con `index.html`.
- [x] 2.2 Añadir `frontend/.dockerignore` (`node_modules/`, `dist/`, `.env*`, `.git/`), `frontend/.gitignore` y `frontend/.env.example` (destinado a futuras `VITE_*`); verificar que el contexto de build no incluye `node_modules` ni el `.env`.
- [x] 2.3 Configurar el proxy de desarrollo de Vite (`/v1` → `http://localhost:8080`) y verificar con `pnpm --dir frontend dev` que una llamada a `/v1/...` llega al backend local.
- [x] 2.4 Escribir `frontend/Dockerfile` multi-stage (`node:24-alpine` con `corepack enable` → `nginx:1.27-alpine`) y verificar que `docker build -f frontend/Dockerfile frontend` produce la imagen `web`.
- [x] 2.5 Escribir `frontend/nginx.conf` con `listen 80`, `root`, `try_files $uri $uri/ /index.html` y `location /v1/ { proxy_pass http://api:8080; }` **sin barra final** más los headers `Host`/`X-Forwarded-*`; verificar contra un `api` en la misma red que `GET /` da 200 con el cliente, `GET /ruta/profunda` da 200 con `index.html` y `GET /v1/...` llega al `api` con el prefijo intacto.

## 3. CI por componente con filtro de `paths`

- [x] 3.1 Adaptar `.github/workflows/ci.yml` a `.github/workflows/backend.yml`: triggers `paths: ['backend/**', '.github/workflows/backend.yml']`, `defaults.run.working-directory: backend`, `go-version-file: backend/go.mod`, los gates actuales y publicación de `ghcr.io/jdgonzalez907/1channel/api:<sha>` + `:latest`; verificar el YAML y que un commit que solo toca `backend/**` lo dispara.
- [x] 3.2 Crear `.github/workflows/frontend.yml` (nuevo) con triggers `paths: ['frontend/**', '.github/workflows/frontend.yml']`, `pnpm/action-setup` + `actions/setup-node` con cache, `pnpm install --frozen-lockfile`, `pnpm type-check` + `pnpm build` y publicación de `ghcr.io/jdgonzalez907/1channel/web:<sha>` + `:latest`; verificar el YAML y que un commit que solo toca `frontend/**` lo dispara.
- [x] 3.3 Verificar que ya no existe `.github/workflows/ci.yml` y que no quedan referencias al workflow único ni a la imagen `1channel:<sha>` en el repositorio.
- [x] 3.4 Verificar que un commit que no toca `backend/**` ni `frontend/**` no dispara ninguno de los dos workflows (revisando los `paths` declarados).

## 4. Compose de producción y Makefile raíz

- [x] 4.1 Escribir `docker-compose.prod.yml` con `web` (imagen `.../web:${WEB_TAG:-latest}`, `ports: ["${WEB_PORT:-80}:80"]`, healthcheck `wget` contra `/`) y `api` (imagen `.../api:${API_TAG:-latest}`, `expose: ["8080"]` sin `ports`, `env_file: [.env]`, healthcheck `wget` contra `/healthz`), `depends_on: api: condition: service_healthy`, sin servicios de base de datos ni de migraciones; verificar con `docker compose -f docker-compose.prod.yml config` que resuelve y que `config --services` lista solo `web` y `api`.
- [x] 4.2 Verificar en ejecución que `api` no publica puerto al host y que `web` sí, que `web` no arranca antes de que `api` esté `healthy`, y que `API_TAG=<sha> docker compose -f docker-compose.prod.yml up -d --no-deps api` recrea solo `api`.
- [x] 4.3 Reescribir el `Makefile` raíz como interfaz única con esquema `verbo`/`verbo-api`/`verbo-web` (`run`, `run-api`, `run-web`, `build`, `build-api`, `build-web`, `check`, `check-api`, `check-web`, `install-web`, `up`, `down`, `migrate-*`, `reset`, `seed`, `prod-up`, `prod-down`) delegando en `backend/docker-compose.yml` con `-f backend/docker-compose.yml --project-directory backend` y en `pnpm --dir frontend`; verificar cada target ejecutándose tal cual.

## 5. Documentación e integración

- [x] 5.1 Actualizar `AGENTS.md` a la estructura monorepo (rutas `backend/`/`frontend/`, `.env` por proyecto, un Makefile raíz, workflows por componente, compose de producción en `:80`, Cloudflare Tunnel como TLS, `api` interno en `:8080`, contrato `/v1`); verificar que cada comando documentado se ejecuta tal cual.
- [x] 5.2 Documentar el orden de despliegue (migrar primero, desplegar después) y el comando de migración manual contra `$POSTGRES_URL` (`docker run --rm ... migrate/migrate -path=/migrations -database "$POSTGRES_URL" up`); verificar el comando contra una base externa de prueba.
- [x] 5.3 Prueba de integración: levantar `docker-compose.prod.yml` con una base externa, verificar `GET /` (200 cliente), `GET /ruta/profunda` (200 `index.html`) y `GET /v1/...` (llega al `api`), y confirmar que desplegar un sha distinto por componente no rompe ninguno de los dos servicios.
