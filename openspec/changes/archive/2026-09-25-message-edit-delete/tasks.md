# Tasks

## 1. Estado failed en MessageStatus

- [x] 1.1 Agregar estado `MessageStatusFailed` y aceptarlo en `NewMessageStatus` y verificar que compila
- [x] 1.2 Agregar tests table-driven para `NewMessageStatus` con `failed` y verificar que pasan

## 2. Errores nuevos

- [x] 2.1 Agregar errores `ErrMessageFailed`, `ErrMessageNotFound`, `ErrMessageAlreadyDeleted`, `ErrMessageNotText`, `ErrMessageNotFromAgent`, `ErrConversationFinished` y verificar que compila

## 3. Helpers de validacion en Message

- [x] 3.1 Implementar `ensureNotFailed() error` en `Message` y verificar que compila
- [x] 3.2 Implementar `ensureEditable(at time.Time) error` en `Message` y verificar que compila
- [x] 3.3 Agregar tests para `ensureEditable` via `EditText`: mensaje válido, mensaje fallido, mensaje eliminado con at posterior, mensaje eliminado con at anterior

## 4. EditText en Message

- [x] 4.1 Implementar `EditText(newText string, at time.Time) error` en `Message` y verificar que compila
- [x] 4.2 Agregar tests table-driven para `EditText`: éxito, texto vacío, texto muy largo, mensaje no es texto, mensaje fallido, mensaje eliminado

## 5. Delete en Message

- [x] 5.1 Implementar `Delete(at time.Time) error` en `Message` y verificar que compila
- [x] 5.2 Agregar tests table-driven para `Delete`: éxito, mensaje ya eliminado (cualquier timestamp), mensaje fallido

## 6. MarkAsFailed en Message

- [x] 6.1 Implementar `MarkAsFailed() error` en `Message` y verificar que compila
- [x] 6.2 Agregar tests table-driven para `MarkAsFailed`: éxito, idempotente, fallar si no es de agente

## 7. registerActivity en Conversation

- [x] 7.1 Implementar `registerActivity(at time.Time)` en `Conversation` y verificar que compila
- [x] 7.2 Reemplazar `c.updatedAt = &at` por `registerActivity(at)` en `ReceiveContactMessage`, `SendAgentMessage` y `AgentReadConversation` y verificar que los tests existentes pasan
- [x] 7.3 Agregar test: un mensaje tardío no retrocede `updatedAt`

## 8. ensureAgentCanModify en Conversation

- [x] 8.1 Implementar `ensureAgentCanModify() error` en `Conversation` y verificar que compila

## 9. AgentEditMessage en Conversation

- [x] 9.1 Implementar `AgentEditMessage(agentID uuid.UUID, msgID uuid.UUID, newText string, at time.Time) error` en `Conversation` y verificar que compila
- [x] 9.2 Agregar tests table-driven: éxito, agente no es dueño, mensaje no encontrado, mensaje no es del agente, conversación finalizada

## 10. AgentDeleteMessage en Conversation

- [x] 10.1 Implementar `AgentDeleteMessage(agentID uuid.UUID, msgID uuid.UUID, at time.Time) error` en `Conversation` y verificar que compila
- [x] 10.2 Agregar tests table-driven: éxito, agente no es dueño, mensaje no encontrado, mensaje no es del agente, conversación finalizada

## 11. MarkAgentMessageFailed en Conversation

- [x] 11.1 Implementar `MarkAgentMessageFailed(agentID uuid.UUID, msgID uuid.UUID, at time.Time) error` en `Conversation` y verificar que compila
- [x] 11.2 Agregar tests table-driven: éxito, agente no es dueño, mensaje no encontrado, mensaje de contacto, conversación finalizada (debe permitir)

## 12. ReceiveContactMessageEdit en Conversation

- [x] 12.1 Implementar `ReceiveContactMessageEdit(contactID uuid.UUID, externalID string, newText string, at time.Time) error` en `Conversation` y verificar que compila
- [x] 12.2 Agregar tests table-driven: éxito, contacto no es dueño, mensaje no encontrado por externalID, mensaje no es del contacto, conversación finalizada (debe permitir)

## 13. ReceiveContactMessageDelete en Conversation

- [x] 13.1 Implementar `ReceiveContactMessageDelete(contactID uuid.UUID, externalID string, at time.Time) error` en `Conversation` y verificar que compila
- [x] 13.2 Agregar tests table-driven: éxito, contacto no es dueño, mensaje no encontrado por externalID, mensaje no es del contacto, conversación finalizada (debe permitir)

## 14. Verificacion final

- [x] 14.1 Ejecutar `go test ./internal/modules/conversations/domain/...` y verificar que todos los tests pasan
- [x] 14.2 Ejecutar `go vet ./...` y verificar que no hay errores
- [x] 14.3 Ejecutar `gofmt -l .` y verificar que el formato es correcto
- [x] 14.4 Ejecutar `go test -coverprofile=coverage.out ./internal/modules/conversations/domain/...` y verificar cobertura 100%

## 15. Monotonia de edicion

- [x] 15.1 Implementar descarte de edicion stale en `EditText` (solo aplica si `at` es posterior a `editedAt`) y verificar que compila
- [x] 15.2 Agregar test Message: una edicion stale conserva el texto y `editedAt` mas recientes
- [x] 15.3 Agregar test Conversation: `ReceiveContactMessageEdit` stale conserva la ultima edicion
- [x] 15.4 Ejecutar `go test -coverprofile=coverage.out ./internal/modules/conversations/domain/...` y verificar cobertura 100%