# Tasks

## 1. Andamiaje del script y ejecución

- [x] 1.1 Crear `db/seed/dev_seed.sql` con encabezado de advertencia dev-only, la guarda que aborta si `current_database()` no contiene `dev`, y el `TRUNCATE users, contacts, conversations, messages RESTART IDENTITY CASCADE`. Verificar con `make reset && make seed` que las cuatro tablas quedan vacías y el script termina sin error.
- [x] 1.2 Agregar el target `seed` al `Makefile` (`docker compose exec -T postgres sh -c 'psql -v ON_ERROR_STOP=1 ...' < db/seed/dev_seed.sql`). Verificar que `make seed` corre el script dentro del contenedor y que un error SQL aborta con código distinto de cero.
- [x] 1.3 Documentar `make seed` en `AGENTS.md` (comando, requisito de base dev, advertencia de que es destructivo y nunca migración/producción). Verificar leyendo la sección y ejecutando el comando tal como quedó escrito.

## 2. Agentes y contactos

- [x] 2.1 Insertar 3 agentes en `users` con `uuidv7()` y `created_at` en el pasado. Verificar `SELECT count(*) FROM users` = 3.
- [x] 2.2 Insertar 50 contactos en `contacts` con `id = uuidv7()` y `external_contact_id = uuidv7()::text`. Verificar `count(*)` = 50 y que `count(DISTINCT external_contact_id)` = 50.

## 3. Plan de conversaciones

- [x] 3.1 Generar el plan de 90 conversaciones: `pending` 12, `assigned` 18 (6/6/6), `resolved` 30 (10/10/10), `expired` 30 (20 con agente 7/7/6 + 10 sin agente), garantizando a lo sumo una abierta por contacto y que las finalizadas de un contacto sean anteriores a su abierta. Verificar los conteos por estado y `open_por_contacto <= 1` con SQL de comprobación.

## 4. Catálogo temático de consolas

- [x] 4.1 Crear los pools de frases por tema (`stock`, `precio`, `envio`, `pago`, `garantia`, `devolucion`, `accesorios`) y por rol (contacto/agente), en español neutro. Verificar que cada conversación toma un tema y que sus textos salen de ese pool y no están vacíos.

## 5. Generación de mensajes

- [x] 5.1 Repartir 3000 mensajes con mínimo 5 y máximo 50 por conversación, sesgando `pending` 5..10, `assigned` 15..45 y `resolved`/`expired` 20..50, ajustando la suma exacta a 3000. Verificar `count(*)` = 3000 y `min/max` por conversación en `[5,50]`.
- [x] 5.2 Asignar dueño y `sent_at` crecientes por conversación (contacto inicia; agente responde cuando aplica). Verificar que `pending` y `expired` sin agente solo tienen mensajes del contacto y que no hay `sent_at` duplicados por conversación.
- [x] 5.3 Aplicar el modelo de lectura: en cada mensaje de agente en `t` se marca `read_at = t` a los del contacto con `sent_at < t`; el resto queda sin leer. Verificar que `read_at` solo aparece en mensajes del contacto y que la racha final sin leer coincide con la regla.
- [x] 5.4 Generar `external_id = uuidv7()::text` en mensajes del contacto y en mensajes de agente enviados; dejarlo `NULL` solo en `failed`. Verificar unicidad global y que `failed` no tiene `external_id`.
- [x] 5.5 Incluir trazas puntuales de `deleted` (contacto o agente), `failed` (solo agente) y `edited`, respetando dueños y el orden de fechas (edición de agente antes de `finished_at`; borrados del contacto pueden ser posteriores). Verificar que `failed` es solo de agente y que ningún `deleted_at`/`edited_at` viola su regla.
- [x] 5.6 Asignar `finished_at` (solo `resolved`/`expired`) posterior al último `sent_at` y `updated_at = max(actividad)` (≥ `finished_at`). Verificar que en finalizadas todo `sent_at < finished_at` y que `updated_at >= finished_at`.
- [x] 5.7 Generar las 10 conversaciones `expired` sin agente del caso borde. Verificar que tienen `user_id NULL`, `finished_at` no nulo, solo mensajes del contacto y `unread_count = total`.

## 6. Modelo de lectura denormalizado

- [x] 6.1 Recomputar por lote `last_message_id`, `last_message_at` y `unread_count` para todas las conversaciones (equivalente a `RefreshConversationLastMessage`). Verificar comparando contra un `SELECT` de control que use `max(sent_at,id)` y el `count` de mensajes del contacto con `read_at NULL`.

## 7. Validaciones y cierre del script

- [x] 7.1 Añadir el bloque final de validaciones que hace `RAISE EXCEPTION` si falla cualquier invariante (conteos, 5..50 por conversación, ≤1 abierta por contacto, denormalizado, fechas, dueño XOR, `read_at`/`failed`). Verificar que `make seed` pasa, y que al alterar a mano una fila para violar una regla el script falla.
- [x] 7.2 Confirmar con una consulta que las 10 `expired` sin agente **no** aparecen en `open` ni en `finished` de ningún agente y que su `id` da 403 en `GET /v1/conversations/{id}` (reproduce el caso borde).

## 8. Verificación de integración

- [x] 8.1 Correr `make reset && make seed`, levantar la API y comprobar manualmente que `GET /v1/conversations?status=open` y `status=finished` devuelven los volúmenes esperados por agente, que la vista previa y el badge de no leídos son coherentes, y que abrir una conversación con historial responde con sus últimos 20 mensajes en orden.
