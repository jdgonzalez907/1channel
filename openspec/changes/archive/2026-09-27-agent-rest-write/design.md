# Design

## Context

Ver `proposal.md - Why`. El estado actual relevante:

- No existe capa HTTP: `cmd/api/` está vacío, no hay router, handlers ni `main`.
- El módulo `internal/modules/agents` solo expone `FindAgentByID`; `domain.NewAgent` existe pero sin validación y sin uso.
- Los casos de uso de conversaciones (`SendAgentMessage`, `AgentReadConversation`, `ResolveConversation`) ya existen y validan el usuario llamando a `agents.AgentsAPI.FindAgentByID`.
- `internal/shared/infra/pgdb` centraliza pool + sqlc; las migraciones son `000001..000004`.
- Go 1.27 (stdlib `uuid`), SQLC, PostgreSQL, testify para tests.

## Goals / Non-Goals

**Goals:**

- Entregar la capa REST de escritura con chi v5.3.2, handlers por módulo y composition root en `cmd/api`.
- Unificar la identidad en una sola tabla/módulo `users`, reutilizable por cualquier módulo.
- Autenticación por Bearer (ID de usuario) con un middleware genérico, sin acoplarse al término "agente".
- Mapeo uniforme de errores a HTTP.

**Non-Goals:**

- Endpoints GET, read models, `read.Queries`, joins vs múltiples queries y `dto.go` de respuestas: change posterior.
- JWT, firma HMAC del webhook de Meta, CORS, rate limiting, versionado por header.
- Persistencia nueva más allá de crear usuarios.

## Decisions

### D1 - chi v5.3.2 como router

`github.com/go-chi/chi/v5` v5.3.2 (última release, ago-2026). Alternativa: `net/http` con `ServeMux`; se descarta por el ecosistema de middlewares y `Group`/`Route` por grupo. Las rutas se montan por grupos para aplicar auth selectivamente.

### D2 - Middleware auth genérico, sin `Resolver`

La auth es infraestructura compartida, no del módulo. No se introduce una interfaz `Resolver` (sobreingeniería): un único middleware genérico en `internal/shared/infra/http/middleware` parsea `Authorization: Bearer <uuid>`, lo deja en el contexto bajo una clave genérica (`UserIDFrom(ctx)`) y responde 401 si falta o no es parseable. Cuando llegue JWT, cambia la implementación interna del mismo middleware, no los puntos de montaje.

La existencia del usuario la validan los propios casos de uso de escritura (ya llaman a `UsersAPI.FindUserByID`). Si el usuario no existe, el handler mapea el error del dominio de `users` a 401. Alternativa considerada: que el middleware consulte la tabla para un 401 parejo; se descarta para no meter una lectura de feature en este cambio.

### D3 - Una sola tabla/módulo `users`, "agent" solo en conversaciones

- Tabla `users(id, created_at)` reemplaza a `agents`.
- `conversations.agent_id` -> `user_id` y `messages.agent_id` -> `user_id`, FK a `users`.
- `internal/modules/agents` -> `internal/modules/users`: `UsersAPI`, `domain.User`, `UserRepository`, queries `FindUserByID`/`CreateUser` sobre `users`.
- Dentro de `conversations` se conserva el lenguaje `Agent`/`agentID`; el mapper `users.id <-> Agent.ID()` vive en `conversations/infra/pg`.

Alternativa considerada: mantener `agents` y duplicar identidad por módulo; descartada porque la misma persona opera varios módulos.

### D4 - Rename de esquema editando en sitio

Como no hay producción, se editan las migraciones existentes (`000001_create_agents` -> `000001_create_users`, `000003`/`000004` FKs) y se regenera sqlc. Alternativa: migración nueva encima; descartada por ruido innecesario pre-producción.

### D5 - Writes por app/domain, reads directas (CQRS ligero)

Los cuatro endpoints son escritura y pasan por `app -> domain -> repository`. El read side queda separado para el change de GETs: `handler -> read.Queries -> infra/pg/read`, sin app ni domain. No se mezclan en este cambio.

### D6 - Handlers por módulo con inyección por constructor

```
internal/modules/users/infra/http/         user_handler.go + dto.go
internal/modules/conversations/infra/http/ conversation_handler.go + dto.go
internal/modules/conversations/infra/messagesender/ message_sender.go (stub)
internal/shared/infra/http/middleware/     auth.go + logging.go + timeout.go
internal/shared/infra/http/httperror/      httperror.go
internal/shared/infra/http/httputil/       httputil.go
cmd/api/{main,config,app,router}.go        composition root
```

