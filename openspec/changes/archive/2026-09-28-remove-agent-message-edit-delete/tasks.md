# Tasks

## 1. Dominio: quitar edición/borrado de agente

- [x] 1.1 Eliminar los métodos `AgentEditMessage` y `AgentDeleteMessage` de `backend/internal/modules/conversations/domain/conversation.go`; verificar con `go -C backend build ./...`
- [x] 1.2 Eliminar `TestConversation_AgentEditMessage` y `TestConversation_AgentDeleteMessage` de `backend/internal/modules/conversations/domain/conversation_test.go`; verificar con `go -C backend test ./internal/modules/conversations/domain/...`

## 2. Verificación de referencias

- [x] 2.1 Confirmar que no quedan referencias a `AgentEditMessage` ni `AgentDeleteMessage` en `backend` (grep), y que `Message.EditText`/`Message.Delete` siguen usados por los use cases del contacto; verificar con `go -C backend build ./...`

## 3. Quality gates

- [x] 3.1 Ejecutar `go -C backend mod verify && gofmt -l backend && go -C backend vet ./...` y verificar que `gofmt` no reporta archivos
- [x] 3.2 Ejecutar `go -C backend build ./...` y verificar que compila
- [x] 3.3 Ejecutar `go -C backend test -race -count=1 ./...` y verificar que todos los tests pasan
