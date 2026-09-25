# Tasks

## 1. Helper ensureNotFinished

- [x] 1.1 Implementar `ensureNotFinished() error` en `Conversation`, reemplazar `ensureAgentCanModify` por el guard y reutilizar `ErrConversationFinished`; verificar que compila

## 2. ExpireConversation

- [x] 2.1 Implementar `ExpireConversation(at time.Time) error` en `Conversation` y verificar que compila
- [x] 2.2 Agregar tests table-driven: expira pending, expira assigned, falla si ya finalizada, registra actividad

## 3. ResolveConversation

- [x] 3.1 Implementar `ResolveConversation(agentID uuid.UUID, at time.Time) error` en `Conversation` y verificar que compila
- [x] 3.2 Agregar tests table-driven: resuelve assigned, falla si agente no es el asignado, falla si no tiene agente, falla si ya finalizada, registra actividad

## 4. Orden de mensajes

- [x] 4.1 Ordenar `Messages()` por `sentAt` ascendente con desempate por `id` y verificar que compila
- [x] 4.2 Agregar test: los mensajes se devuelven ordenados por `sentAt` aunque se hayan agregado en desorden
- [x] 4.3 Agregar test: mensajes con el mismo `sentAt` se desempatan por `id` de forma determinista

## 5. Verificacion final

- [x] 5.1 Ejecutar `go test ./internal/modules/conversations/domain/...` y verificar que todos los tests pasan
- [x] 5.2 Ejecutar `go vet ./...` y verificar que no hay errores
- [x] 5.3 Ejecutar `gofmt -l .` y verificar que el formato es correcto
- [x] 5.4 Ejecutar `go test -coverprofile=coverage.out ./internal/modules/conversations/domain/...` y verificar cobertura 100%