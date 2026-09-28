# Spec Delta

## ADDED Requirements

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
