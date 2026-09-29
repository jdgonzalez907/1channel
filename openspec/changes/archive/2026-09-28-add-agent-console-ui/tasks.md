# Tasks

## 1. Fundamentos del cliente de API

- [x] 1.1 Crear `frontend/src/api/types.ts` con los tipos de entrada/salida de `/v1` (item y respuesta de lista con `next_before_*`, detalle con `contact`/`agent_id`/`unread_count`/`messages`, mensaje con `owner`/`deleted_at`, `Contact`, `User` y el cuerpo de error `problem+json`); verificar con `pnpm --dir frontend type-check` sin errores.
- [x] 1.2 Crear `frontend/src/api/client.ts` con `fetch` nativo, cabecera `Authorization: Bearer <agentId>`, y conversión de cualquier respuesta no-2xx en un `ApiError` (`status`, `title`, `detail` leídos de `application/problem+json` con fallback); incluir `getUser`, `listConversations`, `getConversation`, `sendMessage`, `markRead`, `resolveConversation` y `simulateWebhook`. Verificar con `pnpm --dir frontend type-check` y una llamada manual real contra `make run-api` desde la consola del navegador.

## 2. Sesión del agente y layout

- [x] 2.1 Implementar la barra del agente en `App.vue`: campo de texto para el id, persistencia en `localStorage`, restauración al recargar y validación con `GET /v1/users/{id}`; al fallar (UUID inválido o 404) mostrar el error y no listar. Verificar manualmente con un id válido y con uno inválido, y recargando la página.
- [x] 2.2 Montar el layout de dos paneles (lista a la izquierda, detalle a la derecha) con CSS plano sin colores, apilado en pantallas angostas, y conectar `App.vue` con los hijos (props de `agentId`/`selectedId`, eventos `select` y `changed`). Verificar visualmente en el navegador a ancho completo y angosto.
- [x] 2.3 Documentar en `frontend/README.md` cómo levantar y usar la consola (`make run-web`, id del agente, requisitos de API); verificar que los comandos documentados existan y corran tal cual.

## 3. Lista de conversaciones

- [x] 3.1 Implementar `ConversationList.vue` con el radio `open`/`finished` (default `open`) y la consulta por estado; verificar en el navegador que cambiar de estado recarga la lista.
- [x] 3.2 Agregar el filtro opcional por `external_contact_id` con **Buscar** y **Limpiar**; verificar con un contacto existente (filtra), uno inexistente (muestra el error sin listado previo) y con Limpiar (vuelve al listado completo).
- [x] 3.3 Implementar el item de la lista (dentro de `ConversationList.vue`) con `external_contact_id` en negrita, preview del último mensaje con su dueño, estado y fecha del último mensaje, y `unread_count` solo si es mayor a 0; verificar con conversaciones del seed con y sin no leídos.
- [x] 3.4 Agregar la paginación por cursor: guardar `next_before_sent_at`/`next_before_id` y mostrar **Cargar más** solo cuando ambos no sean nulos, anexando la página siguiente; verificar que aparezcan más de 20 items y que el botón desaparezca al agotarse.
- [x] 3.5 Añadir estados de carga, lista vacía y error en la lista; verificar cortando la API (error), filtrando un contacto inexistente y con un listado vacío.

## 4. Detalle de la conversación

- [x] 4.1 Implementar `ConversationDetail.vue` cargando `GET /v1/conversations/{id}` y mostrando en el encabezado el `external_contact_id` y el estado; verificar seleccionando items y que un 403/404 muestre error sin detalle previo.
- [x] 4.2 Implementar el historial de mensajes (dentro de `ConversationDetail.vue`) con dueño, texto y fecha, marcando los `deleted` como eliminados sin texto, más **Cargar más** para mensajes antiguos; verificar con una conversación del seed con más de 20 mensajes y con mensajes eliminados (via simulador).
- [x] 4.3 Agregar el campo de respuesta visible solo en `pending`/`assigned` y el envío con `POST .../messages`; verificar responder una `pending` (pasa a `assigned` y se refleja), responder una `assigned`, y que en `resolved`/`expired` no haya campo.
- [x] 4.4 Agregar el botón **Resolver** visible solo en `assigned`, con `PATCH /v1/conversations/{id}`; verificar resolver una asignada (204, estado y lista se actualizan) y que no aparezca en `pending` ni en finalizadas.
- [x] 4.5 Marcar como leída automáticamente una conversación `assigned` con `unread_count > 0` al abrirla (`PATCH .../messages`) y refrescar la lista; verificar que el contador baja, que en una `pending` no se intenta y que abrir una `assigned` sin no leídos no dispara la llamada.

## 5. Simulador de webhook

- [x] 5.1 Implementar `WebhookSimulator.vue` como modal abierto desde el panel de la lista, con select de `event` (default `message.received`), campos `external_contact_id`, `external_message_id` y `text` (visible solo para `received`/`edited`), y prellenado de `external_contact_id` cuando hay conversación seleccionada; verificar abriéndolo con y sin selección.
- [x] 5.2 Enviar el evento con `simulateWebhook` y, tras 204, refrescar la lista y el detalle del contacto abierto; verificar que `message.received` con un contacto nuevo cree una conversación visible y que con uno existente adjunte el mensaje.
- [x] 5.3 Mostrar los errores del simulador (422 por campo faltante, 404 por mensaje inexistente) con título y detalle y conservar los valores; verificar enviando un evento sin campo requerido y un `message.read` con `external_message_id` inexistente.

## 6. Verificación de integración

- [ ] 6.1 Ejecutar el flujo completo con `make up`, `make migrate-up`, `make seed`, `make run-api` y `make run-web`: entrar con un id de agente del seed, listar `open`, buscar por contacto, abrir, marcar leído, responder una `pending`, resolver una `assigned`, cargar más en ambos paneles y simular un `message.received`; confirmar que el estado de la UI coincide con la API.
- [x] 6.2 Correr los gates del frontend `pnpm --dir frontend type-check` y `pnpm --dir frontend build` y confirmar que terminan sin errores.

## 7. Ajustes de UI aplicados

- [x] 7.1 Colores en los estados y en el contador de no leídos, mensajes del agente a la derecha y del contacto a la izquierda, y scroll interno en la lista y en el historial; verificado con `pnpm --dir frontend build`.
- [x] 7.2 Posicionar el historial en el último mensaje al abrir y conservar la posición al cargar mensajes más antiguos; verificado con `pnpm --dir frontend build`.
- [x] 7.3 Mostrar el id de la conversación como subtítulo bajo el `external_id` del contacto; verificado con `pnpm --dir frontend build`.
- [x] 7.4 Marcar los mensajes leídos y editados, sin marcar los eliminados; verificado con `pnpm --dir frontend build`.
- [x] 7.5 Mostrar el contador de no leídos en gris para las conversaciones `expired` (no se marcan como leídas al abrir); verificado con `pnpm --dir frontend build`.
- [x] 7.6 Mostrar el contador de no leídos en verde para las conversaciones abiertas y alinearlo a la derecha del item; verificado con `pnpm --dir frontend build`.
- [x] 7.7 Marcar como leída al abrir cualquier conversación asignada al agente (incluye `resolved` y `expired` asignadas) usando `agent_id`, y colorear su contador de no leídos en consecuencia; verificado con `pnpm --dir frontend build`.
- [x] 7.8 Actualizar el detalle en sitio al marcar leído (sin refetch) para no perder el foco/scroll, y refrescar el detalle del webhook solo cuando el contacto del evento coincide; verificado con `pnpm --dir frontend build`.
