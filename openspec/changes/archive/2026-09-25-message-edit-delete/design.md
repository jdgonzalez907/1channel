# Design

## Context

El aggregate `Conversation` soporta envío y recepción de mensajes, pero no permite editarlos ni eliminarlos. Los mensajes tienen campos `editedAt` y `deletedAt` que no se utilizan. Falta el estado `failed` para mensajes no entregados. `updatedAt` se asigna manualmente sin control de monotonía.

Ver proposal.md para la motivación del cambio.

## Goals / Non-Goals

**Goals:**
- Permitir editar texto de mensajes existentes
- Permitir eliminar mensajes (soft delete)
- Marcar como fallidos los mensajes de agente no entregados
- Impedir que el agente modifique mensajes en conversaciones finalizadas
- Permitir que el contacto modifique mensajes incluso en conversaciones finalizadas
- Mantener `updatedAt` siempre en su valor más reciente

**Non-Goals:**
- Editar mensajes de tipo adjunto (para otro momento)
- Reenviar mensajes (para otro momento)
- Validación de timestamps (futuro, pasado)

## Decisions

### 1. `EditText(newText string, at time.Time) error` en Message

```
Flujo:
  1. ensureEditable(at)
  2. if text == nil → ErrMessageNotText
  3. if editedAt != nil && !at.After(*editedAt) → return nil (edición stale, se descarta)
  4. validate newText: no vacío (ErrMessageEmptyText), máx 1000 (ErrMessageTextTooLong)
  5. text = newText
  6. editedAt = at
```

**Por qué valida el texto?** Misma regla que `NewMessage`: un mensaje de texto no puede quedar vacío ni exceder el máximo.

**Por qué descarta ediciones stale?** En sistemas distribuidos las ediciones llegan fuera de orden. La verdad de negocio es la edición más reciente; una edición con timestamp anterior o igual a `editedAt` se descarta silenciosamente (no es un error, es un evento superado).

**Por qué no valida el dueño?** La validación de dueño se hace en `Conversation`. El método en Message solo ejecuta la acción.

### 2. `Delete(at time.Time) error` en Message

```
Flujo:
  1. ensureNotFailed()
  2. if deletedAt != nil → ErrMessageAlreadyDeleted
  3. deletedAt = at
  4. status = MessageStatusDeleted
```

**Por qué rechaza sin importar el timestamp?** Re-eliminar no tiene sentido de negocio y movería `deletedAt` hacia atrás. El "late arrival" solo aplica a ediciones (traza).

**Por qué soft delete?** El mensaje se queda en `found` para que la conversación lo muestre como eliminado.

### 3. `MarkAsFailed() error` en Message

```
Flujo:
  1. if agentID == nil → ErrMessageNotFromAgent
  2. status = MessageStatusFailed
```

**Por qué solo mensajes del agente?** El fallo de entrega es un evento de los envíos salientes (el agente envía a WhatsApp). Un mensaje de contacto no puede "fallar" en entrega.

**Por qué es idempotente sin guard?** Reasignar el mismo valor es naturalmente idempotente. No se necesita un `if` explícito.

**Por qué failed prevalece sobre deleted?** failed es terminal y bloquea todo. Si un mensaje ya eliminado recibe un reporte de fallo, el estado queda `failed`. Es un caso de borde aceptado para MVP.

### 4. Helpers de validación en Message

```
ensureNotFailed() error:
  if status == MessageStatusFailed → ErrMessageFailed
  return nil

ensureEditable(at time.Time) error:
  if err := ensureNotFailed(); err != nil → return err
  if deletedAt != nil && at >= *deletedAt → ErrMessageAlreadyDeleted
  return nil
```

**Por qué `ensureEditable` permite at < deletedAt?** En sistemas distribuidos los edits llegan fuera de orden. Un edit enviado ANTES de la eliminación se aplica como traza del estado anterior.

**Por qué `Delete` no usa `ensureEditable`?** Porque `Delete` debe rechazar cualquier mensaje ya eliminado, sin importar el timestamp.

### 5. `registerActivity(at time.Time)` en Conversation

```
Flujo:
  if updatedAt == nil || at.After(*updatedAt):
    updatedAt = &at
```

