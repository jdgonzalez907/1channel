# Proposal

## Why

El módulo `conversations` solo tiene su dominio: no existe la capa de casos de uso que orqueste la entrada/salida de mensajes y el ciclo de vida de la conversación, ni los módulos `agents` y `contacts` con su API pública. Sin esa capa las conversaciones no se persisten ni se sincronizan con el canal.

## What Changes

- **Nuevos módulos** `agents` y `contacts`, cada uno con su interfaz `ModuleNameAPI` en la raíz (solo los métodos que necesitan los casos de uso de conversaciones).
- **7 casos de uso** en `internal/modules/conversations/app/`: `ReceiveContactMessage`, `ReceiveContactMessageEdit`, `ReceiveContactMessageDelete`, `SendAgentMessage`, `AgentReadConversation`, `ExpireConversation`, `ResolveConversation`.
- **Convenciones de casos de uso**: interface con un solo `Execute(ctx, input)`, input `<Caso>Input`, implementación privada, constructor que devuelve la interface, error base `Err<Gerundio>` por caso, método `joinErr(errs ...error)` y silenciado inline de errores idempotentes con `errors.Is`.
- **Finders dedicados** en `ConversationRepository`: `FindWithoutMessages`, `FindWithContactUnreadMessagesByID`, `FindWithMessageByExternalID`, `FindOpenByContactID`.
- **Puerto outbound** `AgentMessageSender` (solo `Send`) en el dominio de conversaciones.
- **Cambios de dominio**: `NewMessage` recibe `externalID *string`; rehidratación 0..n (`RehydrateConversation`/`RehydrateMessage`); método `AssignAgentMessageExternalID`; error `ErrConversationNotFound`.
- **Fuera del MVP** (los métodos de dominio se conservan): casos de uso `AgentEditMessage`, `AgentDeleteMessage`, `MarkAgentMessageFailed`.

## Capabilities

### New Capabilities

- `agents`: API pública del módulo de agentes (`AgentsAPI`), usada por los casos de uso de conversaciones para validar la existencia del agente.
- `contacts`: API pública del módulo de contactos (`ContactsAPI`), usada para resolver/crear el contacto por su identificador externo y para obtener el identificador externo de destino.

### Modified Capabilities

- `conversation`: se agregan los requerimientos de **casos de uso** (orquestación, idempotencia, errores base, integración con `agents`/`contacts`, puerto outbound) y se modifican requerimientos de dominio (`NewMessage` con `externalID`, rehidratación 0..n, asignación de `externalID` de agente, conversación no encontrada).

## Impact

- Nuevos directorios: `internal/modules/{agents,contacts}/` y `internal/modules/conversations/app/`.
- Modificaciones a `internal/modules/conversations/domain/` (`message.go`, `conversation.go`, `conversation_repository.go`) y nuevo `agent_message_sender.go`.
- Dependencias entre módulos mediante interfaces (`AgentsAPI`, `ContactsAPI`, `AgentMessageSender`); sin acoplamiento directo entre implementaciones.
- Sin cambios en esquema de base de datos en este cambio (la persistencia es un cambio posterior).
