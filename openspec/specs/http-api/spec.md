# http-api Specification

## Purpose

Contrato REST de escritura para que un usuario del sistema autenticado opere conversaciones y para dar de alta usuarios.

## Requirements

### Requirement: Identidad del solicitante por Bearer

Las solicitudes a endpoints autenticados SHALL obtener la identidad del usuario desde el encabezado `Authorization: Bearer <id>`. El identificador SHALL provenir siempre del encabezado y nunca del cuerpo ni de la ruta. El sistema SHALL rechazar con 401 las solicitudes sin encabezado, con un valor no parseable o cuyo usuario no exista.

#### Scenario: Solicitud autenticada

- **WHEN** llega una solicitud con `Authorization: Bearer <uuid>` de un usuario existente
- **THEN** la operación se ejecuta con la identidad de ese usuario

#### Scenario: Falta el encabezado

- **WHEN** llega una solicitud autenticada sin encabezado `Authorization`
- **THEN** el sistema responde 401

#### Scenario: Bearer no parseable

- **WHEN** llega una solicitud con un Bearer que no es un identificador válido
- **THEN** el sistema responde 401

#### Scenario: Usuario inexistente

- **WHEN** llega una solicitud con un Bearer cuyo usuario no existe
- **THEN** el sistema responde 401

### Requirement: Crear usuario del sistema por HTTP

El sistema SHALL exponer `POST /v1/users` para dar de alta un usuario del sistema. El identificador y la fecha de creación SHALL generarse en el servidor. La respuesta exitosa SHALL ser 201 con el identificador del usuario creado en el cuerpo.

#### Scenario: Creación exitosa

- **WHEN** llega `POST /v1/users`
- **THEN** se crea el usuario y el sistema responde 201 con su identificador

#### Scenario: Alta de bootstrap

- **WHEN** llega `POST /v1/users` sin autenticación
- **THEN** el sistema lo acepta para permitir el alta del primer usuario

### Requirement: Enviar mensaje de agente por HTTP

El sistema SHALL exponer `POST /v1/conversations/{id}/messages` para que un usuario autenticado envíe un mensaje de texto a una conversación, usando su identidad como agente. La fecha de envío SHALL generarse en el servidor. La respuesta exitosa SHALL ser 201 con el identificador del mensaje creado en el cuerpo.

#### Scenario: Envío exitoso

- **WHEN** un usuario autenticado envía `{"text":"..."}` a una conversación válida
- **THEN** el mensaje se registra y el sistema responde 201 con su identificador

#### Scenario: Identificador de conversación inválido

- **WHEN** el `id` de la ruta no es un identificador válido
- **THEN** el sistema responde 400

#### Scenario: Conversación inexistente

- **WHEN** la conversación no existe
- **THEN** el sistema responde 404

#### Scenario: Agente no asignado

- **WHEN** el usuario autenticado no es el agente asignado a la conversación
- **THEN** el sistema responde 403

#### Scenario: Conversación que no acepta mensajes

- **WHEN** la conversación está finalizada y no acepta el mensaje
- **THEN** el sistema responde 409

#### Scenario: Texto inválido

- **WHEN** el texto está vacío o excede el máximo permitido
- **THEN** el sistema responde 422

### Requirement: Marcar mensajes del contacto como leídos por HTTP

El sistema SHALL exponer `PATCH /v1/conversations/{id}/messages` con cuerpo `{"status":"read"}` para que el agente marque como leídos los mensajes del contacto. La fecha de lectura SHALL generarse en el servidor. La respuesta exitosa SHALL ser 204. Cualquier otro valor de `status` SHALL rechazarse con 422.

#### Scenario: Marcado exitoso

- **WHEN** un usuario autenticado envía `{"status":"read"}` sobre una conversación asignada a él
- **THEN** los mensajes del contacto se marcan como leídos y el sistema responde 204

#### Scenario: Status no permitido

- **WHEN** el cuerpo trae un `status` distinto de `read`
- **THEN** el sistema responde 422

#### Scenario: Agente no asignado

- **WHEN** el usuario autenticado no es el agente asignado
- **THEN** el sistema responde 403

### Requirement: Resolver conversación por HTTP

El sistema SHALL exponer `PATCH /v1/conversations/{id}` con cuerpo `{"status":"resolved"}` para que el agente asignado resuelva la conversación. La fecha de resolución SHALL generarse en el servidor. La respuesta exitosa SHALL ser 204. Un agente SHALL NOT poder expirar una conversación: `expired` y cualquier otro valor SHALL rechazarse con 422.

#### Scenario: Resolución exitosa

- **WHEN** el agente asignado envía `{"status":"resolved"}`
- **THEN** la conversación queda resuelta y el sistema responde 204

