# Tasks

## 1. Configuración del canal Meta

- [x] 1.1 Crear `backend/internal/modules/conversations/infra/meta/config.go` con la estructura `Config` (`PageID`, `PageAccessToken`, `AppSecret`, `VerifyToken`) y `NewConfigFromEnv()` que las lea desde `META_PAGE_ID`, `META_PAGE_ACCESS_TOKEN`, `META_APP_SECRET` y `META_VERIFY_TOKEN`. Verificar con `go -C backend build ./...`.
- [x] 1.2 Añadir test unitario de `NewConfigFromEnv` (con valores y con valores ausentes). Verificar con `go -C backend test ./internal/modules/conversations/infra/meta/...`.
- [x] 1.3 Documentar las variables `META_*` en `backend/.env.example`. Verificar revisando el diff del archivo.

## 2. Sender de Meta

- [x] 2.1 Implementar `backend/internal/modules/conversations/infra/meta/messenger_sender.go` con `MessengerSender` que implemente `domain.AgentMessageSender` (con aserción `var _ domain.AgentMessageSender = (*MessengerSender)(nil)`): `POST https://graph.facebook.com/{graphAPIVersion}/{PageID}/messages?access_token={PageAccessToken}` con cuerpo `{recipient:{id:to}, messaging_type:"RESPONSE", message:{text}}`, retornando `message_id` como `external_id` y propagando `ctx`. Verificar con `go -C backend build ./...`.
- [x] 2.2 Añadir test unitario del sender con `httptest.Server` que verifique método, ruta/query, cuerpo JSON y retorno de `message_id`; y un caso de error con respuesta HTTP no-2xx. Verificar con `go -C backend test ./internal/modules/conversations/infra/meta/...`.

## 3. Handler del webhook

- [x] 3.1 Crear `backend/internal/modules/conversations/infra/meta/webhook.go` con el handler `MetaWebhookHandler` (dependencias: `app.ReceiveContactMessage`, `app.ReceiveContactMessageEdit`, `Config`) y su `Register(r chi.Router)` con `GET`/`POST /webhooks/meta`. Verificar con `go -C backend build ./...`.
- [x] 3.2 Implementar la verificación `GET`: `hub.mode=subscribe` y `hub.verify_token==VerifyToken` → 200 con `hub.challenge`; en cualquier otro caso 403. Añadir test (token correcto/incorrecto, modo incorrecto). Verificar con `go -C backend test ./internal/modules/conversations/infra/meta/...`.
- [x] 3.3 Implementar la validación de firma: leer el body crudo, calcular `HMAC-SHA256(body, AppSecret)` y comparar con `X-Hub-Signature-256`; sin cabecera o sin coincidencia → 401. Añadir test (firma válida/inválida/ausente). Verificar con `go -C backend test ./internal/modules/conversations/infra/meta/...`.
- [x] 3.4 Implementar el mapeo de eventos: `messages` con `text` y sin `is_echo` → `ReceiveContactMessage` (`sender.id`→contacto, `mid`→mensaje, `timestamp` ms→UTC); `message_edit` → `ReceiveContactMessageEdit`. Añadir test de ambos mapeos. Verificar con `go -C backend test ./internal/modules/conversations/infra/meta/...`.
- [x] 3.5 Ignorar con 200 los eventos no soportados y las notificaciones con `object != page` (`is_echo`, `attachments`, `read`, `delivery`, postbacks, `reply_to`). Añadir test de ignorado. Verificar con `go -C backend test ./internal/modules/conversations/infra/meta/...`.

## 4. Wiring y limpieza

- [x] 4.1 Registrar la ruta pública en `backend/cmd/api/router.go` fuera del grupo `/v1` (`deps.metaWebhookHandler.Register(router)`). Verificar con `go -C backend build ./...`.
- [x] 4.2 En `backend/cmd/api/app.go`, construir `MessengerSender` y `MetaWebhookHandler` con `NewConfigFromEnv()`, inyectarlos en `dependencies` y reemplazar `messagesender.NewStubSender()` por el sender real. Verificar con `go -C backend build ./...`.
- [x] 4.3 Eliminar el paquete `backend/internal/modules/conversations/infra/messagesender/` (y su referencia) al no quedar usos. Verificar con `go -C backend build ./...` y `gofmt -l backend` limpio.

## 5. Verificación integral

- [x] 5.1 Ejecutar las compuertas de calidad: `go -C backend mod verify && gofmt -l backend && go -C backend vet ./... && go -C backend build ./... && go -C backend test -race -count=1 ./...`. Verificar que todo pase.

## 6. Proxy del webhook

- [x] 6.1 Añadir `location /webhooks/` en `frontend/nginx.conf` que proxee a `http://api:8080` con las mismas cabeceras que `/v1/`. Verificar revisando el diff del archivo.

## 7. Compose de producción con postgres

- [x] 7.1 Añadir el servicio `postgres` (volumen persistente, UTC, healthcheck) y cablear `api` a él en `docker-compose.prod.yml`; añadir el servicio `migrate` bajo el perfil `tools`. Verificar con `docker compose -f docker-compose.prod.yml config`.
- [x] 7.2 Actualizar el `Makefile` (comentario de producción y target `prod-migrate`). Verificar con `make help`.

## 8. Documentación

- [x] 8.1 Documentar `GET/POST /webhooks/meta` en `backend/docs/endpoints.md`. Verificar revisando el diff.
- [x] 8.2 Actualizar `AGENTS.md` (variables `META_*`, webhook de Meta, compose de producción con postgres, proxy `/webhooks`). Verificar revisando el diff.
