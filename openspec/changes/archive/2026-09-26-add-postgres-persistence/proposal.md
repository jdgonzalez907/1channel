# Proposal

## Why

Existen el esquema PostgreSQL, los repositorios definidos y sus casos de uso, pero ninguna implementación: no hay pool pgx, no hay queries y no hay forma de persistir ni leer el agregado. Sin esta capa, los casos de uso de conversaciones y contactos no tocan la base.

## What Changes

- Nuevo paquete compartido `internal/shared/infra/pgdb` (package `pgdb`): pool pgx, `DB`, transacción reutilizable y conversión `uuid.UUID <-> pgtype.UUID` y `time.Time <-> pgtype.Timestamptz`.
- Código generado por sqlc en `internal/shared/infra/pgdb/sqlc` (package `sqlc`); queries en `db/queries/{agents,contacts,conversations}/`.
- Implementaciones de repositorio en `internal/modules/<módulo>/infra/pg` (package `pg`), con struct privado y constructor que devuelve la interfaz; reciben `*pgdb.DB`.
- Dominio: `Conversation.DirtyMessages()` y helper `compareMessages`; `NewConversation` marca sus mensajes iniciales como modificados y `RehydrateConversation` no.
- Rename del finder de conversación activa a `FindOpenWithMessageExternalIDsByContactID`, que carga solo los mensajes con `external_id`.
- `Save` en una transacción: upsert de la conversación siempre, y upsert por bucle de los mensajes modificados en orden determinista.
- `sqlc.yaml`: el generado pasa al subpaquete, `package: sqlc`, y `emit_pointers_for_null_types: true`.
- Fuera de alcance: wiring en `cmd/api`, tests y ejecución de migraciones.

## Capabilities

### New Capabilities

- Ninguna.

### Modified Capabilities

- `conversation`: se aclara que la búsqueda de la conversación activa incluye los identificadores externos de sus mensajes, y se agrega el requerimiento de seguimiento de mensajes modificados (creación marca, rehidratación no).
- `persistence`: se agrega el comportamiento del repositorio: lecturas, upserts de conversación/mensajes en transacción, persistencia de contactos y agentes, y propagación de conflictos de unicidad.

## Impact

- Nuevas dependencias: `github.com/jackc/pgx/v5` (pool y tipos). `sqlc` v1.31.1 se usa para generar; `db/queries/` deja de estar vacío.
- Nuevos paquetes y directorios: `internal/shared/infra/pgdb/`, `internal/shared/infra/pgdb/sqlc/`, `internal/modules/*/infra/pg/`.
- Cambios de dominio en `internal/modules/conversations/domain/` (getter, siembra, comparator) y rename en la interfaz de repositorio, su mock y `receive_contact_message.go`.
- No se modifica `cmd/api` ni se ejecutan migraciones; el grafo de la aplicación se arma en otro cambio.
