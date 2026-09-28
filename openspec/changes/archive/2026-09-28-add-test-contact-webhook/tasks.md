# Tasks

## 1. DTO y validación del webhook

- [x] 1.1 Crear `contact_webhook_dto.go` en `conversations/infra/http` con `ContactWebhookRequest` (`Event`, `ExternalContactID` y `ExternalMessageID` obligatorios; `Text *string`; sin timestamp) y verificar con `go -C backend build ./...`
- [x] 1.2 Validar los campos requeridos por evento dentro del `switch` del handler (campo faltante -> 422; evento desconocido -> 200) y verificar con `go -C backend build ./...`
- [x] 1.3 Cubrir con tests el 422 por campo requerido faltante y el 200 para evento desconocido; verificar con `go -C backend test ./internal/modules/conversations/infra/http/...`

## 2. Mapeo de errores compartido

- [x] 2.1 Extraer `writeError` de `ConversationWriteHandler` a una función de paquete `writeConversationError(w, r, err)` en `conversations/infra/http` y hacer que `ConversationWriteHandler` la use; verificar que los tests existentes de `conversation_write_handler_test.go` siguen pasando con `go -C backend test ./internal/modules/conversations/infra/http/...`

## 3. Acuse de lectura del contacto (domain + app + API)

- [x] 3.1 Hacer `Message.MarkAsRead` no-retrocedente, sin reactivar un mensaje `deleted` (conserva el estado) ni modificar uno `failed`, y agregar tests de lectura aplicada, obsoleta, idempotente, sobre eliminado y sobre fallido; verificar con `go -C backend test ./internal/modules/conversations/domain/...`
- [x] 3.2 Agregar `Conversation.ReceiveContactMessageRead(contactID, externalID, at)` espejo de `ReceiveContactMessageEdit`, sin `ensureNotFinished`/`ensureAcceptsMessages`, y tests: lectura aplicada, acuse obsoleto, identificador externo inexistente, mensaje no del agente y lectura en conversación finalizada; verificar con `go -C backend test ./internal/modules/conversations/domain/...`
- [x] 3.3 Crear el use case `ReceiveContactMessageRead` (`app/receive_contact_message_read.go`) espejo de `receive_contact_message_delete.go`, con su mock y tests (éxito, conversación inexistente, error del repositorio); verificar con `go -C backend test ./internal/modules/conversations/app/...`
- [x] 3.4 Exponer `ReceiveContactMessageRead` en `ConversationsAPI` (`api.go`) con su input y delegación, y agregar su test en `api_test.go`; verificar con `go -C backend test ./internal/modules/conversations/...`

## 4. Handler HTTP del webhook

- [x] 4.1 Implementar `ContactWebhookHandler` (dependiente de los cuatro use cases de contacto) con `Register` sobre `POST /webhooks/1channel`, switch por evento, timestamp del servidor (`time.Now().UTC()`) y respuesta 204; verificar con `go -C backend build ./...`
- [x] 4.2 Usar `writeConversationError` para mapear los errores de dominio de los use cases; verificar con `go -C backend build ./...`
- [x] 4.3 Agregar tests con mocks testify: cada evento invoca el use case correcto con el input mapeado (`ExternalContactID`, `ExternalMessageID`, `Text`, timestamp UTC); campo faltante responde 422 y evento desconocido responde 200; JSON malformado responde 400; error de dominio se mapea; verificar con `go -C backend test ./internal/modules/conversations/infra/http/...`
- [x] 4.4 Documentar `POST /v1/webhooks/1channel` y sus cuatro eventos con ejemplos `curl` en `backend/docs/endpoints.md` y verificar que los comandos documentados corren tal cual contra la API local

## 5. Cableado y registro de la ruta

- [x] 5.1 Modificar `newConversationsModule` en `cmd/api/app.go` para construir los cuatro use cases de contacto y devolver también `*convhttp.ContactWebhookHandler`; verificar con `go -C backend build ./...`
- [x] 5.2 Registrar el webhook en `cmd/api/router.go` dentro de `/v1` pero fuera del grupo con `Auth`, y verificar con `go -C backend build ./...`
- [x] 5.3 Verificar el flujo end-to-end manualmente: `make up`, `make migrate-up`, `make run-api` y `curl` de los cuatro eventos contra `POST /v1/webhooks/1channel` (sin `Authorization`), confirmando 204 y el cambio reflejado en `GET /v1/conversations`

## 6. Quality gates

- [x] 6.1 Ejecutar `go -C backend mod verify && gofmt -l backend && go -C backend vet ./...` y verificar que `gofmt` no reporta archivos
- [x] 6.2 Ejecutar `go -C backend build ./...` y verificar que compila
- [x] 6.3 Ejecutar `go -C backend test -race -count=1 ./...` y verificar que todos los tests pasan
