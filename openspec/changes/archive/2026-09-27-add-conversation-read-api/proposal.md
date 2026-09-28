# Proposal

## Why

Hoy el contrato HTTP solo permite escribir (crear usuarios, enviar mensajes, marcar leídos, resolver). Un agente no puede ver la bandeja de conversaciones ni abrir una conversación, así que la UI no puede construirse. Se necesita un contrato REST de lectura que alimente la barra lateral de conversaciones y el detalle del chat.

## What Changes

- Se agrega un contrato REST de **solo lectura** para agentes autenticados:
  - `GET /v1/conversations` — barra lateral paginada (scroll infinito de 20 en 20), visible para `status = pending` o `user_id = agente`. El parámetro `status` es obligatorio y solo acepta `open` o `finished`; opcionalmente filtra por `external_contact_id` (404 si el contacto no existe). Ordena por el último mensaje (`sent_at DESC, id DESC`). Cada fila trae `contact.id`, `contact.external_id`, `unread_count` y preview del último mensaje.
  - `GET /v1/conversations/{id}` — información de la conversación más los últimos 20 mensajes en `sentAt ASC`, paginando hacia atrás por posición (`before_sent_at` / `before_id`) hasta agotar el historial. Accesible solo si la conversación es `pending` o del agente; en otro caso 403.
  - `GET /v1/contacts/{id}` — contacto por identificador.
  - `GET /v1/users/{id}` — usuario/agente por identificador.
- **Decisión de arquitectura**: las lecturas NO pasan por `app` ni `domain` ni repositorios de escritura. El handler HTTP llama directamente a queries de solo lectura (SQLC) y mapea a DTOs. Así el core permanece limpio de DTOs y los caminos de lectura y escritura quedan separados.
- **Read model denormalizado**: para que la bandeja pagine por índice (O(página)) en vez de ordenar todo el historial del agente, `conversations` guarda `last_message_at`, `last_message_id` y `unread_count`. Se recomputan al final de la misma transacción que escribe los mensajes. Es el único punto que toca el write side; `domain` y `app` no cambian.
- Se modifican las restricciones de persistencia: `conversations` incorpora el modelo de lectura denormalizado y se agregan índices compuestos sobre `last_message_at` por agente y por estado, más un índice por `contact_id`. Los índices para ordenar los mensajes de una conversación por `sent_at` pasan a ser **requeridos**. El índice parcial de no leídos sigue prohibido.
- Los mensajes `deleted` se incluyen en el detalle con `text = null` y los campos asociados nulos.

## Capabilities

### New Capabilities

- `http-read-api`: contrato REST de lectura para que un agente autenticado consulte la bandeja de conversaciones, abra una conversación con su historial y consulte contactos y usuarios por identificador.

### Modified Capabilities

- `persistence`: se agrega el modelo de lectura denormalizado de conversación (`last_message_at`, `last_message_id`, `unread_count`) mantenido en la transacción de escritura, con sus invariantes e índices. Se modifica el requirement `Acceso por las búsquedas del MVP`: los índices por agente/estado sobre `last_message_at`, por `contact_id` y de orden de mensajes por `sent_at` pasan a ser requeridos; el índice parcial de no leídos sigue prohibido.

## Impact

- **Nuevo**: paquete de lectura HTTP por módulo (`conversations/infra/http`, `contacts/infra/http`, `users/infra/http`), queries SQLC de solo lectura y DTOs de respuesta.
- **Queries/migraciones**: nuevas `.sql` en `db/queries/`; se editan las migraciones existentes para agregar columnas e índices. `sqlc generate`.
- **Wiring**: `cmd/api/app.go` y `cmd/api/router.go` deben registrar los handlers de lectura bajo el middleware de auth.
- **Write side**: `conversations` suma `last_message_at`, `last_message_id` y `unread_count`; el repositorio de conversaciones los recomputa al final de `Save` dentro de la transacción. **Sin cambios** en `domain` ni `app`.
- **Specs afectados**: `persistence` (modelo de lectura + índices). `http-api` no cambia (sigue siendo el contrato de escritura).
