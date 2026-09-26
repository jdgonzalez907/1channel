# Design

## Context

Ver proposal.md - Why. El esquema ya existe en `db/migrations/`; `sqlc.yaml` apunta a `db/queries/` (vacío) y genera a `internal/postgres/sqlc`. No hay `cmd/api/main.go`, no hay pgx en `go.mod`, y `db/queries/` está vacío. Los identificadores son el paquete stdlib `uuid` de Go 1.27 (`type UUID [16]byte`, `NewV7`). Los repositorios de dominio y sus casos de uso ya existen; `Save` no es idempotente ni transaccional porque no hay implementación.

`Conversation` guarda `dirty map[uuid.UUID]*Message` privado, sin getter. `NewConversation` y `RehydrateConversation` comparten constructor y dejan `dirty` vacío.

## Goals / Non-Goals

**Goals:**

- Pool pgx y una única pieza compartida (`pgdb`) con DB, transacción y conversión de uuid.
- Queries sqlc y las implementaciones de los tres repositorios, con constructor que devuelve la interfaz y struct privado.
- `Save` transaccional y con upserts, escribiendo solo los mensajes modificados.

**Non-Goals:**

- Armar `cmd/api/main.go` ni el grafo de casos de uso y APIs.
- Tests (integración o unitarios de repositorio).
- Ejecutar migraciones o gestionar su aplicación.
- Cambiar el contrato de los repositorios o el comportamiento de los casos de uso.

## Decisions

### Estructura de paquetes

```
internal/shared/infra/pgdb/        package pgdb   DB, New, InTx, Close, conversión uuid
internal/shared/infra/pgdb/sqlc/   package sqlc   generado por sqlc
internal/modules/<m>/infra/pg/     package pg     repos (impl privada, constructor -> interfaz)
db/queries/{agents,contacts,conversations}/*.sql
```

El generado va en su propio subpaquete para que `sqlc generate` no conviva con código a mano en el mismo paquete. La carpeta se llama como el paquete (`pgdb`, `pg`). `pg -> pgdb -> pgdb/sqlc`; sin ciclos.

sqlc resuelve el campo `queries` con `filepath.Glob` y, para un directorio, solo lee su nivel inmediato (no recursivo). Por eso no sirve `db/queries/` (solo vería subdirectorios) y se declara `queries: "db/queries/*/*.sql"`.

### `pgdb.DB` y pool tras interfaz

```go
type Pool interface {
    Exec(ctx, sql, args...) (pgconn.CommandTag, error)
    Query(ctx, sql, args...) (pgx.Rows, error)
    QueryRow(ctx, sql, args...) pgx.Row
    Begin(ctx) (pgx.Tx, error)
    Ping(ctx) error
    Close()
}

type DB struct {
    Pool    Pool
    Queries *sqlc.Queries
}
```

`*pgxpool.Pool` cumple `Pool`. Guardar una interfaz (en vez de `*pgxpool.Pool` concreto) deja la puerta abierta a un mock sin base y evita acoplarse a pgxpool. `New(ctx, dsn)` hace `ParseConfig` + `Ping`; quien lee `POSTGRES_URL` es el `main.go` de otro cambio.

`InTx`:
```go
func (db *DB) InTx(ctx context.Context, fn func(q *sqlc.Queries) error) error
```
Abre `tx`, ejecuta `fn(db.Queries.WithTx(tx))`, hace `Rollback` en el `defer` y `Commit` al final.

### Conversión de tipos

El códec UUID de pgx solo codifica `UUIDValuer` y decodifica en `UUIDScanner`; `uuid.UUID` (`[16]byte`) no implementa esas interfaces. Los helpers viven en `pgdb/uuid.go`:

```go
func UUID(u uuid.UUID) pgtype.UUID
func UUIDPtr(u *uuid.UUID) pgtype.UUID
func FromUUID(p pgtype.UUID) uuid.UUID
func FromUUIDPtr(p pgtype.UUID) *uuid.UUID
```

