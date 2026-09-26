# Tasks

## 1. Rehidratación en dominio

- [x] 1.1 Agregar `RehydrateAgent(id uuid.UUID, createdAt time.Time) *Agent` en `internal/modules/agents/domain/agent.go`, sin validar. Verificar que compila.
- [x] 1.2 Agregar test de dominio para `RehydrateAgent` y verificar que `go test ./internal/modules/agents/domain/...` pasa.
- [x] 1.3 Agregar `RehydrateContact(id uuid.UUID, externalContactID string, createdAt time.Time) *Contact` en `internal/modules/contacts/domain/contact.go`, sin validar. Verificar que compila.
- [x] 1.4 Agregar test de dominio para `RehydrateContact`, incluyendo identificador externo vacío, y verificar que `go test ./internal/modules/contacts/domain/...` pasa.

## 2. Implementaciones de repositorio

- [x] 2.1 Renombrar el struct de `internal/modules/agents/infra/pg` a `postgresAgentRepository` y usar `RehydrateAgent` en `FindByID`. Verificar que compila.
- [x] 2.2 Renombrar el struct de `internal/modules/contacts/infra/pg` a `postgresContactRepository`, usar `RehydrateContact` en `toContact` y quitar su retorno de error (actualizando los llamadores). Verificar que compila.
- [x] 2.3 Renombrar el struct de `internal/modules/conversations/infra/pg` a `postgresConversationRepository`; usar `domain.ConversationStatus(row.Status)`, `domain.MessageStatus(row.Status)` y `domain.MessageType(row.Type)`, y quitar el retorno de error de `toConversation`, `toMessage` y `toMessages` (actualizando los llamadores). Verificar que compila.

## 3. Verificación

- [x] 3.1 Ejecutar `gofmt -l .` sin archivos listados y `go vet ./...` sin errores.
- [x] 3.2 Ejecutar `go build ./...` y `go test -race -count=1 ./...` con todo pasando.