Cada handler recibe sus casos de uso por constructor y expone `Register(r chi.Router)`. `cmd/api` construye pool -> repos -> use cases -> apis -> handlers -> router y monta los grupos.

### D7 - DTOs de request en el handler, respuestas mínimas

Request DTOs en el package de handlers, con `http.MaxBytesReader` y `DisallowUnknownFields`. El handler valida el `status` permitido y arma el input del caso de uso. Respuestas: `201 + {"id"}` para creaciones (sin `Location`, porque los recursos creados no son dereferenciables en este cambio), `204` para PATCH.

### D8 - Errores RFC 7807, mapeo por módulo

Writer compartido `httperror.Write` con constructores por código (`httperror.BadRequest`, `httperror.NotFound`, `httperror.Unprocessable`, ...) y `httperror.Generic` para lo transversal. El cuerpo omite `type` (el RFC 9457 asume `about:blank`) y completa `instance` con el request id (chi `RequestID`). El switch de errores vive en cada handler (conoce sus errores de dominio y usa `errors.Is`, que atraviesa los `errors.Join` de los casos de uso):

```
400 entrada malformada / UUID invalido / content-type
401 sin Bearer, Bearer invalido, usuario inexistente
403 ErrConversationAgentNotOwner
404 ErrConversationNotFound
409 ErrConversationFinished / NotAcceptingMessages
422 Errores de validacion de dominio y status no permitido
504 context.DeadlineExceeded (timeout del request)
500 default
```

### D9 - Timestamps del servidor

El handler fija `time.Now().UTC()` para `SentAt`/`ReadAt`/`ResolvedAt`/`CreatedAt`. Solo el webhook entrante tomará el tiempo de la petición (fuera de este cambio).

### D10 - Semántica de PATCH

`PATCH /conversations/{id}/messages` con `{"status":"read"}` y `PATCH /conversations/{id}` con `{"status":"resolved"}`. El status es atributo real de cada recurso. `expired` no se acepta (422): un agente no expira, la expiración es del sistema. Alternativas: `POST .../resolve` y `POST .../read` (RPC) descartadas por menos REST.

### D11 - Middlewares

```
GLOBAL:  RequestID, Logging (slog), Recoverer, CleanPath/StripSlashes, Timeout (5s propio), /healthz
GRUPO:   Auth (Bearer)
SERVER:  ReadHeaderTimeout 5s, WriteTimeout 10s, IdleTimeout 60s, shutdown grace 10s
```

`Timeout` es propio (no el de chi): solo cancela el contexto; el 504 lo escribe el handler vía `httperror.Generic`. `RealIP` solo con proxy confiable; se difiere. Sin CORS/rate limit en este cambio.

### D12 - Emisor saliente stub

`SendAgentMessage` requiere `AgentMessageSender`. Como el canal externo (Meta/WhatsApp) todavía no existe, se cablea un `StubSender` en `internal/modules/conversations/infra/messagesender` que devuelve un `uuid v7` como identificador externo, para poder ejercitar el flujo HTTP. Se reemplaza por el canal real en un change posterior.

## Risks / Trade-offs

- [Rename tocando Go, SQL y specs] -> No hay producción; se edita en sitio y se corre la suite completa. Verificación: `go build ./...` + `go test -race ./...` + `openspec validate`.
- [`POST /v1/users` público para bootstrap] -> Riesgo de alta no autorizada. Mitigación: endurecer con secreto de admin en un change posterior; queda registrado en el spec `http-api`.
- [sqlc regenerado a mano puede desincronizarse] -> Regenerar con `sqlc generate` y revisar que `models.go` nombre `User` y que `FindUserByID`/`CreateUser` apunten a `users`.
- [Doble lectura de usuario: middleware sin validar + caso de uso] -> Es intencional para no acoplar; si molesta después, se mueve la validación al middleware sin cambiar specs.
- [Lecturas fuera de alcance] -> Los GET no existen todavía; el read model y la decisión joins vs multiples queries quedan para el siguiente change.

## Migration Plan

1. Editar migraciones y queries; `sqlc generate`.
2. Renombrar módulo y ajustar `conversations` (FKs, mapper).
3. Agregar create user (domain/app/repo).
4. Agregar shared (middleware, httperror) y handlers por módulo.
5. Agregar `cmd/api` con router y wiring.
6. Rollback: al no haber producción, revertir commit y recrear la BD con migraciones `down`.
