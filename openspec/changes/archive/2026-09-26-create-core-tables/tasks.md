# Tasks

## 1. Agents

- [x] 1.1 Crear `db/migrations/000001_create_agents.up.sql` con `agents(id uuid PRIMARY KEY, created_at timestamptz NOT NULL)`, sin `DEFAULT`, sin índices extra y sin `INSERT`. Verificar que el archivo contiene esa tabla y no contiene `INSERT` ni `REFERENCES`.
- [x] 1.2 Crear `db/migrations/000001_create_agents.down.sql` con solo `DROP TABLE agents`. Verificar que no elimina otras tablas.

## 2. Contacts

- [x] 2.1 Crear `db/migrations/000002_create_contacts.up.sql` con `contacts(id uuid PRIMARY KEY, external_contact_id text NOT NULL, created_at timestamptz NOT NULL)` y `UNIQUE (external_contact_id)`, sin `DEFAULT` y sin `INSERT`. Verificar que el único extra es ese `UNIQUE`.
- [x] 2.2 Crear `db/migrations/000002_create_contacts.down.sql` con solo `DROP TABLE contacts`. Verificar que no elimina otras tablas.

## 3. Conversations

- [x] 3.1 Crear `db/migrations/000003_create_conversations.up.sql` con `id`, `status`, `agent_id`, `contact_id`, `created_at`, `updated_at` y `finished_at` nulos salvo `id`, `status` y `created_at`. Verificar que no hay `DEFAULT` ni `INSERT`.
- [x] 3.2 Añadir FK `ON DELETE RESTRICT ON UPDATE RESTRICT` a `agents(id)` y `contacts(id)`, checks de estado (`pending`, `assigned`, `expired`, `resolved`), de agente o contacto presente, y de `finished_at` obligatorio si el estado es `expired` o `resolved`. Verificar que el `up` contiene esas tres restricciones y `ON DELETE RESTRICT ON UPDATE RESTRICT`.
- [x] 3.3 Añadir el índice único parcial `conversations_one_open_per_contact` sobre `contact_id` donde `status IN ('pending', 'assigned') AND contact_id IS NOT NULL`. Verificar que no hay índice en `agent_id` ni en `updated_at`.
- [x] 3.4 Crear `db/migrations/000003_create_conversations.down.sql` con solo `DROP TABLE conversations`. Verificar que no elimina `agents` ni `contacts`.

## 4. Messages

- [x] 4.1 Crear `db/migrations/000004_create_messages.up.sql` con `conversation_id uuid NOT NULL` y el resto de columnas del diseño (`status`, `type`, `text`, `agent_id`, `contact_id`, `external_id`, `sent_at`, `read_at`, `edited_at`, `deleted_at`), sin `DEFAULT` y sin `INSERT`. Verificar que `text` es `text`, no `varchar`.
- [x] 4.2 Añadir FK `ON DELETE RESTRICT ON UPDATE RESTRICT` a `conversations`, `agents` y `contacts`, checks de estado, de `type IN ('text')` y de dueño exclusivo `(agent_id IS NULL) <> (contact_id IS NULL)`. Verificar que el `up` contiene las tres FK y el check de dueño.
- [x] 4.3 Añadir índice no único en `conversation_id` e índice único parcial de `external_id` donde `external_id IS NOT NULL`. Verificar que no hay índice en `agent_id`, `contact_id`, `sent_at` ni parcial de `read_at`.
- [x] 4.4 Crear `db/migrations/000004_create_messages.down.sql` con solo `DROP TABLE messages`. Verificar que no elimina las otras tablas.