`emit_pointers_for_null_types: true` hace que `text` salga `*string`, pero `timestamptz` sigue saliendo `pgtype.Timestamptz` (no `*time.Time`). Por eso las fechas también necesitan helpers, en `pgdb/time.go`:

```go
func Timestamp(t time.Time) pgtype.Timestamptz
func TimestampPtr(t *time.Time) pgtype.Timestamptz
func FromTimestamp(p pgtype.Timestamptz) time.Time
func FromTimestampPtr(p pgtype.Timestamptz) *time.Time
```

`sqlc.yaml` mantiene el override `uuid -> pgtype.UUID`.

### Seguimiento de modificados

`dirty` ya existe y los métodos de mutación lo llenan. Faltan dos cosas:
- getter `DirtyMessages() []*Message` (infra está en otro paquete);
- que `NewConversation` siembre `dirty` con sus mensajes iniciales; `RehydrateConversation` no.

Se extrae `compareMessages` (usado por `Messages()` y `DirtyMessages()`) para mantener el mismo orden determinista por `sentAt` y `id`.

### `Save`

```go
func (r *conversationRepository) Save(ctx context.Context, conv *domain.Conversation) error {
    return r.db.InTx(ctx, func(q *sqlc.Queries) error {
        if err := q.UpsertConversation(ctx, toUpsertConversationParams(conv)); err != nil {
            return err
        }
        for _, m := range conv.DirtyMessages() {
            if err := q.UpsertMessage(ctx, toUpsertMessageParams(m, conv.ID())); err != nil {
                return err
            }
        }
        return nil
    })
}
```

La fila de conversación se upserta siempre porque `status`, `agent_id`, `updated_at` y `finished_at` no viven en `dirty`. Los mensajes van por bucle, no por batch: `dirty` suele traer uno; `AgentReadConversation` puede traer varios, pero el umbral donde un `pgx.Batch` compensa (~50) está lejos del caso normal. Se reconsiderará si aparece.

### Finders

| Finder | SQL |
|---|---|
| `FindWithoutMessages` | conversación por id |
| `FindOpenWithMessageExternalIDsByContactID` | conversación `status IN ('pending','assigned')` por contacto + mensajes con `external_id IS NOT NULL` |
| `FindWithContactUnreadMessagesByID` | conversación por id + mensajes `contact_id IS NOT NULL AND read_at IS NULL` |
| `FindWithMessageByExternalID` | mensaje por `external_id` -> su `conversation_id` -> conversación + ese mensaje |

El rename de `FindOpenByContactID` a `FindOpenWithMessageExternalIDsByContactID` hace explícito que el `found` es parcial. Es el dato que `ReceiveContactMessage` necesita para detectar duplicados de forma idempotente antes de `Save`.

### Not found y errores

Toda lectura que no encuentra fila devuelve `(nil, nil)` (`pgx.ErrNoRows` traducido), porque los casos de uso ya traducen ausencia a `Err...NotFound`. Los errores de Postgres (incluido `23505`) se propagan crudos; los duplicados que el dominio detecta se filtran antes de `Save`.

## Risks / Trade-offs

- [Carrera en get-or-create de contacto] dos altas concurrentes del mismo `external_contact_id` -> `23505` crudo; el reintento del webhook lo resuelve. Mapearlo es un cambio posterior.
- [Carrera de conversación abierta] el `UNIQUE` parcial rechaza la segunda; el repositorio no reintenta. Queda fuera.
- [`found` parcial en el finder activo] una reutilización futura que itere `found` obtendría un subconjunto; mitigado por el nombre explícito.
- [`dirty` no se limpia tras `Save`] un segundo `Save` del mismo agregado reescribe los mismos mensajes; los upserts lo hacen idempotente.
- [Sin tests] nada valida el SQL ni las restricciones en este cambio.

## Open Questions

Ninguna que cambie el alcance de este cambio.
