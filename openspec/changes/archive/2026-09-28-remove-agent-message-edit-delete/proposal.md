# Proposal

## Why

No queremos que un agente pueda **editar ni borrar** mensajes: el control del inbox debe quedar sin que el agente reescriba lo enviado. Hoy el dominio todavía expone `Conversation.AgentEditMessage` y `AgentDeleteMessage` (código muerto: sin use case ni endpoint) y la spec de `conversation` lista escenarios de edición/borrado por agente, lo que contradice esa decisión.

## What Changes

- **BREAKING** (interno): se eliminan del agregado `Conversation` los métodos `AgentEditMessage` y `AgentDeleteMessage`, y sus tests.
- Se conservan `Message.EditText` y `Message.Delete` (los usa el flujo del **contacto**).
- En `conversation` se **eliminan** los requirements `Editar mensaje de texto` y `Eliminar mensaje` (que arrastraban los escenarios de agente) y se **modifican** `Caso de uso: editar mensaje de contacto` y `Caso de uso: eliminar mensaje de contacto` para concentrar las reglas de edición/eliminación y su validación, dejando explícito que **el agente no puede** editar ni eliminar.
- No hay cambios de contrato HTTP: nunca existió endpoint de edición/borrado de agente, así que `http-api` no cambia.
- No se tocan el read model, ni `unread_count`, ni la lógica de lectura.

## Capabilities

### New Capabilities

_(ninguna)_

### Modified Capabilities

- `conversation`: se eliminan los requirements generales `Editar mensaje de texto` y `Eliminar mensaje` (solo aplicaban al agente en su parte de edición/borrado) y se modifican los casos de uso del contacto para contener las reglas de validación; se explicita que el agente no edita ni elimina.

## Impact

- `backend/internal/modules/conversations/domain/conversation.go` — eliminar `AgentEditMessage` y `AgentDeleteMessage`.
- `backend/internal/modules/conversations/domain/conversation_test.go` — eliminar `TestConversation_AgentEditMessage` y `TestConversation_AgentDeleteMessage`.
- `openspec/specs/conversation/spec.md` — se sincroniza al archivar.
- Sin cambios en `app`, `infra/http`, `api.go`, queries ni migraciones.
