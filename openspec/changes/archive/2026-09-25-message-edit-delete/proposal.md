# Proposal

## Why

El modulo de conversaciones soporta envío y recepción de mensajes, pero no permite editarlos ni eliminarlos una vez creados. Los agentes necesitan corregir errores en mensajes enviados, y los contactos pueden editar/eliminar mensajes desde WhatsApp. Además, falta el estado `failed` para mensajes que no se pudieron entregar.

## What Changes

- **Nuevo metodo `EditText` en Message**: Permite editar el texto de un mensaje. Solo aplica a mensajes de tipo texto. Valida que el nuevo texto no esté vacío ni exceda el máximo. Si llega una edición con timestamp anterior o igual a la última edición, conserva la más reciente. Actualiza `editedAt` y `text`.
- **Nuevo metodo `Delete` en Message**: Permite eliminar un mensaje (soft delete). Rechaza si el mensaje ya está eliminado o en estado failed. Actualiza `deletedAt` y cambia status a `deleted`.
- **Nuevo metodo `MarkAsFailed` en Message**: Marca como fallido un mensaje enviado por el agente. El estado failed es terminal y bloquea todas las modificaciones.
- **Nuevos helpers de validacion en Message**: `ensureNotFailed` y `ensureEditable` validan el estado del mensaje antes de modificarlo.
- **Nuevo metodo `AgentEditMessage` en Conversation**: Permite al agente editar mensajes por ID. Valida que la conversación no esté finalizada.
- **Nuevo metodo `AgentDeleteMessage` en Conversation**: Permite al agente eliminar mensajes por ID. Valida que la conversación no esté finalizada.
- **Nuevo metodo `MarkAgentMessageFailed` en Conversation**: Entry point para que la infraestructura reporte que un mensaje del agente no se pudo entregar. No valida estado de conversación.
- **Nuevo metodo `ReceiveContactMessageEdit` en Conversation**: Permite recibir edición de texto del contacto por externalID. No valida estado de conversación.
- **Nuevo metodo `ReceiveContactMessageDelete` en Conversation**: Permite recibir eliminación del contacto por externalID. No valida estado de conversación.
- **Nuevo metodo privado `registerActivity` en Conversation**: Actualiza `updatedAt` solo si el timestamp es más reciente. Reemplaza la asignación manual en todos los métodos.
- **Nuevo helper `ensureAgentCanModify` en Conversation**: Valida que la conversación no esté finalizada antes de una modificación del agente.
- **Nuevo estado `failed` en MessageStatus**: Estado terminal que bloquea todas las modificaciones. Modifica el requerimiento "Estado de mensaje".

## Capabilities

### Modified Capabilities

- `conversation`: Agregar requerimientos para edición y eliminación de mensajes por agente y contacto, estado failed, y timestamp de actualización monótono.

### New Capabilities

Ninguna - el comportamiento se agrega al capability existente de conversation.

## Impact

- `internal/modules/conversations/domain/message.go` - Nuevos métodos `EditText`, `Delete`, `MarkAsFailed`, `ensureNotFailed`, `ensureEditable`
- `internal/modules/conversations/domain/message_status.go` - Nuevo estado `failed` + `NewMessageStatus`
- `internal/modules/conversations/domain/conversation.go` - Nuevos métodos `AgentEditMessage`, `AgentDeleteMessage`, `MarkAgentMessageFailed`, `ReceiveContactMessageEdit`, `ReceiveContactMessageDelete`, `registerActivity`, `ensureAgentCanModify`
- `internal/modules/conversations/domain/message_test.go` - Tests nuevos
- `internal/modules/conversations/domain/conversation_test.go` - Tests nuevos y actualizados