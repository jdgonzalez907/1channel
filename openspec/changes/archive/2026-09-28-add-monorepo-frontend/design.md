# Design

## Context

Ver `proposal.md` - Why para la motivación. Restricciones y estado observados en el repositorio y el entorno:

- El módulo Go vive en la raíz (`module github.com/jdgonzalez907/1channel`), con `cmd/`, `internal/`, `db/`, `docs/`, `sqlc.yaml`, `Dockerfile`, `.dockerignore`, `docker-compose.yml`, `Makefile`, `.env` y `.env.example` en la raíz. Los imports internos usan la ruta del módulo, no rutas relativas al directorio.
- `.github/workflows/ci.yml` construye una única imagen `ghcr.io/${{ github.repository }}:latest|<sha>` tras pasar los quality gates (gofmt, vet, build, test) y usa `go-version-file: go.mod`.
- `docker-compose.yml` es de desarrollo: `postgres` + `migrate` (perfil `tools`). La app corre en el host con `go run` (ver el change archivado `setup-dev-environment`).
- `cmd/api/config.go` arma el DSN desde `POSTGRES_URL` o, en su defecto, desde `POSTGRES_*`. La ruta `POSTGRES_*` fija `sslmode=disable`; para una base externa con TLS hay que usar `POSTGRES_URL`. `HTTP_PORT` por defecto es `8080`.
- `cmd/api/router.go` expone `GET /healthz` y agrupa todo lo demás bajo `/v1`.
- El entorno dev tiene Go 1.27, Node 24 y pnpm 11; `sqlc` y `migrate` viven en `$(go env GOPATH)/bin`.
- No existe frontend. No hay tags de release; el despliegue usa el sha del build de CI.
- Restricciones del usuario: mismo VPS y misma red Docker para front y back; despliegue con GitHub Container Registry ya resuelto; base de datos externa (o compartida del VPS) al compose, configurada por envs **inyectadas al ejecutar el compose**; migraciones manuales en prod y `make` en dev; despliegue por sha por componente; un `.env` por proyecto; `.gitignore` por proyecto; un solo `Makefile` en la raíz; sin certificados en los contenedores (TLS lo termina Cloudflare Tunnel); `web` en `:80` y `api` en `:8080` (interno).

## Goals / Non-Goals

**Goals:**

- Un repositorio con dos componentes (`backend`, `frontend`) que se construyen y despliegan por separado.
- Un borde único (`web`) que sirve el SPA y enruta `/v1` al `api` interno; mismo origen, sin CORS.
- Compose de producción con solo `web` + `api`; base de datos externa por entorno.
- Conservar y adaptar los quality gates actuales del backend y sumar los del frontend.
- Dejar el borde preparado (sin implementar) para `/webhooks/*` y `/ws`.

**Non-Goals:**

- UI real del SPA, autenticación/sesión en el front, cliente tipado del API.
- Implementar webhooks entrantes ni WebSocket; solo se reserva el enrutado en el borde.
- TLS/certificados en los contenedores: el borde es HTTP y Cloudflare Tunnel termina HTTPS.
- Migraciones automáticas, base de datos gestionada o runners en el despliegue.
- Versionado semántico o tags de release; la identidad sigue siendo el sha.
- Gestionar el registry de GitHub: ya funciona y queda como está.

## Decisions

### D1 - Layout del monorepo

```
1channel/
├── backend/
│   ├── go.mod  go.sum  sqlc.yaml
│   ├── cmd/  internal/  db/  docs/
│   ├── Dockerfile  .dockerignore
│   ├── docker-compose.yml        # dev: postgres + migrate
│   ├── .env  .env.example        # config de dev del backend
│   └── .gitignore
├── frontend/
│   ├── package.json  pnpm-lock.yaml  tsconfig.json  vite.config.ts
│   ├── index.html  src/
│   ├── Dockerfile  .dockerignore  nginx.conf
│   ├── .env  .env.example        # config del front (futuras VITE_*)
│   └── .gitignore
├── docker-compose.prod.yml       # web + api (sin DB)
├── Makefile                      # interfaz unica (un solo Makefile)
├── .gitignore                    # minimo global (editor/SO/cobertura)
├── .github/workflows/{backend,frontend}.yml
├── openspec/
└── AGENTS.md
```

