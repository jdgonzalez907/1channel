# Design

## Context

Ver `proposal.md - Why`. El módulo de conversaciones ya tiene los use cases `ReceiveContactMessage`, `ReceiveContactMessageEdit` y `ReceiveContactMessageDelete` en `app/`, con mocks y tests. Están declarados en `ConversationsAPI` (`api.go`) pero no se construyen ni se exponen en HTTP. El patrón de escritura existente (`ConversationWriteHandler` en `infra/http`) depende de interfaces de `app`, mapea errores de dominio en `writeError` y se registra dentro del grupo autenticado en `cmd/api/router.go`.

El aggregate ya expone `Message.MarkAsRead` y `AgentReadConversation`, que marca los mensajes del contacto; el read receipt es su operación simétrica. La persistencia ya soporta el cambio: `Save` -> `UpsertMessage` escribe `status` y `read_at`, y `FindWithMessageByExternalID` reconstruye la conversación con el mensaje, por lo que no se requieren migraciones ni queries nuevas.

Restricciones relevantes:
- Son operaciones de **escritura**: el handler depende de `app`, nunca de SQLC, y el contrato vive en la capability `http-api`.
- La ruta es un canal de prueba, abierto y sin configuración: no hay auth, env ni feature flags.
- Los eventos de canal son **asíncronos**: no controlamos al usuario ni las peticiones, así que se ejecutan aunque la conversación esté finalizada.
- Los use cases existentes ya son idempotentes ante duplicados y borrados repetidos.

## Goals / Non-Goals

**Goals:**
- Exponer `POST /v1/webhooks/1channel` que mapea cuatro eventos a los casos de uso de contacto.
- Agregar la operación de acuse de lectura del contacto sobre mensajes del agente.
- Un único DTO compartido con `event`, `external_contact_id` y `external_message_id` obligatorios y `text` según el evento; validación por evento.
- Reutilizar el mapeo de errores de dominio del handler de escritura.
- Dejar el árbol de rutas preparado para `/v1/webhooks/<plataforma>` futuros.

**Non-Goals:**
- No se implementa ningún adaptador de plataforma real (meta, telegram) ni la verificación de firmas.
- No se alteran las reglas de `unread_count` (cuenta solo mensajes del contacto).
- No hay autenticación, autorización, configuración de entorno ni gating por ambiente.
- No se define idempotencia nueva más allá de la no-retroactividad de la lectura.

## Decisions

### 1. Un solo path con envelope de evento

`POST /v1/webhooks/1channel` recibe un envelope con `event`. Se prefiere a rutas REST por operación porque un adaptador de canal real emite eventos y permite reutilizar un único handler; además, `/1channel` es el hermano de prueba de los futuros `/meta` y `/telegram`, que resolverán los mismos use cases. Alternativa considerada: una ruta por operación; descartada por fragmentar el registro y el manejo de errores.

### 2. Un único DTO con punteros, sin timestamp

`ContactWebhookRequest` tiene `Event string`, `ExternalContactID string` y `ExternalMessageID string` (obligatorios en los cuatro eventos) y `Text *string` (solo `message.received`/`message.edited`). El timestamp no forma parte del DTO: lo genera el servidor con `time.Now().UTC()` al construir el input del use case. Alternativa considerada: DTO por evento; descartada por duplicación innecesaria.

### 3. Validación por evento

El handler resuelve todo con un único `switch` sobre `event`: cada caso valida sus campos requeridos y delega en el use case. Falta un campo requerido -> 422. Un `event` no soportado se ignora y responde 200 sin alterar nada. JSON malformado -> 400 (el decode falla antes del switch). Requeridos por evento:

| event | external_contact_id | external_message_id | text |
|---|---|---|---|
| `message.received` | requerido | requerido | requerido |
| `message.edited` | requerido | requerido | requerido |
| `message.deleted` | requerido | requerido | — |
| `message.read` | requerido | requerido | — |

### 4. Reutilizar el mapeo de errores de dominio

Se extrae `writeError` de `ConversationWriteHandler` a una función de paquete en `conversations/infra/http` (p. ej. `writeConversationError`) y la usan ambos handlers. Se descarta moverlo a `shared/infra` porque el mapeo es específico del dominio de conversaciones; se descarta duplicarlo para no divergir.

### 5. Read receipt espejo de la edición del contacto

`Conversation.ReceiveContactMessageRead(contactID, externalID, at)` sigue el patrón de `ReceiveContactMessageEdit`: valida que el contacto pertenezca a la conversación, localiza el mensaje por `externalMsgIdx` y valida que el mensaje sea del agente. No llama a `ensureNotFinished` ni a `ensureAcceptsMessages`: al ser un evento asíncrono, se aplica aunque la conversación esté finalizada. El use case `ReceiveContactMessageRead` es espejo de `receive_contact_message_delete.go` (`GetOrCreateContactIDByExternalID` -> `FindWithMessageByExternalID` -> método del aggregate -> `Save`).

`Message.MarkAsRead` pasa a ser no-retrocedente, igual que `EditText`: si ya hay una lectura posterior o igual, no la modifica. Además no reactiva estados terminales: sobre un mensaje `deleted` registra la lectura pero conserva el estado eliminado, y sobre un mensaje `failed` no hace nada. Es seguro porque `AgentReadConversation` solo lo invoca cuando `read_at` es nil.

### 6. Handler y DTO en el módulo de conversaciones

Nuevos archivos `contact_webhook_handler.go` y `contact_webhook_dto.go` en `conversations/infra/http`, siguiendo la convención de handlers del repo. `ContactWebhookHandler` depende de las interfaces de `app`. Se descarta colgarlo de `ConversationWriteHandler` para separar el borde autenticado del abierto.

### 7. Registro fuera del grupo autenticado

En `router.go`, el webhook se registra en `/v1` fuera del `r.Group` que aplica `Auth`. En `app.go`, `newConversationsModule` construye además los cuatro use cases de contacto y devuelve también `*convhttp.ContactWebhookHandler`.

## Risks / Trade-offs

- **Endpoint de escritura abierto** -> Es intencional para pruebas. Mitigación: no exponerlo en producción, o eliminarlo/ocultarlo al integrar los adaptadores reales; se documenta como ruta de prueba.
- **Cambio en `Message.MarkAsRead`** -> Toca un método compartido. Mitigación: el único uso existente (`AgentReadConversation`) ya filtra `read_at` nil; el cambio además corrige que un mensaje `deleted` ya no pase a `read`. Se cubre con tests.
- **Mapeo de errores compartido** -> Extraer `writeError` toca código existente; el cambio debe preservar el comportamiento actual y sus tests.
- **Crecimiento a plataformas** -> El envelope es propio de `/1channel`; `/meta` y `/telegram` tendrán sus propios DTOs y traducirán a los mismos use cases, sin compartir el DTO de prueba.
