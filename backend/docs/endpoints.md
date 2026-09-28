# 1Channel API

Referencia de los endpoints HTTP y ejemplos `curl`. Todos los `payloads` son JSON y los
timestamps van en UTC (RFC 3339).

- Base local: `http://localhost:8080`
- Auth: `Authorization: Bearer <user id>` (el `id` de un `user` existente)
- Errores: `Content-Type: application/problem+json` con `title`, `status`, `detail` e `instance`
- Paginación: `before_sent_at` (RFC 3339) + `before_id`, siempre juntos; las respuestas
  devuelven `next_before_sent_at` / `next_before_id` (nulos cuando no hay más)

## Preparación

```bash
make up            # postgres 18 (o: make DOCKER="sudo docker" up)
make migrate-up
make seed          # datos de prueba (solo dev; destructivo)
make run           # levanta la API en :8080

# id de un agente para usar como token
sudo docker compose exec -T postgres psql -U dev -d 1channel_dev -tAc \
  "select id from users order by created_at, id limit 1;"
```

En los ejemplos se asume:

```bash
BASE=http://localhost:8080
TOKEN=01a0e5b3-d609-7fa1-8dbe-5586533fc2b5        # id de un user
CID=01a0e5b3-d772-777b-878e-a1bb4b024452          # id de una conversación
CTID=<id interno de un contacto>
```

---

## Health

### GET /healthz

Sin auth. Responde `200` sin cuerpo.

```bash
curl -i "$BASE/healthz"
```

---

## Users

### POST /v1/users

Público (bootstrap). Crea un `user` (agente) nuevo. Sin cuerpo.

- `201` `{"id":"<uuid>"}`

```bash
curl -s -X POST "$BASE/v1/users" | jq
```

### GET /v1/users/{id}

Requiere auth. `id` es el del usuario a consultar (puede ser el mismo del token).

- `200` `{"id":"<uuid>","created_at":"..."}`
- `400` si el `id` no es UUID · `404` si no existe

```bash
curl -s "$BASE/v1/users/$TOKEN" -H "Authorization: Bearer $TOKEN" | jq
```

---

## Contacts

### GET /v1/contacts/{id}

Requiere auth. `id` es el identificador interno del contacto (no el `external_id`).

- `200` `{"id":"<uuid>","external_id":"<uuid>","created_at":"..."}`
- `400` si el `id` no es UUID · `404` si no existe

```bash
curl -s "$BASE/v1/contacts/$CTID" -H "Authorization: Bearer $TOKEN" | jq
```

---

## Conversations (lectura)

### GET /v1/conversations

Requiere auth. Lista la bandeja visible para el agente del token (máximo 20 por página,
ordenada por el `sent_at` del último mensaje descendente).

Query params:

| Param                  | Obligatorio | Descripción                                            |
|------------------------|-------------|--------------------------------------------------------|
| `status`               | sí          | `open` (pending + assigned) o `finished` (resolved + expired) |
| `external_contact_id`  | no          | filtra por contacto; si no existe responde `404`       |
| `before_sent_at`       | no          | posición; debe ir junto con `before_id`                |
| `before_id`            | no          | posición; debe ir junto con `before_sent_at`           |

Visibilidad: sin agente asignado (incluye `pending` y `expired` sin agente), o asignada al
solicitante. Las `resolved`/`expired` con agente se listan en `finished` solo para su agente.

- `200`:

```json
{
  "items": [
    {
      "id": "<uuid>",
      "status": "assigned",
      "contact": { "id": "<uuid>", "external_id": "<uuid>" },
      "last_message": { "text": "...", "sent_at": "...", "owner": "contact" },
      "unread_count": 2
    }
  ],
  "next_before_sent_at": null,
  "next_before_id": null
}
```

- `422` si falta `status` o no es `open`/`finished` · `400` si la posición viene a medias
  o mal formada · `401` sin token válido

```bash
# abiertas
curl -s "$BASE/v1/conversations?status=open" \
  -H "Authorization: Bearer $TOKEN" | jq

# finalizadas
curl -s "$BASE/v1/conversations?status=finished" \
  -H "Authorization: Bearer $TOKEN" | jq

# filtrando por contacto
curl -s "$BASE/v1/conversations?status=open&external_contact_id=<external_id>" \
  -H "Authorization: Bearer $TOKEN" | jq

# página siguiente (usando los next_* de la respuesta anterior)
curl -s "$BASE/v1/conversations?status=open&before_sent_at=<next_before_sent_at>&before_id=<next_before_id>" \
  -H "Authorization: Bearer $TOKEN" | jq
```

### GET /v1/conversations/{id}

Requiere auth. Devuelve la conversación y sus últimos 20 mensajes en orden `sent_at`
ascendente (empate por `id`). Accesible si no tiene agente asignado o si el agente asignado
es el solicitante.

- `200`:

```json
{
  "id": "<uuid>",
  "status": "assigned",
  "contact": { "id": "<uuid>", "external_id": "<uuid>" },
  "agent_id": "<uuid>",
  "unread_count": 2,
  "messages": [
    {
      "id": "<uuid>",
      "status": "read",
      "type": "text",
      "text": "...",
      "owner": "contact",
      "sent_at": "...",
      "read_at": "...",
      "edited_at": null,
      "deleted_at": null
    }
  ],
  "next_before_sent_at": "...",
  "next_before_id": "<uuid>"
}
```

- Los mensajes `deleted` se incluyen con `status":"deleted"` y `text:null`.
- `403` si no es accesible · `404` si no existe · `400` si el `id` no es UUID

