# Tasks

## 1. Dependencias y scaffold

- [x] 1.1 Agregar chi v5.3.2 (`go get github.com/go-chi/chi/v5@v5.3.2`) y verificar con `go mod verify && go build ./...`.
- [x] 1.2 Crear `internal/shared/infra/http/{middleware,httperror}` y verificar que compila con `go build ./...`.

## 2. Persistencia de usuarios

- [x] 2.1 Renombrar `db/migrations/000001_create_agents.*` a `000001_create_users.*` creando `users(id, created_at)`; verificar con `migrate up` sobre una BD limpia que la tabla existe.
- [x] 2.2 Editar `000003_create_conversations` y `000004_create_messages` cambiando `agent_id` por `user_id` y la FK a `users`; verificar `migrate up` y `migrate down` sin error.
- [x] 2.3 Renombrar `db/queries/agents/agents.sql` a `db/queries/users/users.sql` con `FindUserByID` y `CreateUser` (INSERT con ON CONFLICT por id) apuntando a `users`; verificar que `sqlc generate` produce el modelo `User` y compila.
- [x] 2.4 Renombrar `internal/modules/agents/infra/pg` a `internal/modules/users/infra/pg` e implementar `FindByID` y `Save` sobre sqlc; verificar con un test de repositorio contra PostgreSQL (o smoke test) que `Save` inserta y `FindByID` recupera el usuario.

## 3. Módulo users

- [x] 3.1 Renombrar `internal/modules/agents` a `internal/modules/users` y los tipos (`Agent`->`User`, `AgentsAPI`->`UsersAPI`, `ErrAgentNotFound`->`ErrUserNotFound`); verificar `go build ./...` sin referencias colgantes.
- [x] 3.2 Implementar `domain.NewUser` (valida id no nulo) y `RehydrateUser`; verificar con tests table-driven casos válido e id nulo.
- [x] 3.3 Implementar el caso de uso `CreateUser` y el `Save` del repositorio; verificar con test unitario sobre mock que crea y propaga errores.
- [x] 3.4 Exponer `FindUserByID` y `CreateUser` en `UsersAPI`; verificar con test unitario sobre mock de la API.

## 4. Ajustes en conversaciones

- [x] 4.1 Actualizar los casos de uso de conversaciones para depender de `users.UsersAPI` (conservando `Agent` en domain/app); verificar con `go test ./internal/modules/conversations/...`.
- [x] 4.2 Actualizar el mapper de `conversations/infra/pg` para leer/escribir `user_id` y mapear a `Agent.ID()`; verificar `go build ./...` y los tests de conversaciones.

## 5. HTTP compartido

- [x] 5.1 Implementar `middleware.Auth` (Bearer uuid -> contexto) y `UserIDFrom`; verificar con tests table-driven de 401 por header ausente y no parseable, y éxito con Bearer válido.
- [x] 5.2 Implementar `httperror.Write` (RFC 7807, `application/problem+json`, sin `type`); verificar con test que valida `Content-Type` y campos `title`, `status`, `detail`.

## 6. Handlers por módulo

- [x] 6.1 Implementar `users` handler + DTO y `POST /v1/users`; verificar tests que cubren 201 con el id en el cuerpo y 422 por entrada inválida.
- [x] 6.2 Implementar `conversations` handler + DTO y `POST /v1/conversations/{id}/messages`; verificar tests que cubren 201, 400 por id inválido, 403, 404, 409 y 422.
- [x] 6.3 Implementar `PATCH /v1/conversations/{id}/messages` con dispatch de `status=read`; verificar tests que cubren 204, 422 por status no permitido y 403.
- [x] 6.4 Implementar `PATCH /v1/conversations/{id}` con dispatch de `status=resolved`; verificar tests que cubren 204, 422 para `expired` y 403.
- [x] 6.5 Implementar el mapeo de errores de dominio a códigos HTTP en cada handler; verificar con test que cada sentinel acotado (`ErrConversationNotFound`, `ErrConversationAgentNotOwner`, `ErrConversationFinished`, `ErrUserNotFound`) devuelve 404/403/409/401.

## 7. Composition root

- [x] 7.1 Implementar `cmd/api/main.go` con lectura de env, `pgdb.New`, armado de repos/use cases/handlers, router chi con middlewares globales y apagado graceful; verificar que `go run ./cmd/api` arranca y `GET /healthz` responde 200 (la corrida en vivo queda diferida por falta de docker; `/healthz` cubierto por test de router).
- [x] 7.2 Montar el grupo `/v1`: `POST /v1/users` público y el resto bajo `middleware.Auth`; verificar con un test de router que una ruta autenticada sin Bearer responde 401 y `/healthz` responde 200.
- [x] 7.3 Cablear el emisor saliente stub (`messagesender.StubSender`) en los casos de uso de conversaciones; verificar con `go test ./internal/modules/conversations/...`.
- [x] 7.4 Implementar el timeout de request (5s) con mapeo a 504 y los timeouts de servidor; verificar con tests de `httperror`/handlers y `go build ./...`.

## 8. Verificación final

- [x] 8.1 Ejecutar `go mod verify && gofmt -l . && go vet ./... && go build ./... && go test -race -count=1 ./...` y verificar que todo queda limpio.
- [x] 8.2 Ejecutar `openspec validate agent-rest-write --strict` y verificar que pasa.
- [x] 8.3 Revisar que no queden referencias a `agents` fuera de `internal/modules/conversations` (por ejemplo con `grep -rn "agent" internal --include=*.go`) y confirmar que la única aparición es el término de dominio dentro de conversaciones.
