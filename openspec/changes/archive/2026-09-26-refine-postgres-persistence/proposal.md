# Proposal

## Why

Las lecturas de los repositorios PostgreSQL pasan por constructores que validan: `NewContact` rechaza un identificador externo vacío y devuelve error, y `NewAgent` devuelve error aunque nunca falle. La rehidratación desde datos persistidos no debe validar invariantes ni confundirse con la creación, como ya ocurre en `Conversation` y `Message` con sus `Rehydrate*`. Además, la implementación de repositorio no sigue la convención de nombres y las lecturas revalidan estados y tipos que la base ya restringe con `CHECK`.

## What Changes

- Dominio: agregar `RehydrateContact` y `RehydrateAgent`, que reconstruyen la entidad sin validar invariantes. Se conservan `NewContact` y `NewAgent` para la creación.
- Renombrar los structs de implementación a `postgresAgentRepository`, `postgresContactRepository` y `postgresConversationRepository`. Los constructores siguen devolviendo la interfaz.
- Lecturas: rehidratar con `Rehydrate*` y convertir `status`/`type` con cast directo (`domain.ConversationStatus`, `domain.MessageStatus`, `domain.MessageType`), sin revalidar. Los mappers dejan de devolver error.
- Fuera de alcance: comportamiento de casos de uso, repositorios nuevos y tests de repositorio.

## Capabilities

### New Capabilities

- Ninguna.

### Modified Capabilities

- `contacts`: se agrega el requerimiento de rehidratar un contacto sin validar invariantes.
- `agents`: se agrega el requerimiento de rehidratar un agente sin validar invariantes.

## Impact

- `internal/modules/contacts/domain/contact.go` y `internal/modules/agents/domain/agent.go`.
- `internal/modules/{agents,contacts,conversations}/infra/pg/`.
- Sin cambios de esquema, queries, `sqlc.yaml` ni casos de uso.
