# Tasks

## 1. Error nuevo

- [x] 1.1 Agregar error `ErrMessageInvalidOwner` al archivo de errores de Message y verificar que compila

## 2. Invariante en NewMessage

- [x] 2.1 Agregar validacion en `NewMessage`: agentID XOR contactID (exactamente uno) y verificar que compila
- [x] 2.2 Agregar tests para el invariante: sin dueño, con dos dueños, con agentID solo, con contactID solo

## 3. MarkAsRead en Message

- [x] 3.1 Implementar metodo `MarkAsRead(at time.Time)` en `Message` y verificar que compila
- [x] 3.2 Agregar test: marcar mensaje como leído y verificar status y readAt

## 4. AgentReadConversation en Conversation

- [x] 4.1 Implementar metodo `AgentReadConversation(agentID uuid.UUID, at time.Time) error` en `Conversation` y verificar que compila
- [x] 4.2 Agregar test: marcar mensajes del contacto como leídos exitosamente
- [x] 4.3 Agregar test: fallar si agente no es el asignado
- [x] 4.4 Agregar test: no marcar mensajes del agente como leídos
- [x] 4.5 Agregar test: marcar mensajes en conversacion finalizada
- [x] 4.6 Agregar test: no marcar mensajes ya leídos

## 5. Verificacion final

- [x] 5.1 Ejecutar `go test ./internal/modules/conversations/domain/...` y verificar que todos los tests pasan
- [x] 5.2 Ejecutar `go vet ./...` y verificar que no hay errores
- [x] 5.3 Ejecutar `gofmt -l .` y verificar que el formato es correcto
- [x] 5.4 Ejecutar `go test -coverprofile=coverage.out ./internal/modules/conversations/domain/...` y verificar cobertura 100%