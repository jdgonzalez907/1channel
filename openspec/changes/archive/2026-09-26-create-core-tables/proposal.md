# Proposal

## Why

Los repositorios de agentes, contactos y conversaciones ya definen las búsquedas del MVP, pero no existe esquema en PostgreSQL. Sin tablas, claves e índices no se puede persistir ni resolver esas consultas sin barridos o filas huérfanas.

## What Changes

- Cuatro migraciones golang-migrate, una por tabla, en `db/migrations/`: `agents`, `contacts`, `conversations`, `messages`.
- Claves foráneas con `ON DELETE RESTRICT` y `ON UPDATE RESTRICT` entre conversaciones, mensajes, agentes y contactos.
- Índices que cubren las consultas actuales de los repositorios: unicidad de `external_contact_id`, una conversación abierta por contacto, `messages.conversation_id` y unicidad de `external_id` de mensaje.
- Identificadores `uuid` asignados por la aplicación (`uuid.NewV7()`), sin `DEFAULT` en la base.
- Sin `INSERT` de datos, sin repositorios, sin consultas SQLC y sin arrancar la librería de migrate en el proceso.

## Capabilities

### New Capabilities

- `persistence`: esquema PostgreSQL 18 de agentes, contactos, conversaciones y mensajes, con las restricciones e índices que las búsquedas del MVP exigen.

### Modified Capabilities

- Ninguna. Los requerimientos de `agents`, `contacts` y `conversation` no cambian; este cambio solo materializa su persistencia.

## Impact

- Nuevos archivos `db/migrations/000001_create_agents.{up,down}.sql` a `000004_create_messages.{up,down}.sql`.
- `sqlc.yaml` ya apunta `schema` a `db/migrations/`; este cambio no genera código ni añade queries.
- No modifica módulos de dominio ni casos de uso.
- Un envío de agente fallará por FK hasta que exista la fila en `agents`; el alta queda fuera de este cambio.