- Razón: los filtros de `paths` de CI necesitan una frontera de carpeta limpia (`backend/**` / `frontend/**`). Si el Go se queda en la raíz, el patrón del backend sería `'**', '!frontend/**'`, frágil.
- `docs/endpoints.md` se mueve a `backend/docs/`: documenta la API, no el monorepo.
- **`.gitignore` por proyecto** (`backend/`, `frontend/`) para lo propio de cada build, más un `.gitignore` raíz **mínimo** para lo global que no pertenece a ningún componente (`.idea/`, `.vscode/`, `.DS_Store`, `*.out`, `coverage.*`). Así nada queda sin cubrir al mover archivos.
- El módulo Go conserva la ruta `github.com/jdgonzalez907/1channel` en `backend/go.mod`; mover el directorio **no** cambia los imports internos. `sqlc generate`, `docker compose` y `make` deben ejecutarse desde su contexto correspondiente (el Makefile raíz se encarga).
- Alternativa descartada: `apps/api` + `apps/web` (un nivel de anidación sin beneficio para dos componentes).

### D2 - Stack del frontend y su imagen

Vue 3 + TypeScript + Vite + pnpm, con Composition API (`<script setup lang="ts">`). El `frontend/` se generó con el **template oficial** de Vue (`pnpm create vue@latest`, create-vue) en modo TypeScript y sin features extra, y se le agregó el proxy de dev de `/v1`. Trae la combinación de versiones que mantiene el propio template: TypeScript ~6, Vite 8, `@vitejs/plugin-vue` 6, `vue-tsc` 3 y el `@vue/tsconfig` compartido (más el favicon en `public/`). Scripts oficiales: `dev`, `build` (`type-check` + `build-only`), `preview` y `type-check` (`vue-tsc --build`). Node 24 y `nginx:1.27-alpine` en la imagen.

`frontend/Dockerfile` multi-stage:

```
FROM node:24-alpine AS build
  corepack enable && pnpm install --frozen-lockfile && pnpm build   ->  /app/dist
FROM nginx:1.27-alpine
  COPY --from=build /app/dist  /usr/share/nginx/html
  COPY nginx.conf              /etc/nginx/conf.d/default.conf
```

- `frontend/.dockerignore` excluye `node_modules/`, `dist/`, `.env*`, `.git/`; sin él, el contexto de build incluiría `node_modules` y el build sería lentísimo.
- Los gates del frontend son `type-check` (`vue-tsc --build`) + `build` (`vite build`). Lint se agrega cuando se decida la herramienta; el template oficial no lo incluye en esta configuración.
- Razón: el navegador consume archivos estáticos; nginx los sirve y además hace de borde. El `dist/` se congela en tiempo de build, así que el contenedor `web` es a la vez "el front" y "el nginx".
- Alternativa descartada: contenedor de build separado del de runtime y un tercer contenedor nginx (más piezas sin cambiar el resultado).

### D3 - Borde único: nginx sirve el SPA y proxea `/v1`

`frontend/nginx.conf` (principio):

```
server {
  listen 80;
  root /usr/share/nginx/html;

  location /v1/ {
    proxy_pass http://api:8080;                 # SIN barra final: conserva /v1/...
    proxy_set_header Host              $host;
    proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
  }
  # futuro: location /webhooks/ { proxy_pass http://api:8080; }   # firma del proveedor, no Bearer
  # futuro: location /ws { proxy_pass http://api:8080;
  #           proxy_http_version 1.1;
  #           proxy_set_header Upgrade $http_upgrade;
  #           proxy_set_header Connection "upgrade";
  #           proxy_read_timeout 3600s; }

  location / { try_files $uri $uri/ /index.html; }
}
```

- **`proxy_pass` sin barra final** es lo que conserva el prefijo: `location /v1/` + `proxy_pass http://api:8080;` reenvía `/v1/conversations` tal cual. Con `http://api:8080/;` (barra final) nginx **reemplaza** la parte que hizo match y el `api` recibiría `/conversations` → 404.
- Los headers reenviados son porque hay un proxy delante (Cloudflare): sin ellos el `api` no ve la IP ni el esquema reales.
- Razón: mismo origen → sin CORS; el `api` no publica puerto. Los webhooks y el WebSocket, cuando existan, entran por `location` nuevas al mismo `api` interno, sin exponerlo.
- Alternativa descartada: SPA con URL de API absoluta y CORS (más configuración, distinto origen, no aporta aquí).

### D4 - CI: adaptar el workflow existente y agregar el del front

Se reemplaza `.github/workflows/ci.yml` por dos workflows con filtro de `paths`. Se eligen **dos workflows separados** (no `dorny/paths-filter`): la independencia queda declarada en el `on.paths`, cada uno tiene su propio check/badge y no comparten setup relevante; el costo de duplicar `checkout`/setup es trivial.

