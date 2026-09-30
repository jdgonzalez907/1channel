# Design

## Context

El núcleo de ingestión ya existe: los casos de uso `ReceiveContactMessage`, `ReceiveContactMessageEdit`, `ReceiveContactMessageDelete` y `ReceiveContactMessageRead` están implementados en `conversations/app` y expuestos por `conversations.ConversationsAPI`. El adaptador de canal de prueba (`infra/http/contact_webhook_handler.go`) traduce el JSON del simulador a esos casos de uso. El envío del agente usa hoy un `StubSender` (`infra/messagesender`) que inventa un identificador externo.

El contrato de salida lo define `domain.AgentMessageSender.Send(ctx, to, *Message) (string, error)` y el caso de uso `SendAgentMessage` ya resuelve el `to` (PSID) y maneja el ciclo error→`failed` / éxito→`AssignAgentMessageExternalID`.

La ruta raíz no exige `Authorization`: el `router.go` solo aplica `Auth` dentro del grupo `/v1`, por lo que una ruta registrada fuera de ese grupo es pública.

## Goals / Non-Goals

**Goals:**
- Un paquete autocontenido `conversations/infra/meta` con la configuración, el handler del webhook y el sender de Meta.
- Reutilizar los casos de uso de contacto sin tocar `domain` ni `app`.
- Verificación de firma y handshake de verificación correctos.

**Non-Goals:**
- Acuses de lectura (`read`) y entregas (`delivery`): se ignoran.
- Adjuntos/media, `reply_to`, postbacks, reacciones, ecos (`is_echo`) y `object != page`: se ignoran con 200.
- Instagram messaging: solo se procesa `object == "page"`.
- Message tags / `UPDATES` / `TAGGED_MESSAGE`: solo `RESPONSE`.

## Decisions

### 1. Paquete autocontenido `infra/meta`

Todo lo específico de Meta vive en `conversations/infra/meta`: `config.go` (una sola `Config` con PageID, PageAccessToken, AppSecret y VerifyToken), `webhook.go` (handler HTTP con `Register(r)`), los DTOs (`write_dto.go`/`read_dto.go`) y `messenger_sender.go` (producto Messenger/Page; implementa `domain.AgentMessageSender`). Cuando entre otro producto (WhatsApp) se añade su `whatsapp_sender.go`.

- **Alternativa considerada**: repartir el handler en `infra/http` y el sender en `infra/messagesender`. Se descarta: fragmenta un único canal en dos paquetes y complica compartir la configuración y el cliente HTTP.

### 2. Cualquier fallo de procesamiento responde 500 (Meta reintenta)

A diferencia del simulador (`ContactWebhookHandler`), que mapea errores de dominio a `problem+json` (422/404/403), el consumidor de `/webhooks/meta` es una máquina (Meta), no el agente. Meta reintenta ante cualquier no-2xx y no entiende nuestros códigos.

- **Decisión**: tras procesar el lote, si algún evento falló → registrar el error y responder 500 para que Meta reintente; si todo se procesa → 200. No se distingue el tipo de error.
- **Por qué**: Meta no garantiza el orden de entrega, así que un error que parece permanente (p. ej. `ErrMessageNotFound` de una edición que llegó antes que su mensaje) puede resolverse en un reintento posterior; un 200 incondicional perdería ese evento. La idempotencia (dedup por `external_id`, comparación de timestamps) hace seguros los reintentos.
- **Trade-off**: un error genuinamente permanente (p. ej. texto de más de 1000 grafemas) se reintentará hasta que Meta desuscriba el webhook (~1h); se asume para el MVP y se registra.

### 3. Firma sobre el cuerpo crudo

Se lee el body completo con `io.ReadAll`, se calcula `HMAC-SHA256(body, META_APP_SECRET)` y se compara con la cabecera `X-Hub-Signature-256` (tras `sha256=`). Solo después se hace `json.Unmarshal`.

- **Por qué**: usar los bytes crudos evita el problema de re-serialización/escaped-unicode que describe la doc de Meta (la firma se calcula sobre la versión con unicode escapado; si re-serializáramos obtendríamos otra firma).

### 4. Timestamp del payload en milisegundos

