# frontend

SPA de 1Channel (Vue 3 + TypeScript + Vite + pnpm + vue-router). Consume el API REST en
`/v1` en el mismo origen: en dev Vite proxea `/v1` a `http://localhost:8080`, en producción
lo hace nginx. Más allá de Vue y vue-router no hay dependencias (`fetch` nativo, sin store).

## Rutas

| Ruta         | Página                          | Acceso   |
| ------------ | ------------------------------- | -------- |
| `/`          | Landing pública (B2B)           | público  |
| `/consola`   | Consola del agente              | id agente |
| `/privacidad`| Política de privacidad          | público  |
| `/terminos`  | Términos y condiciones          | público  |

La landing y las páginas legales comparten un layout público con encabezado y pie; la
consola queda fuera de ese layout. Una ruta desconocida resuelve a la landing. Las URLs
son direccionables: recargar `/privacidad` o `/consola` funciona porque nginx y Vite
resuelven a `index.html` (SPA fallback).

Los textos de **privacidad** y **términos** son un **borrador pendiente de revisión legal**
(ver la nota visible en cada página); no son asesoría jurídica.

## Sitio público

La landing se sirve en `/` sin credenciales. El color de acento del sitio público es
`#7540BF`, definido en `src/assets/theme.css` sobre el contenedor `.public-site` para no
afectar a la consola. El identificador visual de la consola no cambia.

## Consola del agente

La consola se sirve en `/consola` y tiene tres contenedores: la lista de conversaciones
(izquierda), el detalle (centro) y la persona del contacto (derecha), más un modal para
simular eventos del contacto por webhook.

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