#### Scenario: Agente no puede expirar

- **WHEN** el cuerpo trae `{"status":"expired"}`
- **THEN** el sistema responde 422 y la conversación no cambia

#### Scenario: Agente no asignado

- **WHEN** el usuario autenticado no es el agente asignado
- **THEN** el sistema responde 403

### Requirement: Webhook de contacto de prueba por HTTP

El sistema SHALL exponer `POST /v1/webhooks/1channel`, sin autenticación, para que un canal de prueba active las operaciones de contacto. El cuerpo SHALL ser JSON e incluir `event`, `external_contact_id` y `external_message_id` obligatorios, y `text` cuando el evento lo requiera. Los eventos soportados SHALL ser `message.received`, `message.edited`, `message.deleted` y `message.read`. El timestamp de la operación SHALL generarlo el servidor en UTC y no SHALL requerirse en el cuerpo. La respuesta exitosa SHALL ser 204 sin cuerpo; un `event` no soportado SHALL ignorarse y responder 200 sin alterar nada. El sistema SHALL rechazar con 422 la ausencia de un campo requerido por el evento, y con 400 un cuerpo JSON malformado. La aplicación de cada operación SHALL respetar las reglas de los casos de uso de contacto, mapeando sus errores de dominio a los códigos HTTP definidos para la API de escritura, y SHALL aplicarse aunque la conversación esté finalizada.

#### Scenario: Recepción de mensaje del contacto

- **WHEN** llega `{"event":"message.received","external_contact_id":"...","external_message_id":"...","text":"..."}`
- **THEN** el mensaje se adjunta a la conversación activa del contacto (o se inicia una nueva) y el sistema responde 204

#### Scenario: Edición de mensaje del contacto

- **WHEN** llega `{"event":"message.edited","external_contact_id":"...","external_message_id":"...","text":"..."}`
- **THEN** la edición se aplica al mensaje del contacto localizado por su identificador externo y el sistema responde 204

#### Scenario: Eliminación de mensaje del contacto

- **WHEN** llega `{"event":"message.deleted","external_contact_id":"...","external_message_id":"..."}`
- **THEN** el mensaje del contacto queda eliminado de forma idempotente y el sistema responde 204

#### Scenario: Acuse de lectura del contacto

- **WHEN** llega `{"event":"message.read","external_contact_id":"...","external_message_id":"..."}`
- **THEN** el mensaje del agente localizado por su identificador externo queda marcado como leído y el sistema responde 204

#### Scenario: Aplicación sobre conversación finalizada

- **WHEN** llega un evento de edición, eliminación o lectura sobre una conversación finalizada
- **THEN** el sistema aplica la operación igualmente y responde 204

#### Scenario: Endpoint abierto

- **WHEN** llega `POST /v1/webhooks/1channel` sin encabezado `Authorization`
- **THEN** el sistema procesa la solicitud sin exigir autenticación

#### Scenario: Evento desconocido

- **WHEN** el cuerpo trae un `event` vacío o distinto de los soportados
- **THEN** el sistema responde 200 y no altera ninguna conversación

#### Scenario: Campo requerido faltante

- **WHEN** el evento requiere `external_contact_id`, `external_message_id` o `text` y alguno está ausente
- **THEN** el sistema responde 422 y no altera ninguna conversación

#### Scenario: Cuerpo malformado

- **WHEN** el cuerpo no es JSON válido
- **THEN** el sistema responde 400

#### Scenario: Error de dominio de la operación

- **WHEN** la operación de contacto falla por un recurso inexistente o por falta de propiedad sobre el mensaje
- **THEN** el sistema responde 404 o 403 respectivamente, usando `application/problem+json`

### Requirement: Respuestas de error uniformes

El sistema SHALL responder los errores con `Content-Type: application/problem+json` y un cuerpo con al menos `title`, `status` y `detail`, e `instance` con el identificador del request. Los errores de dominio SHALL mapearse a los códigos: 400 para entrada malformada, 401 para autenticación, 403 para falta de propiedad, 404 para recurso inexistente, 409 para conflictos de estado, 422 para validación, 500 para fallos internos y 504 para timeout del request.

#### Scenario: Formato de error

- **WHEN** una operación falla
- **THEN** la respuesta usa `application/problem+json` con `title`, `status` y `detail`

#### Scenario: Error de validación

- **WHEN** la operación falla por validación de dominio
- **THEN** el sistema responde 422

#### Scenario: Error de propiedad

- **WHEN** un agente opera una conversación que no le pertenece
- **THEN** el sistema responde 403

#### Scenario: Timeout del request

- **WHEN** la operación excede el tiempo máximo del request
- **THEN** el sistema responde 504
