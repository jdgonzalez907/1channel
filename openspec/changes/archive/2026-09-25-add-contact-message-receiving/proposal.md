# Proposal

## Why

El modulo de conversaciones actual permite crear conversaciones y mensajes de forma estatica, pero no tiene la capacidad de recibir mensajes de contactos. Sin esto, no existe el flujo basico de un CRM conversacional: un contacto envia un mensaje a traves de una plataforma (WhatsApp, SMS, etc.) y el sistema lo recibe, valida y agrega a la conversacion correspondiente.

## What Changes

- **Message: campo externalID** — Agregar un campo `externalID *string` al Message para almacenar el identificador que asigna la plataforma externa (ej: WhatsApp message ID). Es nullable porque no todos los mensajes provienen de una plataforma, y es inmutable una vez asignado (set-once via `AssignExternalID`).
- **Message: AssignExternalID** — Metodo de dominio para asignar el identificador externo a un mensaje que aun no tiene uno. Rechaza si ya fue asignado o si el valor es vacio.
- **Conversation: invariante de mensajes minimos** — Una conversacion NUNCA puede tener 0 mensajes. Se valida en el constructor `NewConversation`. Toda conversacion nace con al menos un mensaje.
- **Conversation: ReceiveContactMessage** — Metodo de dominio para recibir un mensaje de un contacto. Valida que el contacto pertenezca a la conversacion y que el mensaje no este duplicado (por externalID). Agrega el mensaje al mapa de mensajes y actualiza el timestamp de la conversacion.
- **Conversation: indice de mensajes externos** — Mapa auxiliar `externalMsgIdx map[string]uuid.UUID` para busquedas O(1) por externalID, manteniendo `found` como unica fuente de verdad.

## Capabilities

### New Capabilities

_(ninguna)_

### Modified Capabilities

- `conversation`: Nuevos requerimientos para recibir mensajes de contactos, campo externalID en mensajes, e invariante de mensajes minimos en conversaciones.

## Impact

- `internal/modules/conversations/domain/message.go` — Nuevo campo, nuevo error, nuevo metodo
- `internal/modules/conversations/domain/conversation.go` — Nuevos errores, nuevo metodo, modificacion del constructor, nuevo indice interno
- `internal/modules/conversations/domain/*_test.go` — Tests actualizados y nuevos
- Sin cambios en infra, app, o API (solo dominio por ahora)
