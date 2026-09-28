# Design

## Context

Ver `proposal.md - Why`. Estado actual de `cmd/api/router.go`:

```go
router.Use(chimiddleware.RequestID)
router.Use(sharedmiddleware.Logging(logger))
router.Use(chimiddleware.Recoverer)      // chi: 500 vacío + PrintPrettyStack
router.Use(chimiddleware.CleanPath)
router.Use(chimiddleware.StripSlashes)
router.Use(sharedmiddleware.Timeout(5 * time.Second))
```

- `httperror` ya produce el contrato `application/problem+json` (`Problem`, `Write`, helpers) y toma el `instance` del request id.
- `Logging` usa `slog` y no setea un log entry de chi, así que el `Recoverer` de chi cae al `PrintPrettyStack` (ANSI a stderr, sin `request_id`).
- chi devuelve 404 con `http.NotFound` (`text/plain`) y 405 sin cuerpo; el `Mux` permite sobreescribirlos con `NotFound(...)` / `MethodNotAllowed(...)`.
- `CleanPath` y `StripSlashes` de chi v5.3.2 reescriben la ruta, no redirigen.

## Goals / Non-Goals

**Goals:**
- Que 404, 405 y 500 inesperado salgan en `application/problem+json`.
- Que el panic se recupere y se registre con el logger de la aplicación.
- Mantener la recuperación global (toda la superficie) y no caer el proceso.

**Non-Goals:**
- No se cambian los errores de dominio de `http-api` / `http-read-api`.
- No se cambia `Timeout` (sigue basado en contexto; su 504 sale por `httperror.Generic` si el handler respeta el `DeadlineExceeded`).
- No se agrega logging a los caminos felices ni se rediseña el access log.

## Decisions

### 1. Middleware de recuperación propio en `shared/infra/http/middleware`

Se reemplaza `chimiddleware.Recoverer` por `middleware.Recover(logger)`. Recupera, registra con `slog` (método, path, `request_id`, panic y `debug.Stack()`) y responde 500 vía `httperror`. Se descarta el `Recoverer` de chi porque fija el 500 sin cuerpo y loguea por fuera del logger. Alternativa: setear un chi log entry en `Logging`; descartada porque igual deja el 500 sin `problem+json`.

### 2. NotFound y MethodNotAllowed problem+json

En `newRouter`, `router.NotFound` y `router.MethodNotAllowed` escriben con `httperror` (404 y 405). El 405 no incluye el header `Allow` porque chi no expone los métodos permitidos al handler propio (`methodsAllowed` es interno); se prioriza el contrato `problem+json` sobre el header. Se agrega un helper en `httperror` para el 405 reutilizando `Problem`.

### 3. Orden y alcance global

El middleware de recuperación se registra en la misma posición que el `Recoverer` actual (después de `Logging`, antes de `CleanPath`), de modo que el access log siga registrando el status final. `NotFound`/`MethodNotAllowed` se registran en el `Mux` raíz, cubriendo todo. El panic log queda separado del access log: el primero con stack, el segundo con el status.

### 4. Preservar `http.ErrAbortHandler`

El middleware re-paniquea `http.ErrAbortHandler` para no convertir un abort intencional en 500, igual que chi.

## Risks / Trade-offs

- **Respuesta ya iniciada** -> Si el handler ya escribió headers antes del panic, no se puede reescribir el status; se registra el panic y se intenta escribir el body. Aceptable: en este código los handlers escriben al final.
- **Doble registro** -> Un panic genera la línea del access log (status 500) y la línea del panic (stack); son complementarias, no duplicadas.
- **`Allow` en 405** -> chi no expone los métodos permitidos al `MethodNotAllowedHandler` propio (`methodsAllowed` es interno), así que se responde 405 `problem+json` sin `Allow`. Decisión aceptada: se prioriza el contrato de error uniforme; el cliente es de primera parte.
- **Tests** -> No hay DB en CI; se cubren con `httptest` un handler que paniquea y rutas inexistentes/no permitidas.
