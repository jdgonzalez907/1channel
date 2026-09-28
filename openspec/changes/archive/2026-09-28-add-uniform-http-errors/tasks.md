# Tasks

## 1. Middleware de recuperación propio

- [x] 1.1 Crear `recover.go` en `shared/infra/http/middleware` con `Recover(logger *slog.Logger) func(http.Handler) http.Handler`: recupera el panic, registra con `slog` (método, path, `request_id`, panic y `debug.Stack()`), responde 500 vía `httperror` y re-paniquea `http.ErrAbortHandler`; verificar con `go -C backend build ./...`
- [x] 1.2 Agregar tests del middleware: un handler que paniquea responde 500 `application/problem+json` y deja una entrada de log; un handler que paniquea con `http.ErrAbortHandler` propaga el panic; verificar con `go -C backend test ./internal/shared/infra/http/middleware/...`

## 2. Handlers de ruta inexistente y método no permitido

- [x] 2.1 Agregar en `httperror` un helper para el 405 (sin header `Allow`) reutilizando `Problem`; verificar con `go -C backend build ./...`
- [x] 2.2 Registrar `router.NotFound` y `router.MethodNotAllowed` en `newRouter` para responder 404 y 405 con `application/problem+json`; verificar con `go -C backend build ./...`

## 3. Cableado global

- [x] 3.1 Reemplazar `chimiddleware.Recoverer` por `middleware.Recover(logger)` en `cmd/api/router.go` (misma posición, después de `Logging`) y ajustar imports; verificar con `go -C backend build ./...`

## 4. Tests del router

- [x] 4.1 En `cmd/api/router_test.go` agregar casos: ruta inexistente responde 404 `application/problem+json` (no `text/plain`) y método no permitido responde 405 `application/problem+json`; verificar con `go -C backend test ./cmd/api/...`

## 5. Documentación

- [x] 5.1 Actualizar la tabla de errores de `backend/docs/endpoints.md` con `405` (método no permitido) y aclarar que una ruta inexistente también responde `404` en `problem+json`; verificar que el texto coincide con el comportamiento

## 6. Quality gates

- [x] 6.1 Ejecutar `go -C backend mod verify && gofmt -l backend && go -C backend vet ./...` y verificar que `gofmt` no reporta archivos
- [x] 6.2 Ejecutar `go -C backend build ./...` y verificar que compila
- [x] 6.3 Ejecutar `go -C backend test -race -count=1 ./...` y verificar que todos los tests pasan
