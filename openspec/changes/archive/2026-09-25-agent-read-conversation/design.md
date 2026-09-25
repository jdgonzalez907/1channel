# Design

## Context

El aggregate `Conversation` actualmente no tiene forma de que el agente marque mensajes del contacto como leídos. Además, `NewMessage` no valida la propiedad exclusiva del mensaje (agentID XOR contactID).

Ver proposal.md para la motivacion del cambio.

## Goals / Non-Goals

**Goals:**
- Permitir al agente marcar mensajes del contacto como leídos
- Validar que un mensaje tenga exactamente un dueño (agentID XOR contactID)
- No importa si la conversacion está finalizada para leer mensajes

**Non-Goals:**
- Contar mensajes no leídos (para otro momento)
- Notificaciones de mensajes leídos (para otro momento)
- Validación de timestamps (futuro, pasado)

## Decisions

### 1. `MarkAsRead(at time.Time)` en Message

Metodo publico en `Message` que marca un mensaje como leído.

```
Logica:
  readAt = at
  status = MessageStatusRead
```

**Por que no valida el dueño?** La validación de dueño se hace en `Conversation.AgentReadConversation`. El método en Message solo ejecuta la acción.

**Por que publico?** Es una operación de negocio que el aggregate root (Conversation) necesita ejecutar.

### 2. `AgentReadConversation(agentID uuid.UUID, at time.Time) error`

Metodo publico en `Conversation` para que el agente lea la conversación.

```
Flujo:
  1. if agentID != conv.agentID → ErrConversationAgentNotOwner
  2. For each msg in found:
       if msg.contactID != nil && msg.readAt == nil:
         msg.MarkAsRead(at)
         dirty[msg.ID] = msg
  3. updatedAt = at
  4. return nil
```

**Por que no valida si la conversación está finalizada?** El agente puede leer conversaciones finalizadas. Pueden haber llegado mensajes antes de finalizar que no se leyeron.

**Por que filtra por contactID?** El agente solo puede marcar como leídos mensajes del contacto, no los suyos.

### 3. Invariante en `NewMessage`: agentID XOR contactID

Validar que un mensaje tenga exactamente un dueño.

```
if agentID == nil && contactID == nil
  → return nil, ErrMessageInvalidOwner

if agentID != nil && contactID != nil
  → return nil, ErrMessageInvalidOwner
```

**Por que?** Todo mensaje debe pertenecer a alguien. Si es del agente, no puede ser del contacto y viceversa. Esto simplifica la lógica de filtrado en `AgentReadConversation`.

### 4. Errores nuevos

```go
ErrMessageInvalidOwner = errors.New("message must have exactly one owner (agent or contact)")
```

**Por que no reutilizar errores existentes?** La validación de propiedad es diferente a otros errores de mensaje (ID inválido, texto vacío, etc.).

### 5. Dirty tracking

Los mensajes marcados como leídos se agregan al map `dirty` para que la capa de persistencia sepa qué actualizar.

```
msg.MarkAsRead(at)
dirty[msg.ID] = msg
```

## Risks / Trade-offs

**[Riesgo] Mensaje sin dueño en datos existentes**: Si hay mensajes en la base de datos sin agentID ni contactID, el invariante fallará al reconstruir el aggregate.
- **Mitigación**: Validar en migración de datos antes de aplicar el cambio.

**[Trade-off] No contar mensajes no leídos**: No devolvemos el número de mensajes no leídos en `AgentReadConversation`.
- **Razon**: Simplicidad. Se puede agregar despues sin cambiar el diseño.

**[Trade-off] No validar timestamps**: No validamos que `at` sea válido (no futuro, no pasado).
- **Razon**: Consistencia con otros métodos del aggregate.

## Open Questions

Ninguna - todas las decisiones se tomaron durante la exploracion.