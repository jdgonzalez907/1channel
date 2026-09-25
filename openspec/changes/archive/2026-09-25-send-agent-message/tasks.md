# Tasks

## 1. Refactor *Message

- [x] 1.1 Cambiar maps `found` y `dirty` de `Message` a `*Message` en struct `Conversation` y verificar que compila
- [x] 1.2 Cambiar parametro `msg Message` a `msg *Message` en `ReceiveContactMessage` y verificar que compila
- [x] 1.3 Cambiar return type de `Messages()` de `[]Message` a `[]*Message` y verificar que compila
- [x] 1.4 Cambiar parametro `messages []Message` a `[]*Message` en `NewConversation` y verificar que compila
- [x] 1.5 Actualizar todos los tests para usar `*Message` en vez de `Message` y verificar que pasan

## 2. Nuevos errores

- [x] 2.1 Agregar errores `ErrConversationNotAcceptingMessages`, `ErrConversationFinishedAtMissing`, `ErrConversationAgentNotOwner` al archivo de errores y verificar que compila

## 3. Invariante en NewConversation

- [x] 3.1 Agregar validacion en `NewConversation`: si status es expired/resolved, finishedAt no puede ser nil y verificar con test

## 4. ensureAcceptsMessages

- [x] 4.1 Implementar metodo privado `ensureAcceptsMessages(at time.Time) error` en `Conversation` y verificar que compila
- [x] 4.2 Agregar tests para `ensureAcceptsMessages`: conversacion abierta, cerrada con mensaje tardio, cerrada con mensaje antes de finishedAt, finishedAt nil

## 5. Actualizar ReceiveContactMessage

- [x] 5.1 Agregar llamada a `ensureAcceptsMessages(at)` como primer paso en `ReceiveContactMessage` y verificar que compila
- [x] 5.2 Agregar test para recibir mensaje en conversacion que no acepta mensajes y verificar que pasa

## 6. SendAgentMessage

- [x] 6.1 Implementar metodo `SendAgentMessage(agentID uuid.UUID, msg *Message, at time.Time) error` en `Conversation` y verificar que compila
- [x] 6.2 Agregar test: enviar mensaje exitosamente en conversacion asignada
- [x] 6.3 Agregar test: reclamar conversacion pendiente (claim)
- [x] 6.4 Agregar test: fallar si agente no es el asignado
- [x] 6.5 Agregar test: sincronizar externalID si existe
- [x] 6.6 Agregar test: enviar mensaje sin externalID

## 7. Verificacion final

- [x] 7.1 Ejecutar `go test ./internal/modules/conversations/domain/...` y verificar que todos los tests pasan
- [x] 7.2 Ejecutar `go vet ./...` y verificar que no hay errores
- [x] 7.3 Ejecutar `gofmt -l .` y verificar que el formato es correcto