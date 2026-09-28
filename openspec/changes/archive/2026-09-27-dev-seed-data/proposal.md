# Proposal: Datos de prueba para la base de datos en dev

## Why

Hoy no hay forma de poblar la base en desarrollo con datos que parezcan reales, así que
probar a mano la bandeja, el historial de una conversación y las reglas de negocio exige
crear filas a mano y es fácil terminar con data que viola los invariantes del dominio
(una conversación abierta por contacto, `unread_count` coherente, fechas válidas, etc.).
Sin data realista no se puede validar visualmente el read model ni reproducir estados
borde.

## What Changes

- Nuevo script SQL puro `db/seed/dev_seed.sql` (Postgres 18) que **trunca y recarga**
  `users`, `contacts`, `conversations` y `messages` con datos verosímiles.
- Volumen: 3 agentes, 50 contactos, 90 conversaciones (30 open, 60 finished) y 3000
  mensajes (5..50 por conversación).
- Todos los identificadores, incluidos `external_contact_id` y `external_id`, se generan
  con `uuidv7()` (nativo de Postgres 18).
- Contenido temático: conversaciones sobre compra de consolas (stock, precio, envío,
  pago, garantía, devolución, accesorios), con hilos coherentes por tema.
- Coherencia de no leídos en conversaciones abiertas: `unread_count` refleja la racha
  final de mensajes del contacto posteriores al último mensaje del agente.
- Bloque de caso borde: 10 conversaciones `expired` sin agente (invisibles en el read
  model actual) para reproducir ese hueco del spec.
- Nuevo target `make seed` que ejecuta el script dentro del contenedor de Postgres.
- El script aborta si la base no parece de desarrollo y no lo referencia CI ni ninguna
  migración.

## Capabilities

### New Capabilities

Ninguna. Este cambio es únicamente tooling de desarrollo: no altera el comportamiento del
producto ni el de sus specs, por lo que el cambio marca `skip_specs: true` en su
`.openspec.yaml`.

### Modified Capabilities

Ninguna.

## Impact

- Archivos nuevos: `db/seed/dev_seed.sql` (fuera de `db/migrations/`).
- `Makefile`: nuevo target `seed` (solo dev).
- No cambia código de producto, API, dependencias ni specs.
- Riesgo: el script es destructivo (`TRUNCATE`). Mitigado con guarda de nombre de base,
  encabezado de advertencia y ubicación fuera de migraciones.
- El hueco de visibilidad de `expired` sin agente queda documentado como caso borde; su
  arreglo es un cambio aparte.
