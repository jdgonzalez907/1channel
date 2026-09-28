# Design

## Context

Ver `proposal.md` - Why. El estado actual que condiciona el enfoque:

- `domain` + `app` modelan únicamente escrituras; los DTOs viven en `infra/http` de cada módulo.
- SQLC genera un paquete compartido `internal/shared/infra/pgdb/sqlc` a partir de `db/queries/*/*.sql`, con tipos `pgtype`.
- Ya existen las queries `FindContactByID` y `FindUserByID`, y el índice `messages_conversation_id_idx`.
- `Message.Delete` hace soft delete (`status='deleted'`, `deleted_at`) pero **conserva el texto** en la base.
- `conversations` no tiene índices por `user_id`/`status` ni columnas de último mensaje/no leídos.
- Todas las mutaciones de conversación pasan por `ConversationRepository.Save` (una transacción por llamada), lo que permite mantener datos derivados de forma confiable.

## Goals / Non-Goals

**Goals:**

- Un camino de lectura paralelo: handler HTTP -> queries de solo lectura -> Response, sin tocar `domain`/`app`.
- Bandeja paginada por scroll (20) ordenada por el `sent_at` del último mensaje, con `status` obligatorio (`open|finished`) y filtro opcional por `external_contact_id`, leyendo solo O(página) sin recorrer el historial del agente.
- Detalle de conversación con metadata y últimos 20 mensajes paginando hacia atrás.
- `GET` de contacto y usuario por identificador reutilizando queries existentes.

**Non-Goals:**

- No se toca el contrato de escritura (`http-api`): ningún endpoint de escritura cambia.
- No se agregan índices de listado por `updated_at` ni un índice parcial de no leídos.

## Decisions

### 1. Las lecturas no pasan por `domain`/`app`/repositorios

El handler de lectura resuelve la query SQL y mapea a Response en `infra/http`. Se descarta reutilizar los repositorios de escritura (obligaría a meter modelos de lectura en el core) y crear servicios de aplicación de lectura (el core no debe conocer las respuestas HTTP).

### 2. Handler -> queries directo, sin interfaz

Cada handler de lectura vive en el `infra/http` de su módulo y recibe `*sqlc.Queries` directamente. No se define una interfaz de lectura: es infra -> infra (el handler HTTP y el paquete SQLC son ambos adaptadores), así que la concreción no cruza capas y una interfaz solo agregaría indirección sin desacoplar. Para un MVP, gana la simplicidad.

El único testing es la lógica pura (parseo de `status`, visibilidad, paginación por posición, mapeo a Response) con unit tests sin base de datos. No se agregan mocks ni tests de integración.

Se separan los handlers de escritura de los de lectura, con nombres simétricos (`ConversationWriteHandler` / `ConversationReadHandler`, `UserWriteHandler` / `UserReadHandler`, `ContactReadHandler`) y sus DTOs en `write_dto.go` / `read_dto.go`, para no mezclar responsabilidades.

### 3. Forma de las queries

- **Bandeja**: lee de `conversations` el `last_message_at`, `last_message_id` y `unread_count` ya denormalizados, más `contacts`. Visibilidad `(status = 'pending' OR user_id = :agente)`, tab obligatorio (`status = ANY(:statuses)`: `open` = `pending`/`assigned`, `finished` = `resolved`/`expired`) y filtro opcional por `contact_id`. `ORDER BY last_message_at DESC, id DESC` con posición `(before_sent_at, before_id)`, `LIMIT page_size + 1`. El preview y el dueño del último mensaje salen de un `JOIN messages lm ON lm.id = c.last_message_id` (lookup por PK sobre las ~21 filas ya ordenadas). Es `JOIN` y no `LEFT JOIN` porque la invariante "toda conversación tiene al menos un mensaje" garantiza que `last_message_id` siempre apunta a una fila, así que no hacen falta defensas por nulo ni `NULLS LAST`.
- **Detalle**: una query propia (`FindConversationWithContactByID`) trae metadata + `contact` + `unread_count` de la fila, más la página de mensajes con posición `(before_sent_at, before_id)` descendente, `LIMIT 21`, invertida a ascendente al mapear.
- **Contacto/Usuario**: se reutilizan `FindContactByID` y `FindUserByID`.

### 4. Texto de mensajes eliminados

La query devuelve `status` y `text` crudos, y la respuesta expone `text = null` cuando el estado es `deleted`. Se resuelve en el mapeo y no en SQL porque SQLC no infiere `*string` de una expresión `CASE` (genera `interface{}` o `string`), y así el texto borrado nunca se serializa.

### 5. Paginación por posición explícita

La paginación (keyset) usa dos parámetros explícitos: `before_sent_at` (RFC 3339) y `before_id` (uuid), que se envían juntos o ninguno; enviar solo uno o un valor no parseable es 400. La respuesta devuelve `next_before_sent_at` y `next_before_id`, nulos al agotar. Se descartó el cursor opaco: el cliente es de primera parte y lo controla el mismo equipo, así que empaquetar la posición en base64 no aportaba nada y solo agregaba código.

### 6. Autorización de lectura

- El middleware `Auth` valida el Bearer y, además, que el usuario exista (401 si no); así la comprobación vive en un solo lugar y los handlers no la repiten. Recibe el lookup como función (`UserLookup`), wireado desde `db.Queries`.
- Bandeja: el filtro de visibilidad va en el SQL.
- Detalle: se carga la conversación y el handler exige `status = 'pending'` o `user_id = :agente`; si no, 403; si no existe, 404.
- Contacto/Usuario: cualquier agente autenticado; 404 si no existe.
- El parseo de `{id}` se centraliza en `httputil.RequirePathUUID`, el mapeo de errores de lookup en `httputil.LookupError` y los helpers de uuid/tiempo en `pgdb`; `Auth` usa `users/domain.ErrUserNotFound` (infra -> domain).

