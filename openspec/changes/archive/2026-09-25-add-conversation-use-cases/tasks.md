# Tasks

## 1. Módulo `agents`

- [x] 1.1 Crear `internal/modules/agents/domain/agent.go` con `Agent`, getters y `ErrAgentNotFound`; verificar con `go build ./...`
- [x] 1.2 Crear `internal/modules/agents/domain/agent_repository.go` con `AgentRepository.FindByID`; verificar con `go build ./...`
- [x] 1.3 Implementar `FindAgentByID` (interface `Execute`, input `FindAgentByIDInput`, impl privada, constructor que devuelve la interface, error base `ErrFindingAgentByID` y `joinErr` variádico) y verificar sus tests unitarios con repositorio mock
- [x] 1.4 Crear `internal/modules/agents/api.go` con `AgentsAPI.FindAgentByID` delegando al caso de uso y devolviendo `uuid.UUID`; verificar con tests de la API y `go build ./...`

## 2. Módulo `contacts`

- [x] 2.1 Crear `internal/modules/contacts/domain/contact.go` con `Contact`, getters y errores `ErrExternalContactIDEmpty`/`ErrContactNotFound`; verificar con `go build ./...`
- [x] 2.2 Crear `internal/modules/contacts/domain/contact_repository.go` con `ContactRepository.FindByID`, `FindByExternalContactID` y `Save`; verificar con `go build ./...`
- [x] 2.3 Implementar `GetOrCreateContactByExternalID` y verificar sus tests unitarios (existente, nuevo, id externo vacío)
- [x] 2.4 Implementar `FindContactByID` y verificar sus tests unitarios (existente, inexistente)
- [x] 2.5 Crear `internal/modules/contacts/api.go` con `ContactsAPI.GetOrCreateContactIDByExternalID` y `FindExternalContactIDByContactID`; verificar con tests de la API y `go build ./...`

## 3. Cambios de dominio en `conversations`

- [x] 3.1 Agregar `externalID *string` a `NewMessage` con validación de vacío no-nil (`ErrMessageExternalIDInvalid`); verificar que los tests de dominio existentes compilan/pasan y que los mensajes de agente pasan `nil`
- [x] 3.2 Agregar `ErrConversationNotFound` en `conversation.go`; verificar con `go build ./...`
- [x] 3.3 Agregar `RehydrateConversation` (rehidratación 0..n sin validar, sobre el constructor privado nil-safe `newConversation`); verificar con test de dominio que rehidrata una conversación sin mensajes y que reconstruye el índice externo
- [x] 3.4 Agregar `AssignAgentMessageExternalID(msgID, externalMessageID, at)` al aggregate; verificar tests de dominio (asignación, mensaje inexistente, actualización de `externalMsgIdx`)
- [x] 3.5 Crear `internal/modules/conversations/domain/agent_message_sender.go` con el puerto `AgentMessageSender.Send`; verificar con `go build ./...`
- [x] 3.6 Extender `domain.ConversationRepository` con los finders `FindWithoutMessages`, `FindWithContactUnreadMessagesByID`, `FindWithMessageByExternalID` y `FindOpenByContactID`; verificar con `go build ./...`
- [x] 3.7 Agregar `RehydrateMessage` (rehidratación directa sin validar, sobre el constructor privado `newMessage`); verificar con test de dominio que mapea los campos tal cual

## 4. Casos de uso de ciclo de vida

- [x] 4.1 Implementar `ExpireConversation` (finder `FindWithoutMessages`, silenciar `ErrConversationFinished`) y verificar tests: expira pending/assigned, idempotente si finalizada, `ErrConversationNotFound`, error de repo envuelto con `ErrExpiringConversation`
- [x] 4.2 Implementar `ResolveConversation` (validar agente con `AgentsAPI`, silenciar `ErrConversationFinished`, `ErrConversationAgentNotOwner` real) y verificar tests: resuelve, agente distinto/sin agente, idempotente, agente inexistente

## 5. Casos de uso de agente

- [x] 5.1 Implementar `SendAgentMessage` (validar agente, `FindWithoutMessages`, construir mensaje sin `externalID`, save 1, destino vía `ContactsAPI`, `AgentMessageSender.Send`, `AssignAgentMessageExternalID` + save 2; fallo → `MarkAgentMessageFailed` + save + error envuelto) y verificar tests: envío exitoso, reclamo de pending, agente no dueño, fallo de envío deja `failed` y retorna error
- [x] 5.2 Implementar `AgentReadConversation` (validar agente, `FindWithContactUnreadMessagesByID`, sin silenciables) y verificar tests: marca no leídos del contacto, sin no leídos no-op, no toca mensajes del agente, agente no asignado

## 6. Casos de uso de contacto

- [x] 6.1 Implementar `ReceiveContactMessage` (`ContactsAPI`, `FindOpenByContactID`, crear conversación nueva si no hay activa, silenciar `ErrConversationDuplicateMessage`) y verificar tests: adjunta a activa, crea nueva, duplicado idempotente, contacto nuevo, conversación de otro contacto
- [x] 6.2 Implementar `ReceiveContactMessageEdit` (`FindWithMessageByExternalID`, sin silenciables) y verificar tests: edición aplicada, stale sin error, `ErrMessageNotFound`, mensaje del agente
- [x] 6.3 Implementar `ReceiveContactMessageDelete` (silenciar `ErrMessageAlreadyDeleted`) y verificar tests: eliminación aplicada, repetida idempotente, `ErrMessageNotFound`, mensaje del agente

## 7. Frontera pública

- [x] 7.1 Crear `internal/modules/conversations/api.go` con `ConversationsAPI` que genera los IDs (`uuid.NewV7`) y delega en los 7 casos de uso; verificar con tests de la API

## 8. Calidad

- [x] 8.1 Ejecutar `gofmt -l . && go vet ./... && go build ./...` y corregir hallazgos
- [x] 8.2 Ejecutar `go test -race -count=1 ./...` y asegurar que toda la suite pasa
- [x] 8.3 Medir cobertura del código real (excluyendo `*_mock.go`): 100% en `domain`/`agents`/`contacts` y 98.5% en `conversations/app` (2 ramas defensivas inalcanzables, documentadas en `design.md`)
