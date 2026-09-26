# Design

## Context

Ver proposal.md - Why. Hoy `agents/infra/pg` y `contacts/infra/pg` leen pasando por `NewAgent`/`NewContact`. `NewContact` valida que el identificador externo no sea vacío y devuelve error; `NewAgent` devuelve error aunque no valide nada. `Conversation` y `Message` ya exponen `RehydrateConversation`/`RehydrateMessage`, que no validan invariantes.

`conversations/infra/pg` convierte `status`/`type` llamando a `NewConversationStatus`, `NewMessageStatus` y `NewMessageType`, que validan contra el conjunto permitido aunque las columnas tengan `CHECK` en la base.

## Goals / Non-Goals

**Goals:**

- Rehidratación de contacto y agente sin validar invariantes, análoga a la de conversación y mensaje.
- Convención de nombres `postgres<Entidad>Repository` en la implementación.
- Lecturas que no revalidan lo que la base ya restringe.

**Non-Goals:**

- Cambiar el comportamiento de los casos de uso.
- Agregar repositorios, queries o esquema.
- Tests de repositorio.

## Decisions

### Rehidratación separada de la creación

```go
func RehydrateContact(id uuid.UUID, externalContactID string, createdAt time.Time) *Contact
func RehydrateAgent(id uuid.UUID, createdAt time.Time) *Agent
```

No validan y no devuelven error. Se conservan `NewContact` y `NewAgent` para la creación, donde la validación sí aplica. Es el mismo patrón que `RehydrateConversation`/`RehydrateMessage`: obtener una entidad desde datos persistidos es distinto de crearla.

### Nombres de implementación

Los structs pasan a `postgresAgentRepository`, `postgresContactRepository` y `postgresConversationRepository`. Los constructores (`NewAgentRepository`, etc.) siguen devolviendo la interfaz del dominio.

### Cast directo en lecturas

En las lecturas, `status` y `type` se convierten con cast directo:

```go
domain.ConversationStatus(row.Status)
domain.MessageStatus(row.Status)
domain.MessageType(row.Type)
```

Justificación: rehidratar no valida (los `CHECK` de la base garantizan los valores), y los constructores `New*` existen para la creación, no para leer. Los mappers (`toConversation`, `toMessage`, `toMessages`, `toContact`) dejan de devolver error.

## Risks / Trade-offs

- [Un valor inválido persistido se castearía en silencio] mitigado por los `CHECK` de `messages.status`, `messages.type` y `conversations.status`; si la base se saltara el check, el dominio recibiría un valor fuera del conjunto.
- [Se conservan `New*`] siguen siendo la vía de creación validada; no se eliminan.

## Open Questions

Ninguna que cambie el alcance de este cambio.
