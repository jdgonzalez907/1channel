# webhook-simulator Specification

## Purpose

Herramienta de desarrollo, integrada en la consola del agente, que simula eventos del contacto contra `POST /v1/webhooks/1channel` para crear y modificar conversaciones sin depender de un canal real.

## Requirements

### Requirement: Acceso al simulador desde la consola

La consola SHALL ofrecer una acción que abra el simulador de webhook, alcanzable aunque no haya una conversación seleccionada. El simulador SHALL enviar los eventos a `POST /v1/webhooks/1channel`, que es público y no requiere `Authorization`.

#### Scenario: Abrir el simulador sin conversación seleccionada

- **WHEN** el agente abre el simulador desde la consola sin haber seleccionado una conversación
- **THEN** el simulador se muestra listo para recibir sus campos

#### Scenario: Envío sin token de agente

- **WHEN** el simulador envía un evento
- **THEN** la solicitud se hace sin `Authorization` y el servidor la acepta por ser un endpoint público

### Requirement: Campos del evento simulado

El simulador SHALL ofrecer la selección del evento entre `message.received`, `message.edited`, `message.deleted` y `message.read`. SHALL incluir `external_contact_id` y `external_message_id`. El campo de texto SHALL mostrarse y SHALL ser obligatorio únicamente para `message.received` y `message.edited`.

#### Scenario: Evento de recepción o edición

- **WHEN** el agente selecciona `message.received` o `message.edited`
- **THEN** el simulador muestra el campo de texto como obligatorio junto con el contacto y el identificador externo del mensaje

#### Scenario: Evento de eliminación o lectura

- **WHEN** el agente selecciona `message.deleted` o `message.read`
- **THEN** el simulador oculta el campo de texto y solo pide el contacto y el identificador externo del mensaje

#### Scenario: Identificador externo de mensaje manual

- **WHEN** el agente necesita simular `message.edited`, `message.deleted` o `message.read` sobre un mensaje existente
- **THEN** el simulador le permite escribir a mano el `external_message_id`, porque la API de lectura no expone el identificador externo de los mensajes

### Requirement: Prellenado del contacto

Cuando haya una conversación seleccionada y el agente abra el simulador, el simulador SHALL prellenar `external_contact_id` con el identificador externo del contacto de esa conversación. Cuando no haya selección, el campo SHALL quedar vacío y editable.

#### Scenario: Prellenado con selección

- **WHEN** el agente abre el simulador con una conversación seleccionada
- **THEN** el campo de contacto llega prellenado con el `external_id` de esa conversación

#### Scenario: Sin selección

- **WHEN** el agente abre el simulador sin conversación seleccionada
- **THEN** el campo de contacto está vacío y se puede escribir

### Requirement: Efecto del evento y refresco

Tras un envío exitoso (204), el simulador SHALL refrescar la lista, y SHALL refrescar el detalle cuando el contacto del evento corresponda a la conversación abierta. Un `message.received` con un contacto nuevo SHALL producir una conversación nueva visible en el listado del estado correspondiente.

#### Scenario: Recepción de un contacto nuevo

- **WHEN** el simulador envía `message.received` con un `external_contact_id` que no existía
- **THEN** tras el 204 la lista se refresca y muestra la conversación nueva del contacto

#### Scenario: Recepción sobre un contacto existente

- **WHEN** el simulador envía `message.received` con un contacto existente
- **THEN** el mensaje se adjunta a su conversación activa y la lista y el detalle del contacto abierto se refrescan

### Requirement: Errores del evento simulado

El simulador SHALL mostrar el error del servidor (título y detalle de `application/problem+json`) cuando el envío falle, incluyendo 422 por campo requerido ausente, 404 por mensaje inexistente y 403 por mensaje de otro contacto, y SHALL conservar los valores ingresados.

#### Scenario: Campo requerido ausente

- **WHEN** el simulador envía un evento sin un campo requerido por ese evento
- **THEN** muestra el error 422 y conserva los valores del formulario

#### Scenario: Mensaje inexistente

- **WHEN** el simulador envía `message.edited`, `message.deleted` o `message.read` con un `external_message_id` inexistente
- **THEN** muestra el error 404 y conserva los valores del formulario
