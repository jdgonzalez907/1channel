# frontend

SPA del agente de 1Channel (Vue 3 + TypeScript + Vite + pnpm). Consume el API REST en
`/v1` en el mismo origen: en dev Vite proxea `/v1` a `http://localhost:8080`, en producción
lo hace nginx. Sin dependencias más allá de Vue (`fetch` nativo, sin router ni store).

## Consola del agente

La pantalla tiene dos contenedores: la lista de conversaciones (izquierda) y el detalle
(derecha), más un modal para simular eventos del contacto por webhook.

1. Levantá la API y sus datos (desde la raíz del repo):

   ```sh
   make up          # postgres 18
   make migrate-up
   make seed        # datos de prueba (solo dev; destructivo)
   make run-api     # API en :8080
   ```

2. Levantá el front:

   ```sh
   make run-web     # Vite con proxy /v1 -> :8080
   ```

3. Obtené el id de un agente del seed y pegalo en el campo **Id del agente**:

   ```sh
   sudo docker compose -f backend/docker-compose.yml --project-directory backend \
     exec -T postgres psql -U dev -d 1channel_dev -tAc \
     "select id from users order by created_at, id limit 1;"
   ```

Con un id válido la consola lista la bandeja (`open` por defecto). Desde la lista podés
filtrar por `external_contact_id`, paginar con **Cargar más** y abrir el simulador de
webhook. En el detalle podés responder (solo conversaciones abiertas), resolver (solo
`assigned`) y ver el historial; al abrir una conversación `assigned` con mensajes sin leer
se marcan como leídos.

## Scripts

```sh
pnpm install
pnpm dev          # Vite
pnpm type-check   # vue-tsc --build
pnpm build        # type-check + build de producción
```