### 7. Esquema e índices (se editan las migraciones existentes)

En dev, sin datos productivos, no se crea una migración nueva: columnas e índices se agregan editando las migraciones ya existentes, de modo que la tabla nazca con la estructura correcta y no haya `SWAP` ni lock al migrar.

- `000003_create_conversations.up.sql`:
  - columnas `last_message_at timestamptz`, `last_message_id uuid` y `unread_count integer NOT NULL DEFAULT 0 CHECK (unread_count >= 0)`.
  - índices `(user_id, last_message_at DESC, id DESC)`, `(status, last_message_at DESC, id DESC)` y `(contact_id)`. Los compuestos cubren por su primera columna los filtros por agente y por estado, así que no se agregan índices sueltos.
- `000004_create_messages.up.sql`: cambiar `messages_conversation_id_idx` por `messages (conversation_id, sent_at DESC, id DESC)` (cubre el prefijo por conversación, el orden por `sent_at` y el conteo de no leídos con filtro en consulta).
- No se crea índice parcial de no leídos.

Como esas migraciones ya corrieron en el dev DB, hay que recrear el esquema para re-aplicarlas.

### 8. Read model denormalizado, recomputado en la transacción

- `conversations` guarda `last_message_at`, `last_message_id` y `unread_count` para que la bandeja ordene y pagine por índice sin recorrer el historial.
- `last_message_id` apunta al mensaje con mayor `(sent_at, id)`; `last_message_at` es su `sent_at`; `unread_count` es la cantidad de mensajes del contacto con `read_at` nulo, **incluyendo los eliminados** (un mensaje borrado sigue representando una interacción del contacto). El conteo no filtra por `status`.
- `last_message_id` no lleva FK (`messages <-> conversations` sería circular); es un puntero derivado. Como en el MVP solo hay soft delete, no puede quedar colgando.
- Se mantienen con `RefreshConversationLastMessage`, que recomputa desde la tabla `messages`, dentro de la transacción de `Save`, después de upsertear los mensajes.
- El refresh se ejecuta **solo si hay mensajes modificados** (`len(dirty) > 0`), condición exacta porque solo los cambios de mensajes afectan a estas columnas; resolver o expirar no dispara refresh. La invariante "toda conversación tiene al menos un mensaje" garantiza que una conversación nueva siempre trae dirty al crearse, así que `last_message_*` queda seteado.
- Se recomputa (no se ajusta incremental) para que sea correcto aunque el agregado se haya cargado parcialmente, que es como lo cargan hoy los use cases.

### 9. Identidad en la API y filtro por contacto

- Los paths direccionan recursos por UUID interno (`/contacts/{id}`), pero la búsqueda de negocio usa el identificador externo. La bandeja se filtra por `external_contact_id`.
- Motivo: el agente conoce el identificador del canal (teléfono/PSID), no el UUID; filtrar por externo evita un pre-lookup. El UUID interno no se filtra, solo se expone para navegar.
- La fila expone `contact.id` y `contact.external_id`: el `id` para linkear a `GET /v1/contacts/{id}` y el `external_id` para mostrar y filtrar.
- Si el `external_contact_id` no existe, el handler responde 404 sin ejecutar la query de conversaciones.

### 10. Configuración de SQLC

Se desactiva `emit_json_tags`: los structs de SQLC son tipos de fila y nunca se serializan a JSON, así que sus tags eran metadata muerta. El contrato JSON vive únicamente en los DTOs. Se mantienen `emit_empty_slices` y `emit_pointers_for_null_types`.

## Risks / Trade-offs

- **Drift del read model** -> todos los writes pasan por `Save` y el refresh recomputa desde la tabla en la misma transacción. Si apareciera un write que no use `Save`, habría que engancharlo al refresh.
- **Refresh sin test automatizado** -> al usar solo unit tests, `RefreshConversationLastMessage` y las migraciones se verifican a mano (`make migrate-up`, `go build`), no con tests. Cubrirlos requeriría un test contra PostgreSQL, que quedó fuera de alcance.
- **JOIN del preview por `last_message_id`** -> el orden lo da el índice de `conversations` y el JOIN es por PK de las ~21 filas devueltas; no reintroduce el sort global.
- **`open` con OR (`pending` o `assigned` mío)** -> puede requerir ordenar un conjunto acotado (contactos activos). Aceptable en MVP; `finished` sí sale ordenado por índice.
- **Posición y precisión de timestamps** -> `timestamptz` en UTC con `RFC3339Nano`; empate resuelto por `id`. Riesgo bajo de saltos si dos elementos comparten instante exacto; mitigado por el desempate por identificador.
- **Duplicación de queries de contacto/usuario** -> se reutilizan las existentes; no hay copia.

## Migration Plan

1. Editar `000003_create_conversations.up.sql` (columnas + índices) y `000004_create_messages.up.sql` (índice de mensajes); los `down` no cambian porque `DROP TABLE` se lleva columnas e índices.
2. Recrear la BD dev para re-aplicar las migraciones (`make migrate-down` hasta la versión 2 y `make migrate-up`, o dropear y recrear el esquema).
3. `sqlc generate` para las queries nuevas (refresh + lecturas).
4. Repositorio de conversaciones: llamar `RefreshConversationLastMessage` al final de la transacción de `Save`.
5. Handlers de lectura + wiring en `cmd/api/app.go` y `cmd/api/router.go`.
6. Rollback: recrear la BD desde las migraciones; los handlers no alteran datos más allá del refresh del read model.

## Open Questions

Ninguna que afecte specs, enfoque ni desglose de tareas.
