# Design: Datos de prueba para la base de datos en dev

## Context

La base tiene cuatro tablas (`users`, `contacts`, `conversations`, `messages`) con las
restricciones descritas en `openspec/specs/persistence/spec.md` y el ciclo de vida de
`openspec/specs/conversation/spec.md`. Puntos que condicionan el seed:

- Una sola conversación abierta (`pending`/`assigned`) por contacto.
- `finished_at` obligatorio en `expired`/`resolved`; en conversaciones finalizadas todo
  mensaje debe tener `sent_at < finished_at`.
- Un mensaje tiene exactamente un dueño; `read_at` solo aplica a mensajes del contacto;
  `failed` es solo de agente y terminal.
- El read model denormalizado (`last_message_id`, `last_message_at`, `unread_count`) debe
  quedar coherente con los mensajes.
- Postgres 18 trae `uuidv7()` nativo, así que no hace falta `pgcrypto` ni `uuid-ossp`.
- No hay `psql` en el host; Postgres corre en Docker Compose.
- El script es solo de desarrollo: nunca migración, nunca producción, no lo toca CI.

Ver `proposal.md - Why` para la motivación.

## Goals / Non-Goals

**Goals:**

- Un único `.sql` re-ejecutable que deje la base de dev con data verosímil y 100%
  consistente con las reglas de negocio.
- Volumen objetivo: 3 agentes, 50 contactos, 90 conversaciones, 3000 mensajes.
- Conversaciones temáticas sobre compra de consolas, con hilos coherentes.
- `unread_count` con sentido en las conversaciones abiertas.
- Reproducir el caso borde de `expired` sin agente (10 conversaciones).

**Non-Goals:**

- No arreglar el hueco de visibilidad de `expired` sin agente (cambio aparte).
- No cambiar el esquema, las queries, la API ni las specs.
- No ser determinista byte a byte: los UUID v7 son aleatorios en cada corrida.
- No cubrir tests automáticos (CI no tiene base).

## Decisions

### 1. SQL puro, no comando Go

Se descarta reusar `cmd/seed` con SQLC. El script es dev-only y se quiere ejecutar a
voluntad sin compilar; un `.sql` es autónomo y portable. Contrapartida: hay que replicar
en SQL el denormalizado y la coherencia de fechas, cosa que se asume con validaciones
explícitas al final. Para el reparto (conteos y orden temporal) se usa un bloque `DO`
plpgsql dentro del mismo archivo; el resto es SQL set-based.

### 2. Ubicación y ejecución

`db/seed/dev_seed.sql`, fuera de `db/migrations/`, más un target `make seed`:

```
seed:
	docker compose exec -T postgres sh -c 'psql -v ON_ERROR_STOP=1 -U "$$POSTGRES_USER" -d "$$POSTGRES_DB"' < db/seed/dev_seed.sql
```

`ON_ERROR_STOP=1` garantiza que cualquier violación (incluidas las validaciones del paso
final) aborte con código distinto de cero.

### 3. UUID v7 para todo

`id` de todas las entidades y `external_contact_id`/`external_id` se generan con
`uuidv7()::text`. Ventaja: unicidad garantizada y orden temporal. Nota: el timestamp que
embebe el v7 es el de ejecución del seed, no el `sent_at` histórico; al dominio le es
indiferente porque el id externo es opaco.

### 4. Volumen y mezcla

```
C = 90
  open    30: pending 12  +  assigned 18 (6/6/6 por agente)
  finish  60: resolved 30 (10/10/10)
              expired  30 = 20 con agente (7/7/6) + 10 sin agente
```

30 conversaciones abiertas ocupan 30 contactos distintos (≤ 50). Las 60 finalizadas se
apilan sobre los 50 contactos (0..3 por contacto, siempre anteriores a la abierta del
mismo contacto).

### 5. Distribución de mensajes

3000 mensajes, mínimo 5 y máximo 50 por conversación. Se sesga por estado para que sea
verosímil y para llegar al promedio (3000/90 ≈ 33.3):

```
pending              5..10   (del contacto, sin agente)
assigned            15..45
resolved / expired  20..50
```

El bloque `DO` ajusta la suma a exactamente 3000 respetando los topes.

### 6. Modelo de lectura y no leídos

```
pending / expired sin agente:  solo contacto, read_at NULL
                               unread_count = total de mensajes
con agente:                    en cada mensaje del agente en t se marcan
                               leídos los del contacto con sent_at < t
                               => unread = racha final de mensajes del contacto
```

Así, cuando `unread_count > 0` el último mensaje es del contacto; cuando es 0, el último
es del agente. La vista previa y el badge quedan coherentes en la bandeja.

### 7. Realismo

Empresa que vende consolas. Cada conversación elige un tema y las frases del contacto y
del agente salen de pools de ese tema:

```
stock | precio | envio | pago | garantia | devolucion | accesorios
```

Los `sent_at` usan huecos realistas (minutos/horas dentro de una sesión, días entre
sesiones). Se incluyen trazas puntuales de `deleted` (contacto/agente), `failed`
(agente, `external_id` NULL) y `edited`, respetando dueños y fechas.

### 8. Denormalizado por lote

Al final, un único `UPDATE` recomputa `last_message_id`, `last_message_at` y
`unread_count` para todas las conversaciones, equivalente a
`RefreshConversationLastMessage` (queries SQLC):

```
last_message_id/at = mensaje con max(sent_at, id) de la conversación
unread_count       = count(mensajes del contacto con read_at NULL)
```

### 9. Caso borde `expired` sin agente

10 conversaciones con `user_id NULL`, `contact_id` presente, `finished_at` tras el último
mensaje y solo mensajes del contacto. Se generan para reproducir que hoy ningún listado
las muestra y `GET /v1/conversations/{id}` responde 403. El seed **no** las hace visibles
ni las valida como visibles.

### 10. Validaciones finales

Al terminar, el script verifica y hace `RAISE EXCEPTION` si algo falla: conteos por tabla,
5..50 mensajes por conversación, ≤1 abierta por contacto, `unread_count` y
`last_message_*` recomputados contra `messages`, todo `sent_at < finished_at` en
finalizadas, `read_at` solo en mensajes del contacto, `failed` solo de agente y sin
`external_id`, dueño XOR, y `updated_at >= finished_at >= max(sent_at)`.

### 11. Seguridad dev-only

Encabezado de advertencia, `TRUNCATE ... RESTART IDENTITY CASCADE`, y aborto si
`current_database()` no contiene `dev`.

## Risks / Trade-offs

- [Destructivo] Un `TRUNCATE` apuntado a prod borra todo → guarda por nombre de base,
  ubicación fuera de migraciones, target solo dev y advertencia visible.
- [SQL complejo y frágil] El reparto y el orden temporal en plpgsql pueden desviarse →
  validaciones finales que fallan ruidoso con `ON_ERROR_STOP=1`.
- [UUID v7 externos] Su tiempo es el de ejecución, no el del mensaje → irrelevante para
  el dominio; documentado para que nadie se confunda al inspeccionar.
- [No determinista] Los IDs cambian entre corridas → aceptable porque el seed trunca y
  recarga; las distribuciones (conteos/estados) sí son estables.
- [Hueco del spec] Las 10 `expired` sin agente quedan invisibles → es intencional para
  reproducir; su arreglo es un cambio aparte.

## Open Questions

- Locale del texto (español neutro vs. Colombia) y catálogo final de frases: se puede
  ajustar al implementar sin tocar el enfoque.
