# Proposal

## Why

El modulo de conversaciones soporta actualmente solo mensajes entrantes de contactos (`ReceiveContactMessage`). Para completar el flujo bidireccional del CRM, necesitamos que los agentes puedan enviar mensajes a las conversaciones. Esto es esencial para soporte, ventas y otras interacciones donde el agente responde al contacto.

Además, existe un bug de diseño: `Message` es una entidad pero se pasa por valor (`Message`) en vez de por puntero (`*Message`), lo que rompe la identidad del objeto según DDD.

## What Changes

- **Nuevo metodo `SendAgentMessage`**: Permite al agente enviar mensajes a una conversacion. Incluye logica de "claim" - si la conversacion esta pending sin agente, el agente que envia el primer mensaje se asigna automaticamente.
- **Nuevo metodo `ensureAcceptsMessages`**: Validacion compartida que verifica si la conversacion acepta mensajes en un momento dado (estado + finishedAt).
- **Refactor `*Message`**: Cambiar todos los usos de `Message` a `*Message` para respetar la identidad de la entidad.
- **Invariante en `NewConversation`**: Validar que conversaciones expired/resolved tengan finishedAt.
- **Update `ReceiveContactMessage`**: Agregar validacion de `ensureAcceptsMessages` como primer paso.

## Capabilities

### Modified Capabilities

- `conversation`: Agregar requerimiento para envio de mensajes por agentes, validacion de estado de conversacion, y correccion del patron de entidad con puntero.

### New Capabilities

Ninguna - el comportamiento se agrega al capability existente de conversation.

## Impact

- `internal/modules/conversations/domain/conversation.go` - Cambios principales
- `internal/modules/conversations/domain/conversation_test.go` - Tests actualizados y nuevos
- `internal/modules/conversations/domain/message.go` - Sin cambios (ya usa punteros internamente)
- Todos los tests que crean `Message` deben cambiar a `*Message`