**Por qué?** `updatedAt` representa la última actividad real. Un mensaje tardío no debe retroceder el timestamp. Reemplaza `c.updatedAt = &at` en `ReceiveContactMessage`, `SendAgentMessage`, `AgentReadConversation` y todos los métodos nuevos.

### 6. `ensureAgentCanModify() error` en Conversation

```
Flujo:
  if status == expired || status == resolved → ErrConversationFinished
  return nil
```

**Por qué distinto de `ensureAcceptsMessages`?** `ensureAcceptsMessages` acepta llegadas tardías (at < finishedAt). El agente, al estar bajo nuestro control, no puede modificar una conversación finalizada sin importar el timestamp.

### 7. `AgentEditMessage` y `AgentDeleteMessage` en Conversation

```
Flujo (ambos):
  1. ensureAgentCanModify()
  2. if agentID != conv.agentID → ErrConversationAgentNotOwner
  3. find msg by msgID
  4. if msg == nil → ErrMessageNotFound
  5. if msg.AgentID() == nil → ErrConversationAgentNotOwner
  6. msg.EditText/Delete
  7. dirty[msgID] = msg
  8. registerActivity(at)
```

**Por qué valida conversación finalizada?** Tenemos control sobre el agente; si la conversación terminó, el agente debe estar sincronizado.

### 8. `MarkAgentMessageFailed(agentID, msgID, at)` en Conversation

```
Flujo:
  1. if agentID != conv.agentID → ErrConversationAgentNotOwner
  2. find msg by msgID
  3. if msg == nil → ErrMessageNotFound
  4. msg.MarkAsFailed() → propaga ErrMessageNotFromAgent
  5. dirty[msgID] = msg
  6. registerActivity(at)
```

**Por qué no valida conversación finalizada?** El fallo de entrega es un reporte tardío de la infraestructura, no una acción de usuario. Igual que las operaciones del contacto.

**Por qué no repite el chequeo de dueño?** `MarkAsFailed` ya valida que el mensaje sea del agente y retorna `ErrMessageNotFromAgent`; el aggregate solo propaga ese error.

### 9. `ReceiveContactMessageEdit` y `ReceiveContactMessageDelete` en Conversation

```
Flujo (ambos):
  1. if contactID != conv.contactID → ErrConversationContactNotOwner
  2. find msgID by externalID in externalMsgIdx
  3. if msgID not found → ErrMessageNotFound
  4. if msg.ContactID() == nil → ErrConversationContactNotOwner
  5. msg.EditText/Delete
  6. dirty[msgID] = msg
  7. registerActivity(at)
```

**Por qué no valida conversación finalizada?** No tenemos control sobre el contacto; los eventos de WhatsApp llegan aun después de finalizada.

### 10. Estado `failed` en MessageStatus

```go
MessageStatusFailed MessageStatus = "failed"
```

`NewMessageStatus` debe aceptar `failed` además de sent, read y deleted. Requiere MODIFIED del requerimiento "Estado de mensaje".

### 11. Errores nuevos

```go
ErrMessageFailed         = errors.New("message failed to deliver, no modifications allowed")
ErrMessageNotFound       = errors.New("message not found")
ErrMessageAlreadyDeleted = errors.New("message is already deleted")
ErrMessageNotText        = errors.New("message is not a text message")
ErrMessageNotFromAgent   = errors.New("message was not sent by an agent")
ErrConversationFinished  = errors.New("conversation is finished, agent cannot modify")
```

## Risks / Trade-offs

**[Trade-off] failed prevalece sobre deleted**: Un mensaje eliminado que luego se marca fallido pierde el status `deleted` (conserva `deletedAt`).
- **Razón**: `failed` es terminal y bloquea todo. Caso de borde aceptado para MVP.

**[Trade-off] Soft delete vs hard delete**: Usamos soft delete en vez de remover el mensaje del map.
- **Razón**: Mantiene el historial y permite mostrar el mensaje como eliminado en la UI.

**[Riesgo] Datos existentes sin dueño**: El invariante de dueño ya existe; no aplica a este cambio.

## Open Questions

Ninguna - todas las decisiones se tomaron durante la exploración.