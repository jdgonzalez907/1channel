# Proposal

## Why

Hoy `frontend/` es solo el shell de Vue: un `App.vue` placeholder sin forma de operar la bandeja. El backend ya expone todos los casos de uso del agente por REST (leer bandeja y detalle, responder, marcar leído, resolver), pero no existe una superficie web para usarlos. Este cambio entrega la pantalla mínima del agente: dos contenedores, uno que lista conversaciones y otro que muestra y opera su detalle, sin dependencias nuevas.

## What Changes

- **Nueva consola del agente** (SPA, mismo origen `/v1`), compuesta por dos contenedores:
  - **Lista de conversaciones**: campo de texto para el `id` del agente (token Bearer), radio `open`/`finished`, filtro opcional por `external_contact_id` con botones **Buscar** y **Limpiar**, y un botón **Cargar más** que usa el cursor de paginación (`before_sent_at`/`before_id`). Cada item muestra en negrita el `external_contact_id`, la vista previa del último mensaje con su dueño (`agent`/`contact`), el estado, la fecha del último mensaje y la cantidad de no leídos solo si es mayor a 0.
  - **Detalle de la conversación**: título con el `external_contact_id` y el estado, botón **Resolver** solo cuando está `assigned`, campos de la conversación y su historial de mensajes. El historial tiene **Cargar más** para los mensajes más antiguos. El campo de texto para responder aparece solo si la conversación está abierta (`pending`/`assigned`).
- **Identidad del agente**: el `id` se guarda en `localStorage`, viaja como `Authorization: Bearer` y se valida al entrar con `GET /v1/users/{id}`.
- **Marcar leído**: al abrir el detalle de una conversación asignada al agente (`agent_id` igual al suyo) con `unread_count > 0`, se marca como leída (`PATCH .../messages`) y se refresca la lista; aplica también a finalizadas asignadas (`resolved`/`expired`). En conversaciones sin agente no se puede (el dominio lo rechaza) y no se dispara.
- **Simulador de webhook** (en el mismo change): un modal que hace `POST /v1/webhooks/1channel` (público) para inyectar eventos del contacto (`message.received`, `message.edited`, `message.deleted`, `message.read`) y poder probar la bandeja sin un canal real.
- **Estilo**: HTML/CSS básico (sin framework de UI), con colores en los estados y en el contador de no leídos (verde salvo en `expired`, que va gris), mensajes alineados según su dueño (agente a la derecha) y scroll interno en la lista y en el historial; estados vacíos, de carga y de error legibles.
- **Marcas de mensajes**: un mensaje leído (`read_at`) muestra **leído** y uno editado (`edited_at`) muestra **editado**, con colores distintos; un mensaje eliminado no muestra marcas.
- **Desplazamiento del historial**: al abrir una conversación el historial se posiciona en el último mensaje; al cargar mensajes más antiguos conserva la posición de lectura.
- **Encabezado del detalle**: además del `external_id` del contacto y el estado, se muestra el id de la conversación (ayuda para probar el simulador).
- **Sin dependencias nuevas**: solo `vue`; `fetch` nativo, sin router, sin store y sin cliente HTTP.

## Capabilities

### New Capabilities

- `agent-console`: la pantalla del agente que lista conversaciones, abre su detalle, responde, marca leído y resuelve, incluyendo la identidad por agente, los filtros y la paginación por cursor.
- `webhook-simulator`: el modal de desarrollo que simula eventos del contacto contra `POST /v1/webhooks/1channel` para generar y modificar conversaciones desde la misma pantalla.

### Modified Capabilities

- Ninguna. No cambian requisitos de dominio ni del contrato HTTP existente; la consola solo consume `/v1`.

## Impact

- **Código**: únicamente `frontend/src/` (y `frontend/index.html`/estilos si hiciera falta). Nuevos: cliente de API tipado, tipos de las respuestas, componentes de lista, detalle y modal.
- **Backend/DB/API**: sin cambios. No se toca `backend/`, ni SQLC, ni migraciones.
- **CI**: siguen aplicando los gates del frontend (`pnpm type-check` y `pnpm build`); no se agrega framework de tests.
- **No-goals**: login/sesión real, tiempo real/WebSocket, un sistema de tema configurable, enrutador, store, cobertura de tests del front, y exponer el `external_id` de los mensajes (la API de lectura no lo trae; el simulador lo pide a mano).
