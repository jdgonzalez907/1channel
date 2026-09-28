# Proposal

## Why

La API no responde siempre su contrato de error: las rutas inexistentes (404), los métodos no permitidos (405) y los panics (500) usan los defaults de chi, que devuelven `text/plain` o cuerpo vacío en lugar de `application/problem+json`. Además, el stack de un panic se imprime con el formateador propio de chi (ANSI a stderr) y no pasa por el logger de la aplicación. Un error inesperado no debería romper ni el contrato HTTP ni el logging.

## What Changes

- Nueva capability `http-errors`: contrato uniforme de errores HTTP para toda la superficie del API, independiente de los endpoints de escritura y lectura.
- Handler propio para rutas inexistentes (`404`) que responde `application/problem+json`.
- Handler propio para método no permitido (`405`, con `Allow`) que responde `application/problem+json`.
- Middleware propio de recuperación de panics que reemplaza al `Recoverer` de chi: registra el panic con el logger de la aplicación (método, path, `request_id`, stack) y responde `500` en `application/problem+json`.
- Se preserva el comportamiento de `http.ErrAbortHandler` (no se convierte en 500; se deja abortar la conexión).
- El resto de la cadena global (`RequestID`, `Logging`, `CleanPath`, `StripSlashes`, `Timeout`) no cambia; los nuevos handlers quedan globales en el router raíz.
- Sin cambios en los errores de dominio de `http-api` / `http-read-api`, que siguen siendo la fuente de sus mapeos.

## Capabilities

### New Capabilities

- `http-errors`: contrato uniforme de respuestas de error de toda la API (formato `application/problem+json`, `instance` con el request id, recuperación de panics y su registro), incluidos los casos de ruteo (404/405) y los fallos inesperados (500).

### Modified Capabilities

_(ninguna)_

## Impact

- `backend/internal/shared/infra/http/httperror/` — reutilizado para 404/405/500; posible helper para el 405 con `Allow`.
- `backend/internal/shared/infra/http/middleware/` — nuevo middleware de recuperación (reemplaza al de chi) con logging por `slog`.
- `backend/cmd/api/router.go` — registrar `NotFound`/`MethodNotAllowed` problem+json y el middleware de recuperación propio; quitar `chimiddleware.Recoverer`.
- Tests de `middleware` y de `cmd/api/router_test.go` para 404/405/500 y panic.
- No se toca `domain`, `app` ni los handlers de módulos.
