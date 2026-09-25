# Design

## Context

El aggregate `Conversation` actualmente soporta solo mensajes entrantes de contactos via `ReceiveContactMessage`. Los mensajes se almacenan como valores (`Message`) en maps, lo que rompe la identidad de la entidad. No existe validacion de estado de conversacion antes de recibir mensajes.

Ver proposal.md para la motivacion del cambio.

## Goals / Non-Goals

**Goals:**
- Permitir que agentes envien mensajes a conversaciones
- Implementar logica de "claim" para conversaciones pendientes
- Validar que la conversacion acepta mensajes antes de recibir/enviar
- Corregir el patron de entidad usando punteros (*Message)

**Non-Goals:**
- Eventos de dominio (para MVP)
- Validacion de timestamps (futuro, pasado)
- Control de concurrencia optimista (capa de infra)
- Validacion de consistencia entre mensaje y conversacion (redundante)

## Decisions

### 1. `ensureAcceptsMessages(at time.Time) error`

Metodo privado en `Conversation` que valida si la conversacion acepta mensajes en un momento dado.

```
Logica:
  if status != expired && status != resolved
    return nil  (abierta, acepta mensajes)

  if at < finishedAt
    return nil  (mensaje tardio valido)

  return ErrConversationNotAcceptingMessages  (llego tarde)
```

**Por que no valida finishedAt?** El invariante se valida en `NewConversation`. El aggregate garantiza su propia coherencia en construccion, los metodos operan sobre un estado valido.

**Por que privado?** Es logica interna del aggregate, no expone comportamiento a otros modulos.

**Por que `ensure` y no `validate`?** Sigue convenciones DDD - `ensure` expresa un invariante que se cumple.

### 2. `SendAgentMessage(agentID uuid.UUID, msg *Message, at time.Time) error`

Metodo publico en `Conversation` para envio de mensajes por agentes.

```
Flujo:
  1. ensureAcceptsMessages(at)
  2. if conv.agentID == nil → claim (assign agent, status=assigned)
  3. if conv.agentID != agentID → ErrConversationAgentNotOwner
  4. if msg.ExternalID() != nil → sync externalMsgIdx
  5. found[msg.ID()] = msg
     dirty[msg.ID()] = msg
  6. updatedAt = at
```

**Por que no validar status despues del claim?** Es redundante - acabamos de ponerlo en assigned.

**Por que no validar duplicado de externalID?** El agente crea mensajes nuevos con UUID propio. El externalID es opcional y solo se sincroniza el indice.

### 3. Refactor `*Message`

Cambiar todos los usos de `Message` a `*Message` para mantener identidad de entidad.

```
Antes:                          Despues:
found   map[uuid.UUID]Message   found   map[uuid.UUID]*Message
dirty   map[uuid.UUID]Message   dirty   map[uuid.UUID]*Message

ReceiveContactMessage(          ReceiveContactMessage(
  msg Message,                    msg *Message,
)                                 )

Messages() []Message            Messages() []*Message
```

**Por que?** Entidades en DDD se identifican por identidad, no por atributos. Pasar por valor crea copias que pierden referencia.

### 4. Invariante en `NewConversation`

Validar que conversaciones expired/resolved tengan finishedAt.

```
if status == expired || status == resolved
  if finishedAt == nil
    return nil, ErrConversationFinishedAtMissing
```

**Por que?** Garantiza que el aggregate siempre esta en estado valido. `ensureAcceptsMessages` puede confiar en este invariante.

### 5. Errores nuevos

```go
ErrConversationNotAcceptingMessages  // mensaje tardio, conversacion cerrada
ErrConversationFinishedAtMissing     // invariante violado
ErrConversationAgentNotOwner         // agente no es el asignado
```

### 6. Condiciones de carrera (claim)

No se resuelve en el dominio. La capa de infraestructura usa optimistic locking con version en la tabla de conversaciones.

```
UPDATE conversations 
SET agent_id = ?, status = ?, version = version + 1
WHERE id = ? AND version = ?
```

## Risks / Trade-offs

**[Riesgo] Mensaje tardio de agente**: Un agente podria enviar un mensaje a una conversacion ya cerrada si el timestamp es anterior a finishedAt. En arquitectura async con colas, esto es posible.
- **Mitigacion**: Es comportamiento esperado. El `ensureAcceptsMessages` valida correctamente.

**[Trade-off] No validar consistencia del mensaje**: No validamos que `msg.AgentID()` coincida con `conv.agentID` o que `msg.ContactID()` coincida con `conv.contactID`.
- **Razon**: Redundancia innecesaria. El caller construye el mensaje correctamente.

**[Trade-off] No eventos de dominio**: Para MVP, no emitimos eventos cuando el agente envia un mensaje o reclama una conversacion.
- **Razon**: Simplicidad. Se puede agregar despues sin cambiar el diseno.

## Open Questions

Ninguna - todas las decisiones se tomaron durante la exploracion.