# Design

## Context

No hay `db/migrations/`. `sqlc.yaml` ya declara `engine: postgresql` y `schema: db/migrations/`. Los repositorios solo buscan por clave: `agents.FindByID`, `contacts.FindByID` / `FindByExternalContactID`, y en conversaciones `FindWithoutMessages`, `FindOpenByContactID`, `FindWithMessageByExternalID`, `FindWithContactUnreadMessagesByID` y `Save`. PostgreSQL 18. Los identificadores de aplicación son UUIDv7. Ver proposal.md para el alcance.

## Goals / Non-Goals

**Goals:**

- Cuatro migraciones golang-migrate, una por tabla, reversibles, en el orden que exigen las FK.
- Restricciones e índices alineados con las consultas y las invariantes que SQL puede expresar.

**Non-Goals:**

- Cablear `github.com/golang-migrate/migrate` en el proceso, generar SQLC, escribir queries o implementar repositorios.
- Insertar agentes u otros datos.
- Ejecutar `migrate up`/`down` contra una base.
- Índices de bandeja (`agent_id`, `updated_at`, `sent_at`) o índice parcial de no leídos.

## Decisions

### Una migración por tabla

Archivos secuenciales. Cada `down` solo elimina su tabla. Índices y FK viajan con la tabla que los declara.

1. `000001_create_agents`
2. `000002_create_contacts`
3. `000003_create_conversations` — depende de agents y contacts
4. `000004_create_messages` — depende de las tres anteriores

Alternativa descartada: un solo archivo. El usuario pidió una migración por tabla.

### Tipos

- `uuid` sin `DEFAULT`. La aplicación asigna el id antes del `Save`. PostgreSQL 18 tiene `uuidv7()`, pero no es la fuente de verdad.
- `timestamptz` para todas las fechas.
- `text` + `CHECK` para estado y tipo, no enum de Postgres, para poder ampliarlos con un `ALTER` de check.
- `text` para el cuerpo del mensaje. El máximo de 1000 grafemas vive en el dominio (`uniseg`); `varchar` o `char_length` no lo expresan.

### agents

`id uuid PRIMARY KEY`, `created_at timestamptz NOT NULL`. Sin índices extra: la única consulta es por PK y no hay `Save`.

### contacts

`id`, `external_contact_id text NOT NULL`, `created_at`. `UNIQUE (external_contact_id)` cubre `FindByExternalContactID` y la carrera del get-or-create.

### conversations

Columnas: `id`, `status`, `agent_id`, `contact_id`, `created_at`, `updated_at`, `finished_at`. `agent_id` y `contact_id` son nulos. `updated_at` y `finished_at` son nulos.

FK: `agent_id REFERENCES agents(id) ON DELETE RESTRICT ON UPDATE RESTRICT`, `contact_id REFERENCES contacts(id) ON DELETE RESTRICT ON UPDATE RESTRICT`.

Checks:

- `status IN ('pending', 'assigned', 'expired', 'resolved')`
- `agent_id IS NOT NULL OR contact_id IS NOT NULL`
- `status IN ('pending', 'assigned') OR finished_at IS NOT NULL`

Índice:

```sql
CREATE UNIQUE INDEX conversations_one_open_per_contact
    ON conversations (contact_id)
    WHERE status IN ('pending', 'assigned')
      AND contact_id IS NOT NULL;
```

Abierta significa `pending` o `assigned`. El índice único parcial es a la vez la consulta de `FindOpenByContactID` y la barrera ante dos altas concurrentes. Un `UNIQUE (contact_id)` impediría el historial.

### messages

Columnas: `id`, `conversation_id NOT NULL`, `status`, `type`, `text`, `agent_id`, `contact_id`, `external_id`, `sent_at`, `read_at`, `edited_at`, `deleted_at`. El dominio no guarda `conversation_id` en `Message`; es dato de persistencia.

FK `ON DELETE RESTRICT ON UPDATE RESTRICT` hacia `conversations`, `agents` y `contacts`. El `ON UPDATE RESTRICT` declara que un `id` referenciado por otras tablas no se puede reescribir; el `NO ACTION` por defecto daría el mismo resultado con claves inmutables, pero aquí queda explícito.

Checks:

- `status IN ('sent', 'read', 'deleted', 'failed')`
- `type IN ('text')`
- `(agent_id IS NULL) <> (contact_id IS NULL)`

Índices:

- `messages (conversation_id)` — carga de mensajes y no leídos; el filtro `read_at IS NULL AND contact_id IS NOT NULL` va en la consulta. Un parcial de no leídos no vale en el MVP.
- único parcial `messages (external_id) WHERE external_id IS NOT NULL`. `FindWithMessageByExternalID` no recibe conversación y devuelve una sola fila, así que la unicidad es global. El índice existe por ese contrato, no por volumen.

No se indexan `agent_id`, `sent_at` ni `updated_at`. Postgres no indexa solo las FK; el índice de `conversation_id` se crea a mano. Las FK de `agent_id` y `contact_id` no se indexan: no hay borrados ni búsquedas por esas columnas en el MVP.

### Sin datos

Ninguna migración inserta filas. La FK de agente rechazará el envío hasta que otro cambio cree el alta. Es el comportamiento esperado.

## Risks / Trade-offs

- [Alta de agente inexistente] → el envío falla por FK hasta un cambio posterior. No se siembra un uuid falso.
- [Dos webhooks abren conversación a la vez] → el único parcial rechaza la segunda. El repositorio aún no reintenta; queda fuera de este cambio.
- [`external_id` global] → más estricto que la unicidad por conversación del dominio. Encaja con la firma del repositorio. Si un canal colisionara entre conversaciones, habría que cambiar firma e índice juntos.
- [Check de `type IN ('text')`] → un tipo nuevo exige otra migración. Aceptado para no usar enum.

## Migration Plan

Este cambio solo crea los archivos de migración; ejecutar `up` o `down` contra una base queda fuera de su alcance. Para cuando se apliquen: `up` en orden `000001`..`000004`, rollback `down` desde `000004` hasta `000001`. Cada `down` hace `DROP TABLE` de su tabla; el orden inverso respeta las FK. No hay datos que migrar.

## Open Questions

Ninguna que cambie el esquema de este cambio.
