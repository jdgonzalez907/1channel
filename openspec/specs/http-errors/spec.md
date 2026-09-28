# http-errors Specification

## Purpose

Define el contrato uniforme de respuestas de error de toda la API: formato `application/problem+json` con el identificador del request, y el tratamiento y registro de los fallos de ruteo y de los fallos inesperados.

## Requirements

### Requirement: Respuestas de error uniformes en toda la API

El sistema SHALL responder toda respuesta de error (4xx y 5xx) con `Content-Type: application/problem+json` y un cuerpo con al menos `title`, `status` y `detail`, e `instance` con el identificador del request. Esto SHALL aplicar también a las peticiones a rutas inexistentes, a los métodos no permitidos y a los fallos inesperados. El sistema SHALL mapear los códigos: 400 para entrada malformada, 401 para autenticación, 403 para falta de propiedad, 404 para recurso o ruta inexistente, 405 para método no permitido, 409 para conflictos de estado, 422 para validación, 500 para fallos inesperados y 504 para timeout del request.

#### Scenario: Formato de error

- **WHEN** cualquier respuesta de error se produce
- **THEN** usa `application/problem+json` con `title`, `status` y `detail`

#### Scenario: Ruta inexistente

- **WHEN** llega una petición a una ruta que no está registrada
- **THEN** el sistema responde 404 con `application/problem+json`, no `text/plain`

#### Scenario: Método no permitido

- **WHEN** llega una petición con un método no permitido para una ruta registrada
- **THEN** el sistema responde 405 con `application/problem+json`

#### Scenario: Fallo inesperado

- **WHEN** un handler termina en un fallo inesperado
- **THEN** el sistema responde 500 con `application/problem+json`, no con cuerpo vacío

#### Scenario: Identificador del request

- **WHEN** se produce una respuesta de error
- **THEN** su `instance` coincide con el identificador del request

### Requirement: Recuperación de panics

El sistema SHALL recuperarse de un panic producido en cualquier handler o middleware, SHALL NOT terminar el proceso y SHALL responder 500 con el formato de error uniforme. Un abort de conexión (`http.ErrAbortHandler`) SHALL propagarse sin convertirse en 500.

#### Scenario: Panic en un handler

- **WHEN** un handler entra en panic
- **THEN** el sistema responde 500 con `application/problem+json` y continúa sirviendo peticiones

#### Scenario: Abort de conexión

- **WHEN** se produce `http.ErrAbortHandler`
- **THEN** el sistema no lo convierte en 500 y deja abortar la conexión

### Requirement: Registro de fallos inesperados

El sistema SHALL registrar todo fallo inesperado (panic recuperado) con el logger de la aplicación, incluyendo el método, el path, el identificador del request y el stack. SHALL usar el mismo logger que el resto de la aplicación.

#### Scenario: Panic registrado

- **WHEN** un panic es recuperado
- **THEN** el log incluye el método, el path, el identificador del request y el stack, emitido por el logger de la aplicación
