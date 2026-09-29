# Spec Delta

## ADDED Requirements

### Requirement: Nombre de presentación del contacto

El contacto SHALL tener un `display_name` opcional. La etiqueta de presentación de un contacto SHALL resolverse como el nombre completo de su persona asociada (`first_name` y `last_name`) cuando existe; en su defecto, el `display_name`; en su defecto, el `external_contact_id`.

#### Scenario: Contacto con persona asociada

- **WHEN** un contacto está asociado a una persona
- **THEN** su etiqueta de presentación es el nombre completo de la persona

#### Scenario: Contacto sin persona con display_name

- **WHEN** un contacto no tiene persona asociada pero tiene `display_name`
- **THEN** su etiqueta de presentación es el `display_name`

#### Scenario: Contacto sin persona ni display_name

- **WHEN** un contacto no tiene persona asociada ni `display_name`
- **THEN** su etiqueta de presentación es el `external_contact_id`

### Requirement: Nombre del contacto provisto por el canal

El `display_name` del contacto SHALL setearse a partir del dato que el canal envía en el webhook y SHALL aplicarse únicamente con el evento `message.received`. Cuando el evento de recepción traiga un `display_name`, el contacto (existente o recién creado) SHALL quedar con ese valor; cuando no lo traiga, el valor existente SHALL conservarse. Los eventos distintos de `message.received` SHALL NOT modificar el `display_name`.

#### Scenario: Recepción con nombre

- **WHEN** llega un `message.received` con `display_name`
- **THEN** el contacto queda con ese `display_name`

#### Scenario: Recepción sin nombre

- **WHEN** llega un `message.received` sin `display_name`
- **THEN** el `display_name` existente del contacto no cambia

#### Scenario: Otro evento

- **WHEN** llega un evento distinto de `message.received`
- **THEN** el `display_name` del contacto no cambia

## MODIFIED Requirements

### Requirement: Rehidratar contacto

El módulo de contactos SHALL permitir reconstruir un contacto a partir de datos persistidos sin validar sus invariantes, aceptando un identificador externo vacío. La rehidratación SHALL conservar el `display_name` y la referencia a la persona, que pueden estar ausentes. La rehidratación SHALL ser distinta de la creación, que sí valida el identificador externo.

#### Scenario: Rehidratar contacto con identificador externo

- **WHEN** se rehidrata un contacto con identificador, identificador externo y fecha de creación
- **THEN** se obtiene el contacto con esos valores

#### Scenario: Rehidratar contacto con identificador externo vacío

- **WHEN** se rehidrata un contacto con identificador externo vacío
- **THEN** se obtiene el contacto sin validar el identificador externo

#### Scenario: Rehidratar contacto con persona y display_name

- **WHEN** se rehidrata un contacto con `display_name` y referencia a una persona
- **THEN** se obtiene el contacto con esos valores

#### Scenario: Rehidratar contacto sin persona

- **WHEN** se rehidrata un contacto sin referencia a persona
- **THEN** se obtiene el contacto con la referencia ausente