`timestamp` (y `entry[].time`) vienen en milisegundos epoch. Se convierten con `time.UnixMilli(ms).UTC()` y se usan como fecha de recepción/edición. Se usa el timestamp del payload (no `time.Now()`), porque el dominio ordena por ese timestamp.

### 5. Sender con cliente HTTP del stdlib

`meta.MessengerSender` construye el JSON `{recipient:{id}, messaging_type:"RESPONSE", message:{text}}` y hace `POST https://graph.facebook.com/{graphAPIVersion}/{META_PAGE_ID}/messages?access_token={META_PAGE_ACCESS_TOKEN}`, respetando `ctx`. Extrae `message_id` de la respuesta 200 y lo retorna; ante error de red o HTTP no-2xx retorna error (el caso de uso lo marca `failed`). La versión de Graph es una constante de código (`graphAPIVersion`), no configuración: subirla es un despliegue deliberado y no depende de env.

- **Alternativa**: tipificar códigos de error de Meta (190, 1545041…). Se descarta: el dominio ya colapsa todo error a `failed`; los códigos solo importan en logs.

### 6. Ruta pública en la raíz

El handler se registra en el router raíz (fuera del grupo `/v1`) con `router.Get("/webhooks/meta", ...)` y `router.Post("/webhooks/meta", ...)`. No lleva `Auth`. El proxy de nginx es responsabilidad del front y queda fuera de alcance.

### 7. Se elimina `StubSender`

Se borra `infra/messagesender/message_sender.go` y su única referencia (`app.go`). Sin `META_PAGE_ACCESS_TOKEN`, el envío falla y el mensaje queda `failed` — comportamiento asumido. Las variables `META_*` se documentan en `backend/.env.example`.

### 8. Meta como proveedor, `object` como producto

`/webhooks/meta` es el webhook del **proveedor** Meta: un mismo endpoint recibe los productos que Meta envía, identificados por el campo `object` (`page`, `instagram`, `whatsapp_business_account`). Hoy el handler solo procesa `object == "page"` (Messenger y, con Page, Instagram) y responde 200 al resto. El sender usa el Send API de Page (`/{PAGE_ID}/messages`), compartido por Messenger e Instagram. WhatsApp, aunque es de Meta, es otro producto con su propio endpoint y token: cuando se implemente será otra implementación de `domain.AgentMessageSender`, no un reemplazo de este sender. Telegram es otro proveedor con su propio contrato y paquete.

### 9. Proxy en `web`

El `web` (nginx) proxya `/webhooks/` al servicio `api`, igual que `/v1/`, porque la callback vive en la raíz y no bajo `/v1`. La app Vue no cambia; solo se añade una `location` en `nginx.conf`.

### 10. Compose de producción con base de datos incluida

`docker-compose.prod.yml` añade un servicio `postgres` (`postgres:18-alpine`, UTC, volumen `pgdata`, healthcheck) y cablea `api` a él (`POSTGRES_HOST=postgres`, `POSTGRES_PORT=5432`, `depends_on` con healthcheck). Un servicio `migrate` (perfil `tools`) aplica migraciones contra esa base, y el `Makefile` expone `prod-migrate`. Si se define `POSTGRES_URL`, el `api` mantiene su prioridad para apuntar a una base externa.

## Risks / Trade-offs

- **[Reloj]** Se compara/compara poco; la fecha de recepción sale del payload de Meta, no del reloj local → sin riesgo de skew para el orden.
- **[Latencia vs timeout]** El `Timeout(5s)` global envuelve el `POST /messages` del agente. → El sender propaga `ctx` y usa un cliente con timeout ≤ el del contexto.
- **[Reintento hasta desuscribir]** Un error permanente responde 500 y Meta reintenta durante ~1h hasta desuscribir el webhook. → Se registra; se acepta porque no se distingue por tipo (Meta no garantiza el orden).
- **[Dev sin token]** Borrar `StubSender` rompe el envío en dev sin `META_PAGE_ACCESS_TOKEN`. → Se documenta en `.env.example`; asumido.

## Migration Plan

1. Desplegar el backend con las variables `META_*` presentes.
2. En el App Dashboard de Meta, fijar el Callback URL a `/webhooks/meta` con el verify token y suscribir los campos `messages` y `message_edits`.
3. Sin migración de BD ni cambios de esquema.
4. Rollback: desplegar la imagen anterior; el `StubSender` ya no existe, pero no hay datos que migrar.
