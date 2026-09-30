# Proposal

## Why

El sistema solo puede recibir mensajes de contacto por el endpoint de prueba `POST /v1/webhooks/1channel` y envía las respuestas del agente por un `StubSender` que inventa un identificador externo. No existe ningún canal real: no se reciben clientes de Messenger ni se entregan las respuestas a Meta. Este cambio integra el primer canal real (Messenger/Meta) para recibir mensajes y editarlos, y para enviar respuestas del agente a través del Send API.

## What Changes

- Nuevo endpoint público de webhook en la raíz `GET/POST /webhooks/meta` que:
  - Resuelve la verificación de Meta (`hub.mode=subscribe` + `hub.verify_token` → 200 `hub.challenge`, o 403).
  - Valida la firma `X-Hub-Signature-256` (HMAC-SHA256 del body con `META_APP_SECRET`).
  - Traduce `messages` (texto, sin `is_echo`) y `message_edits` a los casos de uso existentes `ReceiveContactMessage` / `ReceiveContactMessageEdit`.
  - Ignora con 200 el resto de eventos (read, delivery, echo, attachments, postbacks, reactions, reply_to).
- Nuevo sender de Meta que implementa `domain.AgentMessageSender` y llama al Send API (`messaging_type: RESPONSE`), devolviendo `message_id` como `external_id`.
- Se elimina `StubSender` (`infra/messagesender`). Sin `META_PAGE_ACCESS_TOKEN` el envío falla y el mensaje queda `failed`.
- Nuevas variables de configuración `META_PAGE_ID`, `META_PAGE_ACCESS_TOKEN`, `META_APP_SECRET`, `META_VERIFY_TOKEN`.
- El `web` (nginx) proxya `/webhooks/` al servicio `api` para exponer la callback de Meta, que vive en la raíz y no bajo `/v1`.
- El compose de producción (`docker-compose.prod.yml`) incluye su propio `postgres` (volumen persistente) y un servicio `migrate` bajo perfil de herramientas; el `api` se conecta a él por variables de entorno.

## Capabilities

### New Capabilities

- `meta-channel`: integración del canal Messenger/Meta — webhook de entrada (verificación, firma, mapeo de eventos `messages`/`message_edits`) y sender de salida (Send API `RESPONSE` que devuelve el `external_id`).

### Modified Capabilities

- `delivery`: el enrutado por el mismo origen pasa a incluir `/webhooks/*` además de `/v1/*`, y el compose de producción pasa a incluir su propio `postgres` en vez de depender solo de una base externa.

## Impact

- `backend/internal/modules/conversations/infra/meta/` (nuevo): `config.go`, `webhook.go`, `write_dto.go`, `read_dto.go`, `messenger_sender.go` + tests unitarios.
- `backend/internal/modules/conversations/infra/messagesender/` (eliminado).
- `backend/cmd/api/app.go`, `backend/cmd/api/router.go` (wiring y registro de la ruta pública).
- `frontend/nginx.conf` (nueva `location /webhooks/`).
- `backend/.env.example` (documentar las variables `META_*`).
- `backend/docs/endpoints.md`, `AGENTS.md` (documentar el webhook y actualizar despliegue/variables).
- `docker-compose.prod.yml`, `Makefile` (postgres propio y target `prod-migrate`).
- Sin cambios en `domain` ni `app`: los casos de uso de contacto ya existen y se reutilizan tal cual.