```
.github/workflows/backend.yml
  on:
    push: { branches: [main], paths: ['backend/**', '.github/workflows/backend.yml'] }
    pull_request: { branches: [main], paths: ['backend/**', '.github/workflows/backend.yml'] }

.github/workflows/frontend.yml
  on:
    push: { branches: [main], paths: ['frontend/**', '.github/workflows/frontend.yml'] }
    pull_request: { branches: [main], paths: ['frontend/**', '.github/workflows/frontend.yml'] }
```

- `backend.yml`: es el `ci.yml` actual adaptado — `defaults.run.working-directory: backend`, `go-version-file: backend/go.mod`, y los mismos gates (`go mod verify`, `gofmt -l .`, `go vet ./...`, `go build ./...`, `go test -race -count=1 ./...`); publica `ghcr.io/jdgonzalez907/1channel/api:<sha>` + `:latest`.
- `frontend.yml` (nuevo): `pnpm/action-setup` (leo `packageManager`) + `actions/setup-node` con cache de pnpm, `pnpm install --frozen-lockfile`, `pnpm type-check` y `pnpm build`; publica `ghcr.io/jdgonzalez907/1channel/web:<sha>` + `:latest`.
- La publicación de imagen sigue el patrón actual: solo en push a `main` (`if: github.event_name == 'push' && github.ref == 'refs/heads/main'`).
- Cada componente mantiene su propio `:latest` como puntero a su último build; el sha es la identidad inmutable de rollback.
- Un commit que solo toca `frontend/**` no dispara `backend.yml`, así que `api:<sha-nuevo>` **no existe**: el despliegue sigue usando el último sha de `api`.

### D5 - Compose de producción sin base de datos, sin TLS, con healthchecks

`docker-compose.prod.yml` (raíz). Las versiones se inyectan por entorno al ejecutar compose (`API_TAG`, `WEB_TAG`), de modo que subir un componente es cambiar una variable y recrear solo ese servicio:

```
services:
  web:
    image: ghcr.io/jdgonzalez907/1channel/web:${WEB_TAG:-latest}
    ports: ["${WEB_PORT:-80}:80"]           # HTTP; TLS lo aporta Cloudflare Tunnel
    depends_on:
      api: { condition: service_healthy }
    healthcheck:
      test: ["CMD", "wget", "-q", "-O", "/dev/null", "http://localhost/"]
      interval: 10s
      timeout: 3s
      retries: 5
      start_period: 5s
    restart: unless-stopped
  api:
    image: ghcr.io/jdgonzalez907/1channel/api:${API_TAG:-latest}
    expose: ["8080"]                        # interno; sin `ports`
    env_file: [.env]                        # `.env` del VPS, no versionado
    healthcheck:
      test: ["CMD", "wget", "-q", "-O", "/dev/null", "http://localhost:8080/healthz"]
      interval: 10s
      timeout: 3s
      retries: 5
      start_period: 10s
    restart: unless-stopped
```

- **Config por envs inyectadas al ejecutar compose**: el `.env` vive en el VPS junto al compose (mismo patrón que hoy en dev) y aporta `POSTGRES_URL` (o `POSTGRES_*`) más `API_TAG`/`WEB_TAG`/`WEB_PORT` si se quieren override. Nada de secretos en el repo.
- El despliegue independiente: `API_TAG=<sha> docker compose -f docker-compose.prod.yml up -d --no-deps api` (o `web`).
- **Sin TLS en los contenedores**: `web` escucha HTTP en `:80`; Cloudflare Tunnel termina HTTPS. Si `cloudflared` corre en la misma red Docker, apunta a `http://web:80` y no se publica puerto; si corre en el host, `WEB_PORT` permite atar a `127.0.0.1:80`.
- **Healthchecks** con `wget -q -O /dev/null` (busybox de Alpine, sin depender de `curl`). El `api` se chequea contra `GET /healthz` real (cmd/api/router.go), no un puerto abierto a ciegas; `depends_on: service_healthy` evita el 502 de nginx al arrancar.
- No hay `postgres` ni `migrate` en este compose: la base es externa o compartida del VPS.

### D6 - Flujo de desarrollo

