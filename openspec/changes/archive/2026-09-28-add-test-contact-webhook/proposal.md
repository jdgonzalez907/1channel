# Proposal

## Why

El módulo de conversaciones ya implementa y testea los casos de uso de recibir, editar y eliminar mensajes de contacto (`app.ReceiveContactMessage`, `...Edit`, `...Delete`), pero no existe ningún borde HTTP que los active: los use cases están declarados en `ConversationsAPI` y nunca se cablean. Sin un endpoint de entrada no se puede ejercitar el flujo contacto -> sistema en pruebas locales ni preparar el terreno para los adaptadores de plataformas reales. Falta además la operación simétrica de acuse de lectura del contacto sobre los mensajes del agente.

## What Changes

- Nuevo endpoint abierto `POST /v1/webhooks/1channel` (sin `Authorization`) que recibe un envelope JSON con `event`, `external_contact_id` y `external_message_id` obligatorios y `text` según el evento; el timestamp lo genera el servidor.
- Cuatro eventos mapeados a los casos de uso de contacto:
  - `message.received` -> `ReceiveContactMessage`
  - `message.edited` -> `ReceiveContactMessageEdit`
  - `message.deleted` -> `ReceiveContactMessageDelete`
  - `message.read` -> `ReceiveContactMessageRead` (nuevo)
- Nuevo caso de uso `ReceiveContactMessageRead`: el contacto marca como leído un mensaje del agente, localizándolo por su identificador externo.
- Nuevo método de aggregate `Conversation.ReceiveContactMessageRead`, espejo de `ReceiveContactMessageEdit`; **no** rechaza por conversación finalizada porque los eventos de canal son asíncronos y deben ejecutarse igual.
- `Message.MarkAsRead` pasa a ser no-retrocedente (como `EditText`), no reactiva un mensaje `deleted` (conserva el estado) y no modifica uno `failed`.
- Un único DTO compartido (`ContactWebhookRequest`) con `event`, `external_contact_id` y `external_message_id` obligatorios y `text` según el evento; la validación de campos requeridos depende del evento.
- El timestamp (`at`) lo calcula el webhook en el servidor en UTC; no viaja en el cuerpo.
- Códigos: `204` en éxito, `422` por campo requerido faltante, `400` para JSON malformado, `200` sin efecto para un `event` no soportado, reutilizando el mapeo de errores de dominio del handler de escritura.
- Cableado de los cuatro use cases en `cmd/api` y registro de la ruta fuera del grupo autenticado.
- Sin variables de entorno, sin configuración ni feature flags: la ruta queda siempre activa. A futuro se decide si se conserva, se oculta en producción o se elimina.
- La ruta queda preparada para crecer a `/v1/webhooks/meta`, `/v1/webhooks/telegram`, etc., que resolverán los mismos use cases.

## Capabilities

### New Capabilities

_(ninguna)_

### Modified Capabilities

- `http-api`: se agrega el requerimiento de un webhook de contacto de prueba, abierto y sin autenticación, que expone las operaciones de recepción, edición, eliminación y acuse de lectura de mensajes del contacto.
- `conversation`: se agrega el caso de uso de marcar como leído un mensaje del agente a partir de un acuse de lectura del contacto.

## Impact

- `backend/internal/modules/conversations/domain/` — nuevo método de aggregate y ajuste de `MarkAsRead`.
- `backend/internal/modules/conversations/app/` — nuevo use case `receive_contact_message_read` con mock y tests.
- `backend/internal/modules/conversations/api.go` — nueva operación en `ConversationsAPI`.
- `backend/internal/modules/conversations/infra/http/` — nuevo handler y DTO del webhook.
- `backend/cmd/api/app.go` y `backend/cmd/api/router.go` — construcción de use cases y registro de la ruta.
- `backend/docs/endpoints.md` — documentar el endpoint y sus eventos.
