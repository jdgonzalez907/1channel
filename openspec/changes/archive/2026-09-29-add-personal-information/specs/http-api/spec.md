# Spec Delta

## ADDED Requirements

### Requirement: Guardar la persona de un contacto por HTTP

El sistema SHALL exponer `PUT /v1/contacts/{id}/personal-information` para que un usuario autenticado guarde la persona de un contacto. El cuerpo SHALL ser JSON e incluir `identification_number` (obligatorio, inmutable) y los datos opcionales `first_name`, `last_name`, `phone_number`, `email` y `address`. El servidor SHALL hacer upsert por `identification_number` —insertar la persona si el documento no existe, o reemplazar sus datos si existe, sin crear una fila duplicada— y luego SHALL asociar el contacto a esa persona. La respuesta exitosa SHALL ser 200 con el identificador y los datos de la persona persistida. Un `id` de ruta inválido SHALL responder 400; un cuerpo JSON malformado SHALL responder 400; un contacto inexistente SHALL responder 404; un documento ausente, vacío, con caracteres no permitidos o que exceda 100 grafemas, o un dato opcional que exceda su máximo (100 grafemas para nombre, apellido y teléfono; 254 para email y dirección), SHALL responder 422; la ausencia de autenticación SHALL responder 401.

#### Scenario: Alta y asociación de una persona nueva

- **WHEN** un usuario autenticado guarda un `identification_number` que no existe, con sus datos
- **THEN** el sistema crea la persona, asocia el contacto a ella y responde 200 con el identificador y los datos

#### Scenario: Guardado sobre un documento existente

- **WHEN** un usuario autenticado guarda un `identification_number` que ya existe
- **THEN** los datos de esa persona se reemplazan
- **AND** el contacto queda asociado a ella
- **AND** se responde 200 con el mismo identificador
- **AND** no se crea una segunda persona

#### Scenario: Solo con documento

- **WHEN** el cuerpo envía `identification_number` y omite o deja vacíos todos los datos opcionales
- **THEN** el sistema crea o actualiza la persona, asocia el contacto y responde 200

#### Scenario: Documento ausente, inválido o demasiado largo

- **WHEN** el cuerpo omite `identification_number`, o trae un documento vacío, con caracteres no permitidos o que supera 100 grafemas
- **THEN** el sistema responde 422

#### Scenario: Dato opcional que excede la longitud

- **WHEN** el cuerpo trae un dato opcional que supera su máximo de grafemas
- **THEN** el sistema responde 422

#### Scenario: Contacto inexistente

- **WHEN** el `id` de la ruta corresponde a un contacto que no existe
- **THEN** el sistema responde 404

#### Scenario: Identificador de contacto inválido

- **WHEN** el `id` de la ruta no es un identificador válido
- **THEN** el sistema responde 400

#### Scenario: Cuerpo JSON malformado

- **WHEN** el cuerpo no es JSON válido
- **THEN** el sistema responde 400

#### Scenario: Solicitud sin autenticación

- **WHEN** llega `PUT /v1/contacts/{id}/personal-information` sin un Bearer válido
- **THEN** el sistema responde 401

### Requirement: Nombre del contacto provisto por el canal en el webhook

El cuerpo de `POST /v1/webhooks/1channel` SHALL aceptar un campo opcional `display_name`. Cuando el evento sea `message.received`, el servidor SHALL propagar ese valor al contacto (existiendo o creándolo); cuando no venga, SHALL conservar el valor del contacto. Para los eventos distintos de `message.received`, el servidor SHALL ignorar `display_name`. La ausencia de `display_name` SHALL NOT rechazar la solicitud.

#### Scenario: Recepción con display_name

- **WHEN** llega `message.received` con `display_name`
- **THEN** el contacto queda con ese `display_name` y el sistema responde 204

#### Scenario: Recepción sin display_name

- **WHEN** llega `message.received` sin `display_name`
- **THEN** el contacto conserva su `display_name` y el sistema responde 204

#### Scenario: display_name en otro evento

- **WHEN** llega `message.edited`, `message.deleted` o `message.read` con `display_name`
- **THEN** el `display_name` del contacto no cambia
