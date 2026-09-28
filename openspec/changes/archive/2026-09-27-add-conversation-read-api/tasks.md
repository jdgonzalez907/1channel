# Tasks

## 1. Esquema y queries del read model

- [x] 1.1 Editar las migraciones existentes (dev, nada productivo): en `000003_create_conversations.up.sql` agregar las columnas `last_message_at timestamptz`, `last_message_id uuid` y `unread_count integer NOT NULL DEFAULT 0 CHECK (unread_count >= 0)` y los índices `(user_id, last_message_at DESC, id DESC)`, `(status, last_message_at DESC, id DESC)` y `(contact_id)`; en `000004_create_messages.up.sql` cambiar `messages_conversation_id_idx` a `messages (conversation_id, sent_at DESC, id DESC)`; recrear la BD (`make reset`, o `make migrate-down` hasta la versión 2 y `make migrate-up`) y verificar que las migraciones corren sin error y las columnas e índices existen.
- [x] 1.2 Agregar en `db/queries/conversations/` la query `RefreshConversationLastMessage` (CTE del mensaje de mayor `(sent_at, id)` + CTE del conteo de no leídos, y `UPDATE conversations` de `last_message_id`, `last_message_at` y `unread_count`); verificar que `sqlc generate` corre sin error y aparece el método.
- [x] 1.3 Agregar las queries de lectura: `ListConversationsForAgent` (sin `LATERAL`, con `JOIN messages` por `last_message_id`, visibilidad, `status = ANY`, filtro opcional `contact_id`, posición `(last_message_at, id)` y `LIMIT page_size + 1`), `FindConversationWithContactByID` (metadata + contacto + `unread_count`) y `ListConversationMessagesPage` (posición `(sent_at, id)`); verificar que `sqlc generate` corre sin error y aparecen los tres métodos.

## 2. Write side: refresh en la transacción de Save

- [x] 2.1 En el repositorio de conversaciones (`internal/modules/conversations/infra/pg`), llamar `RefreshConversationLastMessage` dentro del `InTx` de `Save`, después de upsertear los mensajes, y **solo si hay mensajes modificados** (`len(dirty) > 0`); verificar que `go build ./...` compila.

## 3. Lectura de conversaciones (bandeja y detalle)

- [x] 3.1 Definir los DTOs de lista y detalle (fila con `contact.id` y `contact.external_id`, preview con texto y emisor) y la paginación por posición (`before_sent_at` / `before_id`, sin cursor opaco) en `internal/modules/conversations/infra/http`; verificar con unit tests de la lógica pura.
- [x] 3.2 Implementar `GET /v1/conversations` con `status` obligatorio (`open|finished`; ausente u otro 422) y `external_contact_id` opcional (404 si el contacto no existe, sin consultar conversaciones) más paginación (20, `next_before_sent_at` / `next_before_id` nulos al agotar); verificar con unit tests de la lógica pura (visibilidad, `open`/`finished`, preview, `unread_count`, paginación).
- [x] 3.3 Implementar `GET /v1/conversations/{id}` con metadata, últimos 20 mensajes en orden ascendente, página hacia atrás y `deleted` con texto nulo; verificar `go build` y los unit tests de la lógica pura (visibilidad, mapeo, paginación).
- [x] 3.4 Sin interfaz intermedia: el handler recibe `*sqlc.Queries` directo y la lógica pura (status, visibilidad, paginación, mapeo) queda en funciones testeadas sin base; verificar con `go build ./...` y `go test ./internal/modules/conversations/...`.

## 4. Lectura de contactos y usuarios

- [x] 4.1 Implementar `GET /v1/contacts/{id}` en `internal/modules/contacts/infra/http` reutilizando `FindContactByID`, con 400 para `id` inválido y 404 si no existe; verificar con tests `httptest`.
- [x] 4.2 Implementar `GET /v1/users/{id}` en `internal/modules/users/infra/http` reutilizando `FindUserByID`, con 400 para `id` inválido y 404 si no existe; verificar con tests `httptest`.

## 5. Wiring e integración

- [x] 5.1 Cablear los handlers de lectura en `cmd/api/app.go` con `db.Queries` y registrarlos bajo el middleware `Auth` en `cmd/api/router.go`; verificar que `go build ./...` compila.
- [x] 5.2 Agregar un test en `cmd/api/router_test.go` que confirme que las cuatro rutas de lectura responden 401 sin `Authorization` y no responden 404 (están registradas); verificar con `go test ./cmd/api/...`.
- [x] 5.3 Correr los quality gates del proyecto sobre todo el repo: `go mod verify && gofmt -l . && go vet ./... && go build ./... && go test -race -count=1 ./...`; verificar que todos pasan limpios.
