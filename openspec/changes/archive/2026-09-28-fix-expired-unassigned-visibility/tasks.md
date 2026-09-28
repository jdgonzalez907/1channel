# Tasks

## 1. Visibilidad en la consulta de la bandeja

- [x] 1.1 En `db/queries/conversations/conversations.sql` cambiar la condición de visibilidad de `ListConversationsForAgent` a `(c.user_id IS NULL OR c.user_id = sqlc.arg('agent_id'))`; correr `sqlc generate` (desde `backend/`) y verificar con `go -C backend build ./...`

## 2. Visibilidad en el detalle y tests

- [x] 2.1 Actualizar `isConversationVisible` en `conversation_read_handler.go` para devolver verdadero cuando la conversación no tiene agente asignado, además de cuando el asignado es el solicitante; verificar con `go -C backend build ./...`
- [x] 2.2 En `conversation_read_handler_test.go`, invertir el caso "finished without agent is not visible" a visible y agregar un caso de `expired` sin agente junto a uno de `expired` con otro agente (no visible); verificar con `go -C backend test ./internal/modules/conversations/infra/http/...`

## 3. Documentación

- [x] 3.1 Actualizar la nota de visibilidad de conversaciones en `backend/docs/endpoints.md` para reflejar "sin agente asignado (incluye `expired` sin agente) o asignada al solicitante"; verificar que el texto documentado coincide con el comportamiento

## 4. Verificación manual con base de datos

- [x] 4.1 Con `make reset && make seed` y `make run-api`, confirmar con `curl` que `GET /v1/conversations?status=finished` incluye las `expired` sin agente del seed (`conv 21..30`) para cualquier agente y que `GET /v1/conversations/{id}` las abre sin 403

## 5. Quality gates

- [x] 5.1 Ejecutar `go -C backend mod verify && gofmt -l backend && go -C backend vet ./...` y verificar que `gofmt` no reporta archivos
- [x] 5.2 Ejecutar `go -C backend build ./...` y verificar que compila
- [x] 5.3 Ejecutar `go -C backend test -race -count=1 ./...` y verificar que todos los tests pasan
