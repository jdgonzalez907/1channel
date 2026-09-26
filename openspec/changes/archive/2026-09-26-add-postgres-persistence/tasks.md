# Tasks

## 1. Dependencias y sqlc

- [x] 1.1 Agregar `github.com/jackc/pgx/v5` a `go.mod` con `go get github.com/jackc/pgx/v5@v5.9.2` y verificar con `go mod verify` que no hay errores.
- [x] 1.2 Instalar sqlc con `go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1` y verificar que `sqlc version` responde.
- [x] 1.3 Actualizar `sqlc.yaml` a `package: "sqlc"`, `out: "internal/shared/infra/pgdb/sqlc"` y agregar `emit_pointers_for_null_types: true`, manteniendo el override de `uuid` a `pgtype.UUID`. Verificar que el archivo contiene esas tres claves y sigue siendo YAML válido.

## 2. Queries y generado

- [x] 2.1 Crear `db/queries/agents/agents.sql` con `FindAgentByID` seleccionando `id, created_at` por `id`. Verificar que el archivo existe.
- [x] 2.2 Crear `db/queries/contacts/contacts.sql` con `FindContactByID`, `FindContactByExternalContactID` y `UpsertContact` (`INSERT ... ON CONFLICT (id) DO UPDATE`). Verificar que el archivo existe.
- [x] 2.3 Crear `db/queries/conversations/conversations.sql` con `FindConversationWithoutMessages`, `FindOpenConversationWithMessageExternalIDsByContactID`, `FindConversationWithContactUnreadMessagesByID` y `UpsertConversation` (`ON CONFLICT (id) DO UPDATE`). Verificar que el archivo existe.
- [x] 2.4 Crear `db/queries/conversations/messages.sql` con la búsqueda del mensaje por `external_id`, la carga de mensajes de una conversación, y `UpsertMessage` (`ON CONFLICT (id) DO UPDATE`). Verificar que el archivo existe.
- [x] 2.5 Ejecutar `sqlc generate` y verificar que aparecen archivos en `internal/shared/infra/pgdb/sqlc/` y que `go build ./internal/shared/infra/pgdb/sqlc/...` compila.

## 3. Dominio

- [x] 3.1 Extraer el comparador a `compareMessages(a, b *Message) bool` y usarlo en `Messages()`. Verificar que `go test ./internal/modules/conversations/domain/...` sigue pasando.
- [x] 3.2 Agregar `DirtyMessages() []*Message` que devuelve un slice nuevo ordenado con `compareMessages`. Verificar que compila.
- [x] 3.3 Hacer que `NewConversation` marque sus mensajes iniciales en `dirty`, sin tocar `RehydrateConversation`. Verificar que compila y que `go test ./internal/modules/conversations/domain/...` pasa.
- [x] 3.4 Renombrar `FindOpenByContactID` a `FindOpenWithMessageExternalIDsByContactID` en la interfaz `ConversationRepository`, su mock, `receive_contact_message.go` y los tests que lo referencian. Verificar que `go test ./internal/modules/conversations/...` pasa.

## 4. Paquete pgdb

- [x] 4.1 Crear `internal/shared/infra/pgdb/pgdb.go` (package `pgdb`) con la interfaz `Pool`, el struct `DB` con `Pool` y `Queries`, `New(ctx, dsn)` con `ParseConfig` + `Ping`, y `Close`. Verificar que `go build ./internal/shared/infra/pgdb/...` compila.
- [x] 4.2 Agregar `InTx(ctx, fn func(q *sqlc.Queries) error) error` con `Begin`, `Rollback` diferido y `Commit`. Verificar que compila.
- [x] 4.3 Crear `internal/shared/infra/pgdb/uuid.go` con `UUID`, `UUIDPtr`, `FromUUID` y `FromUUIDPtr`. Verificar que compila.
- [x] 4.4 Crear `internal/shared/infra/pgdb/time.go` con `Timestamp`, `TimestampPtr`, `FromTimestamp` y `FromTimestampPtr` (necesarios porque `timestamptz` se genera como `pgtype.Timestamptz`). Verificar que compila.

## 5. Implementaciones de repositorio

- [x] 5.1 Crear `internal/modules/agents/infra/pg/agent_repository.go` (package `pg`) con struct privado, `NewAgentRepository(db *pgdb.DB) domain.AgentRepository` y `FindByID` con `(nil, nil)` ante ausencia. Verificar que compila.
- [x] 5.2 Crear `internal/modules/contacts/infra/pg/contact_repository.go` con `FindByID`, `FindByExternalContactID` y `Save` con upsert. Verificar que compila.
- [x] 5.3 Crear `internal/modules/conversations/infra/pg/conversation_repository.go` con los cuatro finders, mapeando filas a `RehydrateConversation`/`RehydrateMessage` y devolviendo `(nil, nil)` ante ausencia. Verificar que compila.
- [x] 5.4 Implementar `Save` en el repositorio de conversaciones: `db.InTx` con upsert de la conversación y bucle de `UpsertMessage` sobre `DirtyMessages()` usando `conversation.ID()` como `conversation_id`. Verificar que compila.

## 6. Verificación

- [x] 6.1 Ejecutar `go mod tidy` y verificar que `go.mod` y `go.sum` quedan consistentes.
- [x] 6.2 Ejecutar `gofmt -l .` sin archivos listados y `go vet ./...` sin errores.
- [x] 6.3 Ejecutar `go build ./...` y `go test -race -count=1 ./...` con todo pasando.
