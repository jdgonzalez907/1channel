# Spec Delta

## Purpose

Integra el canal Messenger de Meta: recibe mensajes y ediciones de contacto a través del webhook y envía las respuestas del agente mediante el Send API de Meta.

## ADDED Requirements

### Requirement: Verificación del webhook

El sistema SHALL exponer `GET /webhooks/meta` para validar la solicitud de verificación de Meta. Cuando `hub.mode` sea `subscribe` y `hub.verify_token` coincida con `META_VERIFY_TOKEN`, SHALL responder 200 con el valor de `hub.challenge` como cuerpo. En cualquier otro caso SHALL responder 403.

#### Scenario: Verificación exitosa

- **WHEN** llega `GET /webhooks/meta` con `hub.mode=subscribe` y un `hub.verify_token` que coincide con `META_VERIFY_TOKEN`
- **THEN** el sistema responde 200 con el valor de `hub.challenge` como cuerpo

#### Scenario: Token de verificación incorrecto

- **WHEN** llega una verificación con `hub.verify_token` distinto de `META_VERIFY_TOKEN`
- **THEN** el sistema responde 403

#### Scenario: Modo de verificación incorrecto

- **WHEN** llega una verificación con `hub.mode` distinto de `subscribe`
- **THEN** el sistema responde 403

### Requirement: Validación de la firma

El sistema SHALL validar la autenticidad de cada notificación verificando la cabecera `X-Hub-Signature-256` contra el HMAC-SHA256 del cuerpo calculado con `META_APP_SECRET`. Una notificación sin cabecera o con firma que no coincide SHALL rechazarse con 401 y no SHALL alterar conversaciones.

#### Scenario: Firma válida

- **WHEN** llega un `POST /webhooks/meta` cuya firma `X-Hub-Signature-256` coincide con el HMAC-SHA256 del cuerpo
- **THEN** el sistema procesa la notificación

#### Scenario: Firma inválida o ausente

- **WHEN** llega un `POST /webhooks/meta` sin cabecera `X-Hub-Signature-256` o con una firma que no coincide
- **THEN** el sistema responde 401 y no altera ninguna conversación

### Requirement: Recepción de mensajes de texto

El sistema SHALL traducir cada evento `messages` cuyo `message` contiene `text` y no contiene `is_echo` a un mensaje de contacto, usando `sender.id` como identificador externo del contacto, `message.mid` como identificador externo del mensaje y `timestamp` en milisegundos como fecha de recepción, y SHALL responder 200.

#### Scenario: Mensaje de texto entrante

- **WHEN** llega un evento `messages` con `message.text` y sin `is_echo`
- **THEN** el mensaje se adjunta a la conversación activa del contacto (o inicia una nueva) y el sistema responde 200

### Requirement: Edición de mensajes

El sistema SHALL traducir un evento `message_edit` a una edición del mensaje de contacto localizado por `message_edit.mid`, aplicando el nuevo `message_edit.text`, y SHALL responder 200.

#### Scenario: Edición de mensaje

- **WHEN** llega un evento `message_edit` con `mid` y `text`
- **THEN** la edición se aplica al mensaje localizado por su identificador externo y el sistema responde 200

### Requirement: Eventos no soportados se ignoran

El sistema SHALL responder 200 e ignorar, sin alterar conversaciones, cualquier notificación que no sea un `messages` de texto ni un `message_edit`. Esto incluye notificaciones con `object` distinto de `page`, mensajes con `is_echo`, `attachments`, acuses de lectura, entregas, postbacks, reacciones y `reply_to`.

#### Scenario: Evento no soportado

- **WHEN** llega una notificación que no es un `messages` de texto ni un `message_edit`
- **THEN** el sistema responde 200 y no altera ninguna conversación

### Requirement: Manejo de fallos y reintentos

El sistema SHALL procesar cada evento del lote y SHALL responder 500 si alguno falla, para que Meta reintente la notificación; SHALL responder 200 si todos se procesan. La respuesta SHALL depender solo de si el procesamiento tuvo éxito, no del tipo de error.

#### Scenario: Fallo de procesamiento

- **WHEN** el procesamiento de un evento del lote falla
- **THEN** el sistema registra el fallo y responde 500 para que Meta reintente

#### Scenario: Lote procesado

- **WHEN** todos los eventos del lote se procesan correctamente
- **THEN** el sistema responde 200

### Requirement: Envío del mensaje del agente por Send API

El canal Meta SHALL enviar los mensajes de texto del agente mediante el Send API con `messaging_type: RESPONSE`, dirigidos al `recipient.id` igual al identificador externo del contacto, y SHALL retornar el `message_id` de la respuesta como identificador externo del mensaje.

#### Scenario: Envío exitoso

- **WHEN** se envía un mensaje del agente y el Send API responde con `message_id`
- **THEN** el canal retorna ese `message_id` como identificador externo del mensaje

#### Scenario: Fallo del Send API

- **WHEN** la llamada al Send API falla
- **THEN** el canal retorna un error para que el mensaje quede marcado como fallido
