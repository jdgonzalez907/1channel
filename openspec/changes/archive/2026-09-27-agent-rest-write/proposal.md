# Proposal

## Why

El sistema ya tiene dominio y casos de uso para agentes y conversaciones, pero no existe capa de entrega: no hay entrypoint, router ni handlers, así que ningún cliente puede operar el sistema. Además la identidad del operador está modelada como `agents`, un nombre atado a conversaciones, cuando en realidad es el usuario del sistema que otros módulos reutilizarán.

## What Changes

- Agregar una API REST de escritura con chi v5.3.2:
  - `POST /v1/users`
  - `POST /v1/conversations/{id}/messages`
  - `PATCH /v1/conversations/{id}/messages`
  - `PATCH /v1/conversations/{id}`
- **BREAKING** renombrar la identidad del sistema de `agents` a `users`: módulo Go `internal/modules/agents` -> `internal/modules/users`, tabla SQL `agents` -> `users`, queries sqlc, repositorio y API pública. La tabla es compartida: significa "agente" dentro de conversaciones y "usuario" en cualquier otro módulo.
- Actualizar migraciones: `conversations.agent_id` -> `user_id` y `messages.agent_id` -> `user_id`, referenciando `users`.
- Agregar el caso de uso `CreateUser` con su `Save` de repositorio y query sqlc.
- Agregar middleware de autenticación genérico por Bearer (ID de usuario) y respuestas de error RFC 7807.
- Agregar timeout de request (5s) con mapeo a `504`, timeouts de servidor (`ReadHeaderTimeout`/`WriteTimeout`/`IdleTimeout`) y apagado graceful.
- Cablear un emisor saliente stub temporal hasta implementar el canal externo real.
- La resolución por HTTP solo acepta `resolved`; un agente no puede expirar (eso queda fuera del endpoint).
- Fuera de alcance: endpoints GET, read models y cualquier lectura de feature.

## Capabilities

### New Capabilities

- `users`: identidad de los usuarios del sistema; API pública para validar existencia, crear y rehidratar un usuario.
- `http-api`: contrato REST de escritura para operar usuarios y conversaciones (rutas, autenticación Bearer, códigos de estado y envelope de errores).

### Modified Capabilities

- `agents`: capability retirada; sus requisitos pasan a `users` al renombrar el módulo.

`persistence` no lleva delta: el rename de tabla y columnas no cambia ninguna restricción ni comportamiento observable, es un refactor de implementación.

## Impact

- **Nuevo código**: `cmd/api` (composition root, router, servidor), `internal/modules/users` (renombrado), handlers HTTP por módulo, `internal/modules/conversations/infra/messagesender` (emisor stub), `internal/shared/infra/http` (middleware, `httputil` y writer de errores).
- **Base de datos**: `db/migrations/000001` (tabla `users`), `db/queries/users`, `internal/shared/infra/pgdb/sqlc` regenerado.
- **Contratos existentes**: `conversations` cambia sus FKs y el mapper `users.id <-> Agent`; el término "agent" se conserva únicamente dentro de conversaciones.
- **Dependencias**: se incorpora `github.com/go-chi/chi/v5` v5.3.2.
- **Riesgo**: el rename toca specs, migraciones y sqlc; al no haber producción, se editan en sitio en vez de encadenar migraciones nuevas.
