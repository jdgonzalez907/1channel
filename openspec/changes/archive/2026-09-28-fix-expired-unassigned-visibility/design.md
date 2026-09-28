# Design

## Context

Ver `proposal.md - Why`. La visibilidad vive en dos lugares que deben coincidir:
- `ListConversationsForAgent` (`db/queries/conversations/conversations.sql`): `WHERE (c.status = 'pending' OR c.user_id = :agent_id)`.
- `isConversationVisible` (`conversation_read_handler.go`): `pending` o `user_id == agente`.

El detalle usa `FindConversationWithContactByID` (sin filtro de visibilidad) y luego `isConversationVisible`. El read model (`last_message_*`, `unread_count`) no interviene en la visibilidad. El estado de las conversaciones no cambia con este fix.

## Goals / Non-Goals

**Goals:**
- Que una conversación sin agente asignado (incluida la `expired`) sea visible y abrible por cualquier agente.
- Mantener la privacidad de `assigned`/`resolved`/`expired` con agente.

**Non-Goals:**
- No se cambian estados, expiración, ni el read model.
- No se toca el seed (sus `expired` sin agente pasan a mostrarse).
- No se agrega paginación ni filtros nuevos.

## Decisions

### 1. Regla por agente asignado, no por estado

Se reemplaza `(c.status = 'pending' OR c.user_id = :agent_id)` por `(c.user_id IS NULL OR c.user_id = :agent_id)`. Es uniforme: cubre `pending` (siempre sin agente) y `expired` sin agente con la misma condición, sin enumerar estados. Alternativa considerada: `status = 'pending' OR (status = 'expired' AND user_id IS NULL) OR user_id = :agent_id`; descartada por más verbosa y acoplada al estado.

### 2. Misma regla en la query y en el detalle

`isConversationVisible` pasa a devolver verdadero cuando la conversación no tiene agente asignado o cuando el asignado es el solicitante. Si no se cambia, una fila recién listada respondería 403 al abrirse.

### 3. Sin cambios en persistencia ni datos

`ListConversationsForAgent` ya ordena por `last_message_at, id`; solo cambia el `WHERE`. No hay migración ni `RefreshConversationLastMessage`.

## Risks / Trade-offs

- **Índice y OR sobre `user_id`** -> `(user_id IS NULL OR user_id = :agent_id)` no usa tan directo el índice `(user_id, last_message_at DESC, id DESC)` para el brazo `IS NULL`. Mitigación: en MVP el conjunto de conversaciones sin agente es acotado; el orden sigue saliendo del índice y el `JOIN` del preview es por PK. Si escalara, se evaluaría un índice parcial.
- **Exposición de conversaciones sin agente** -> Es intencional: no tienen dueño, el historial no debe perderse. No aplica a las que sí tienen agente.
- **Consistencia lista/detalle** -> El test existente "finished without agent is not visible" se invierte; hay que cubrir ambos caminos en `TestIsConversationVisible`.
