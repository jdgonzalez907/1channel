# Tasks

## 1. Estructura del modulo y Value Objects

- [x] 1.1 Crear directorio `internal/module/conversations/domain/` y verificar que existe
- [x] 1.2 Crear archivo `conversation_status.go` con tipo `ConversationStatus`, constantes (pending, assigned, expired, resolved), error `ErrConversationStatusInvalid`, constructor `NewConversationStatus` que valida contra enum, y metodo `String()` - verificar que compila
- [x] 1.3 Crear archivo `message_status.go` con tipo `MessageStatus`, constantes (sent, read, deleted), error `ErrMessageStatusInvalid`, constructor `NewMessageStatus` que valida contra enum, y metodo `String()` - verificar que compila
- [x] 1.4 Crear archivo `message_type.go` con tipo `MessageType`, constante (text), error `ErrMessageTypeInvalid`, constructor `NewMessageType` que valida contra enum, y metodo `String()` - verificar que compila
- [x] 1.5 Crear tests para los constructores de Value Objects con casos validos e invalidos - verificar que tests pasan

## 2. Entity Message

- [x] 2.1 Crear archivo `message.go` con constantes `MinTextLength` y `MaxTextLength` - verificar que compila
- [x] 2.2 Crear struct `Message` con todos sus campos (id, status, type, text, agentID, contactID, sentAt, readAt, editedAt, deletedAt) - verificar que compila
- [x] 2.3 Implementar errores scoped: `ErrMessageInvalidID`, `ErrMessageEmptyText`, `ErrMessageTextTooLong` - verificar que compila
- [x] 2.4 Implementar constructor `NewMessage(...)` con validaciones: uuid no Nil, texto con longitud valida (1-1000 runas) - verificar que compila
- [x] 2.5 Implementar getters en una linea para todos los campos - verificar que compila
- [x] 2.6 Crear tests para `NewMessage` con casos validos e invalidos (uuid invalido, status invalido, type invalido, texto vacio, texto muy largo) - verificar que tests pasan

## 3. Aggregate Conversation

- [x] 3.1 Crear archivo `conversation.go` con struct `Conversation` y todos sus campos (id, status, found, dirty, agentID, contactID, createdAt, updatedAt, finishedAt) - verificar que compila
- [x] 3.2 Implementar errores scoped: `ErrConversationInvalidID`, `ErrConversationMissingContactAndAgent` - verificar que compila
- [x] 3.3 Implementar constructor `NewConversation(...)` con validaciones: uuid no Nil, contactID o agentID presente, llena `found` con mensajes, `dirty` vacio - verificar que compila
- [x] 3.4 Implementar getters en una linea para todos los campos, `Messages()` itera `found` y construye slice - verificar que compila
- [x] 3.5 Crear tests para `NewConversation` con casos validos e invalidos (uuid invalido, status invalido, sin contactID ni agentID) - verificar que tests pasan

## 4. Verificacion final

- [x] 4.1 Ejecutar `go vet ./internal/module/conversations/...` sin errores
- [x] 4.2 Ejecutar `go test ./internal/module/conversations/...` con todos los tests pasando
