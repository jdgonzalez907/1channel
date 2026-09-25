# Design

## Context

El aggregate `Conversation` solo tiene la transición `pending -> assigned` (claim implícito en `SendAgentMessage`). `finishedAt` solo se asigna por constructor, así que no se puede finalizar una conversación en runtime. `Messages()` itera un `map`, por lo que el orden es aleatorio.

Ver proposal.md para la motivación del cambio.

## Goals / Non-Goals

**Goals:**
- Permitir expirar y resolver una conversación en runtime
- Hacer las transiciones terminales explícitas y validadas
- Devolver los mensajes en orden de timeline (`sentAt`)

**Non-Goals:**
- Transferencia de agente, reintentos, contadores de no leídos
- Read receipts del contacto sobre mensajes del agente
- Expiración automática por configuración (app/infra)
- Casos de uso y persistencia

## Decisions

### 1. `ExpireConversation(at time.Time) error` en Conversation

```
Flujo:
  1. ensureNotFinished()
  2. status = ConversationStatusExpired
  3. finishedAt = &at
  4. registerActivity(at)
```

**Por qué no valida agente?** Expirar es un evento de sistema (configuración/tiempo), no una acción del agente.

**Por qué puede expirar una pending?** Una conversación que nadie tomó también vence.

### 2. `ResolveConversation(agentID uuid.UUID, at time.Time) error` en Conversation

```
Flujo:
  1. ensureNotFinished()
  2. if c.agentID == nil || *c.agentID != agentID → ErrConversationAgentNotOwner
  3. status = ConversationStatusResolved
  4. finishedAt = &at
  5. registerActivity(at)
```

**Por qué valida agente?** Resolver es una acción del agente; una conversación pending (sin agente) no puede resolverse.

### 3. `ensureNotFinished() error` en Conversation

```
Flujo:
  if status == expired || status == resolved → ErrConversationFinished
  return nil
```

**Por qué un solo guard?** La condición de "finalizada" se usa en cuatro operaciones (`ExpireConversation`, `ResolveConversation`, `AgentEditMessage`, `AgentDeleteMessage`). Un guard que devuelve el error evita duplicar la condición y el error.

### 4. `Messages()` ordenado por `sentAt`

```
Flujo:
  1. copiar los mensajes de found a un slice
  2. ordenar por sentAt ascendente
  3. desempatar por id con uuid.UUID.Compare
```

**Por qué ordenar en el dominio?** El orden de la conversación es una propiedad del aggregate (timeline), no del read-model. Un mensaje tardío con `sentAt` anterior se ubica en el pasado.

**Por qué desempatar por id?** Dos mensajes pueden compartir `sentAt`. Sin desempate, el orden sería no determinista (map). `uuid.UUID` es comparable con `Compare` y los ids V7 son time-ordered.

**Alternativa descartada**: orden de inserción (llegada). No refleja la conversación real cuando hay mensajes fuera de orden.

### 5. Errores

No se agregan errores nuevos. Se reutiliza `ErrConversationFinished` tanto para modificar como para finalizar una conversación ya terminada, con el mensaje ajustado a `"conversation is finished"`.

## Risks / Trade-offs

**[Trade-off] Un único error para "finalizada"**: modificar y finalizar comparten `ErrConversationFinished`.
- **Razón**: la distinción entre "no podés modificar" y "ya está finalizada" no aporta al consumidor; ambos casos son el mismo estado de negocio.

**[Trade-off] Ordenar en cada llamada a `Messages()`**: se ordena en cada invocación.
- **Razón**: simplicidad para MVP. Si el volumen crece, se puede cachear o mover al read-model.

## Open Questions

Ninguna - todas las decisiones se tomaron durante la exploración.