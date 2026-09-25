# Proposal

## Why

El agente necesita marcar mensajes del contacto como leídos cuando lee una conversacion. Actualmente no hay forma de rastrear qué mensajes han sido leídos por el agente. Además, el invariante de propiedad del mensaje (agentID XOR contactID) no está validado en `NewMessage`.

## What Changes

- **Nuevo metodo `AgentReadConversation`**: Permite al agente marcar mensajes del contacto como leídos. Valida que el agente sea el asignado a la conversacion. No importa si la conversacion está finalizada.
- **Nuevo metodo `MarkAsRead` en Message**: Marca un mensaje como leído con timestamp. Cambia status a `read` y asigna `readAt`.
- **Invariante en `NewMessage`**: Validar que un mensaje tenga exactamente un dueño (agentID XOR contactID).
- **Nuevo error `ErrMessageInvalidOwner`**: Para cuando un mensaje no cumple el invariante de propiedad.

## Capabilities

### Modified Capabilities

- `conversation`: Agregar requerimiento para que el agente pueda leer conversaciones y marcar mensajes del contacto como leídos, e invariante de propiedad de mensaje.

### New Capabilities

Ninguna - el comportamiento se agrega al capability existente de conversation.

## Impact

- `internal/modules/conversations/domain/conversation.go` - Nuevo metodo `AgentReadConversation`
- `internal/modules/conversations/domain/message.go` - Nuevo metodo `MarkAsRead`, invariante en `NewMessage`
- `internal/modules/conversations/domain/conversation_test.go` - Tests nuevos
- `internal/modules/conversations/domain/message_test.go` - Tests nuevos