- **Un solo `Makefile` en la raíz** como interfaz única, con esquema `verbo` (ambos), `verbo-api` (backend) y `verbo-web` (frontend): `run`/`run-api`/`run-web`, `build`/`build-api`/`build-web`, `check`/`check-api`/`check-web`, más `install-web` y los de datos/prod. Delega en `backend/docker-compose.yml` con `-f ... --project-directory backend` y en `pnpm --dir frontend`.
- Backend: `make up` levanta `postgres`; `make migrate-up` aplica migraciones; `make run-api` inyecta `backend/.env` y corre `go -C backend run ./cmd/api`.
- Frontend: `make run-web` corre `pnpm --dir frontend dev` (Vite) con `server.proxy` reenviando `/v1` a `http://localhost:8080`, para reproducir el mismo origen en desarrollo.
- Alternativa descartada: un Makefile por proyecto (se duplican y se desincronizan).

### D7 - Migraciones manuales contra la base externa

En **dev** siguen por `make migrate-up` (sin cambios). En **prod** se ejecutan a mano, fuera del ciclo de despliegue. El VPS no tiene el binario `migrate`, así que la forma práctica es:

```
docker run --rm -v "$PWD/backend/db/migrations:/migrations" migrate/migrate \
  -path=/migrations -database "$POSTGRES_URL" up
```

- Consecuencia a documentar: **migrar antes de desplegar** un `api` que dependa del nuevo esquema; migración y despliegue van desacoplados.

### D8 - Compatibilidad del contrato bajo `/v1`

El cliente consume `/v1`. Un cambio incompatible se publica bajo un prefijo nuevo y `/v1` se mantiene mientras haya un cliente desplegado que lo use. Esto es lo que permite que `web` y `api` corran shas distintos sin romperse.

### D9 - Configuración y artefactos por proyecto

- `backend/.env` + `backend/.env.example` (dev; no versionado el primero), `backend/.dockerignore` y `backend/.gitignore`.
- `frontend/.env` + `frontend/.env.example` (destinado a futuras variables `VITE_*`; hoy puede estar casi vacío), `frontend/.dockerignore` y `frontend/.gitignore`.
- `.gitignore` raíz mínimo para lo global (editor/SO/cobertura).
- Los `.env` de producción viven en el VPS, fuera del repo, y se inyectan al ejecutar compose.

## Risks / Trade-offs

- [Mover todo el backend rompe rutas y comandos conocidos] → refactor mecánico; los imports no cambian por ser módulo-relativos. `AGENTS.md` y el `Makefile` raíz se actualizan en el mismo change.
- [`paths` + required status checks: un PR que no toca un componente no ejecuta ese workflow y el check requerido queda "esperando"] → usar checks requeridos por workflow acorde a los paths, o un job "gate" agregador.
- [El `api` arranca más lento que nginx] → `healthcheck` + `depends_on: service_healthy` en D5.
- [Sin TLS en el contenedor, alguien podría pegarle directo al `:80` del VPS si se publica en `0.0.0.0`] → `WEB_PORT=127.0.0.1:80` cuando `cloudflared` corre en el host; o no publicar si es contenedor.
- [Desfase esquema/deploy por migraciones manuales] → orden documentado: migrar primero, desplegar después.
- [Sin tests de integración del borde (nginx + SPA)] → verificación manual por HTTP (`GET /`, `GET /ruta`, `GET /v1/...`); CI no levanta contenedores.

## Migration Plan

1. Crear `backend/` y mover ahí el módulo, `db/`, `docs/`, `Dockerfile`, `.dockerignore`, `docker-compose.yml`, `.env` y `.env.example`; crear `backend/.gitignore` y reducir el `.gitignore` raíz.
2. Verificar el backend desde `backend/` (`go build ./...`, `go test ./...`, `sqlc generate` sin diff, `make up && make migrate-up`).
3. Crear `frontend/` con el shell Vue+TS+Vite+pnpm, `Dockerfile`, `.dockerignore`, `nginx.conf`, `.env.example` y `.gitignore`.
4. Adaptar `.github/workflows/ci.yml` a `backend.yml` y crear `frontend.yml`, con filtros de `paths` e imágenes `api`/`web`.
5. Reescribir el `Makefile` raíz (interfaz única) y añadir `docker-compose.prod.yml`.
6. Actualizar `AGENTS.md` a la nueva estructura y comandos.
7. Rollback: revertir el change; al no haber cambios de esquema ni de API, basta con volver a la estructura de raíz y al workflow único.

## Open Questions

- Ninguna pendiente. Versiones fijadas en D2; nombres de imagen `.../1channel/api` y `/web`; un solo Makefile en la raíz.
