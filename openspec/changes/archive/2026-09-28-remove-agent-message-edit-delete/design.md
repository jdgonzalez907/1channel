# Design

## Context

Ver `proposal.md - Why`. El agregado `Conversation` expone `AgentEditMessage` (`conversation.go:293`) y `AgentDeleteMessage` (`conversation.go:320`), usados solo por sus tests (`conversation_test.go:616` y `:710`). No hay use case ni endpoint que los invoque. La spec de `conversation` lista escenarios de edición/borrado por agente dentro de los requirements `Editar mensaje de texto` y `Eliminar mensaje`.

`Message.EditText` y `Message.Delete` son la primitiva de mensaje y los usan los use cases del **contacto** (`receive_contact_message_edit.go`, `receive_contact_message_delete.go`), así que se conservan.

## Goals / Non-Goals

**Goals:**
- Quitar del agregado la capacidad de editar/borrar mensajes por parte del agente.
- Dejar la spec coherente: edición/borrado solo originados por el contacto.

**Non-Goals:**
- No se toca el flujo del contacto (`Message.EditText`/`Delete`, use cases del contacto).
- No se cambia el contrato HTTP (no existía endpoint de agente para esto).
- No se toca el read model ni `unread_count`.

## Decisions

### 1. Eliminar los métodos del agregado, no solo no exponerlos

Como no hay ningún consumidor, se borran `AgentEditMessage` y `AgentDeleteMessage` del agregado en lugar de dejarlos inertes. Evita que alguien los cablee por accidente y refleja la decisión de producto.

### 2. Conservar la primitiva de mensaje

`Message.EditText`/`Delete` siguen siendo necesarios para el flujo del contacto; el cambio es solo sobre el agregado (ownership del agente en conversación).

### 3. Forma del delta: REMOVED de los requirements generales + MODIFIED de los casos de uso

El validador de OpenSpec no permite que un requirement `MODIFIED` pierda escenarios. Como los escenarios de agente viven en los requirements generales `Editar mensaje de texto` y `Eliminar mensaje`, se **eliminan** esos requirements completos y las reglas de validación no-agente se **mueven** a los requirements de caso de uso del contacto (`Caso de uso: editar mensaje de contacto` / `... eliminar ...`), que son su lugar natural ahora que solo el contacto edita/elimina.

### 4. Sin cambios de contrato ni de persistencia

No hay endpoint, use case, query ni migración afectados.

## Risks / Trade-offs

- **Referencias ocultas** -> Antes de cerrar, `grep` de `AgentEditMessage`/`AgentDeleteMessage` en `backend` debe quedar vacío fuera de nada; si apareciera una, se detiene.
- **Pérdida de capacidad futura** -> Reintroducir edición/borrado de agente requeriría un change nuevo; es intencional.