```bash
curl -s "$BASE/v1/conversations/$CID" \
  -H "Authorization: Bearer $TOKEN" | jq

# mensajes más antiguos
curl -s "$BASE/v1/conversations/$CID?before_sent_at=<next_before_sent_at>&before_id=<next_before_id>" \
  -H "Authorization: Bearer $TOKEN" | jq
```

---

## Conversations (escritura)

### POST /v1/conversations/{id}/messages

Requiere auth. El agente envía un mensaje de texto. Si la conversación está `pending` sin
agente, el solicitante queda asignado y pasa a `assigned`.

Body: `{"text":"..."}` (1..1000 grafemas).

- `201` `{"id":"<uuid>"}`
- `403` si la conversación es de otro agente
- `404` si la conversación no existe
- `409` si la conversación está finalizada y ya no acepta mensajes
- `422` si el texto es vacío o excede el máximo

```bash
curl -s -X POST "$BASE/v1/conversations/$CID/messages" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"text":"Hola, con gusto te ayudo con la compra de la consola."}' | jq
```

### PATCH /v1/conversations/{id}/messages

Requiere auth. Marca como leídos los mensajes del contacto que estén sin leer. Solo el
agente asignado; funciona también en conversaciones finalizadas.

Body: `{"status":"read"}`.

- `204` sin cuerpo
- `403` si no es el agente asignado · `404` si no existe · `422` si `status` no es `read`

```bash
curl -s -o /dev/null -w "%{http_code}\n" -X PATCH "$BASE/v1/conversations/$CID/messages" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"read"}'
```

### PATCH /v1/conversations/{id}

Requiere auth. El agente asignado resuelve la conversación (idempotente si ya está
finalizada).

Body: `{"status":"resolved"}`.

- `204` sin cuerpo
- `403` si no es el agente asignado · `404` si no existe · `422` si `status` no es `resolved`

```bash
curl -s -o /dev/null -w "%{http_code}\n" -X PATCH "$BASE/v1/conversations/$CID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"resolved"}'
```

---

## Webhooks

### POST /v1/webhooks/1channel

Público (sin auth). Canal de prueba para las operaciones del contacto. El timestamp lo
genera el servidor; no viaja en el cuerpo. El cuerpo es un envelope único con `event`
obligatorio y los campos que cada evento requiera.

Eventos y campos:

| `event`            | `external_contact_id` | `external_message_id` | `text`   |
|--------------------|-----------------------|-----------------------|----------|
| `message.received` | requerido             | requerido             | requerido |
| `message.edited`   | requerido             | requerido             | requerido |
| `message.deleted`  | requerido             | requerido             | —        |
| `message.read`     | requerido             | requerido             | —        |

- `204` sin cuerpo
- `400` si el cuerpo no es JSON válido
- `422` si falta un campo requerido · `200` sin efecto si `event` es desconocido
- `403` si el mensaje es de otro contacto · `404` si el mensaje no existe

```bash
# recibir
curl -s -o /dev/null -w "%{http_code}\n" -X POST "$BASE/v1/webhooks/1channel" \
  -H "Content-Type: application/json" \
  -d '{"event":"message.received","external_contact_id":"5491112345678","external_message_id":"wamid.001","text":"Hola, quiero comprar una consola."}'

# editar
curl -s -o /dev/null -w "%{http_code}\n" -X POST "$BASE/v1/webhooks/1channel" \
  -H "Content-Type: application/json" \
  -d '{"event":"message.edited","external_contact_id":"5491112345678","external_message_id":"wamid.001","text":"Hola, quiero comprar una consola roja."}'

# eliminar
curl -s -o /dev/null -w "%{http_code}\n" -X POST "$BASE/v1/webhooks/1channel" \
  -H "Content-Type: application/json" \
  -d '{"event":"message.deleted","external_contact_id":"5491112345678","external_message_id":"wamid.001"}'

# acuse de lectura de un mensaje del agente
curl -s -o /dev/null -w "%{http_code}\n" -X POST "$BASE/v1/webhooks/1channel" \
  -H "Content-Type: application/json" \
  -d '{"event":"message.read","external_contact_id":"5491112345678","external_message_id":"wamid.agent.001"}'
```

---

## Errores

Formato uniforme:

```json
{
  "title": "Forbidden",
  "status": 403,
  "detail": "conversation is not accessible",
  "instance": "debian/qsPySCohma-000009"
}
```

| Código | Cuándo                                                        |
|--------|---------------------------------------------------------------|
| 400    | Entrada malformada (JSON, `id` de path, posición de paginación) |
| 401    | Sin token, token mal formado o usuario inexistente            |
| 403    | Sin permiso sobre el recurso                                  |
| 404    | Recurso inexistente                                           |
| 409    | Conflicto de estado (conversación finalizada, mensaje failed, etc.) |
| 422    | Validación (`status` inválido, texto vacío o demasiado largo) |
| 500    | Error interno                                                 |
| 504    | Timeout del request                                           |

Ejemplo de error forzado:

```bash
curl -s "$BASE/v1/conversations?status=open" | jq          # 401 sin token
curl -s "$BASE/v1/conversations" -H "Authorization: Bearer $TOKEN" | jq   # 422 sin status
```

---

## Notas

- El `POST /v1/users` es público (bootstrap); el resto exige `Authorization`.
- Las operaciones del contacto (recibir, editar, eliminar y acuse de lectura) llegan por el
  webhook de prueba `POST /v1/webhooks/1channel`, público y sin auth. Es la contraparte de
  prueba de los futuros adaptadores por plataforma (`/v1/webhooks/meta`, `/v1/webhooks/telegram`).
- Un agente no puede pasar una conversación a `expired` por HTTP (responde `422`).
- Los `external_id` y los `id` son UUID v7.
