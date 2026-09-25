# Proposal

## Why

El aggregate `Conversation` no puede finalizar una conversación en runtime: `finishedAt` solo se asigna por constructor, así que las reglas de "conversación finalizada" (`ensureAcceptsMessages` y las validaciones de agente) son inalcanzables en operación normal. Además, `Messages()` itera un `map`, por lo que el orden de los mensajes es aleatorio y no refleja la conversación real.

## What Changes

- **Nuevo metodo `ExpireConversation` en Conversation**: Cierra una conversación por expiración (evento de sistema, sin agente). Puede expirar una conversación pending o assigned.
- **Nuevo metodo `ResolveConversation` en Conversation**: Cierra una conversación porque el agente la resolvió. Valida que el agente sea el asignado.
- **Nuevo helper `ensureNotFinished` en Conversation**: Guard único que concentra la condición de conversación finalizada y devuelve `ErrConversationFinished`. Reemplaza al wrapper `ensureAgentCanModify`.
- **Reutiliza `ErrConversationFinished`**: No se agregan errores nuevos; modificar o finalizar una conversación terminada comparten el mismo error.
- **Orden de mensajes en `Messages()`**: Los mensajes se devuelven ordenados por `sentAt` ascendente (timeline), con desempate por `id`.

## Capabilities

### Modified Capabilities

- `conversation`: Agregar requerimientos para expirar y resolver una conversación, y modificar "Obtener mensajes de conversacion" para devolver los mensajes ordenados por `sentAt`.

### New Capabilities

Ninguna - el comportamiento se agrega al capability existente de conversation.

## Impact

- `internal/modules/conversations/domain/conversation.go` - Nuevos métodos `ExpireConversation`, `ResolveConversation`, helper `ensureNotFinished`; `Messages()` ordenado
- `internal/modules/conversations/domain/conversation_test.go` - Tests nuevos y actualizados