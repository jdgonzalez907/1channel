# Proposal

## Why

Una conversación que expira sin haber sido asignada a ningún agente queda invisible para todos: no aparece en `GET /v1/conversations?status=finished` y `GET /v1/conversations/{id}` responde 403. El paso `pending` (visible para todos) -> `expired` sin agente (invisible) pierde el historial del inbox. En dev esto es visible porque el seed crea 10 conversaciones `expired` sin agente (`conv 21..30`), que nunca se listan.

## What Changes

- Se generaliza la regla de visibilidad de conversaciones: una conversación SHALL ser visible para un agente cuando **no tiene agente asignado** (`user_id IS NULL`) o cuando su agente asignado es el solicitante.
- Se actualiza `ListConversationsForAgent` (`db/queries/conversations/conversations.sql`) para usar `(c.user_id IS NULL OR c.user_id = :agent_id)` en lugar de `(c.status = 'pending' OR c.user_id = :agent_id)`. Requiere `sqlc generate`.
- Se actualiza `isConversationVisible` en `conversation_read_handler.go` para permitir el acceso cuando la conversación no tiene agente asignado, de modo que el detalle sea consistente con la bandeja (hoy daría 403 a una fila que sí aparece).
- Consecuencia: una `expired` sin agente pasa a listarse en `finished` y a poder abrirse por cualquier agente. `assigned`, `resolved` y `expired` con agente siguen siendo privadas de su agente.
- No cambian los estados, ni la creación/expiración, ni el read model denormalizado.

## Capabilities

### New Capabilities

_(ninguna)_

### Modified Capabilities

- `http-read-api`: cambia la regla de visibilidad de la bandeja y del detalle de conversación, de "pending o asignada al solicitante" a "sin agente asignado o asignada al solicitante".

## Impact

- `backend/db/queries/conversations/conversations.sql` — `ListConversationsForAgent` (WHERE de visibilidad) y `sqlc generate`.
- `backend/internal/modules/conversations/infra/http/conversation_read_handler.go` — `isConversationVisible`.
- `backend/internal/modules/conversations/infra/http/conversation_read_handler_test.go` — el caso "finished without agent is not visible" pasa a visible; se agregan casos.
- `openspec/specs/http-read-api/spec.md` — se sincroniza al archivar.
- El seed (`dev_seed.sql`) no cambia; sus `expired` sin agente pasan a ser visibles.
