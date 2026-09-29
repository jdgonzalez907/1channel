# Spec Delta

## ADDED Requirements

### Requirement: Consultar una persona por su documento

El sistema SHALL exponer `GET /v1/personal-information/{identification_number}` para que un agente autenticado consulte una persona por su documento (clave natural). La respuesta exitosa SHALL ser 200 con el identificador y los datos de la persona. Un documento vacío, con caracteres no permitidos o que exceda 100 grafemas SHALL responder 400; cuando la persona no exista SHALL responder 404.

#### Scenario: Persona encontrada

- **WHEN** el agente consulta un `identification_number` que existe
- **THEN** el sistema responde 200 con el identificador y los datos de la persona

#### Scenario: Documento inválido

- **WHEN** el agente consulta un `identification_number` vacío, con caracteres no permitidos o que supera 100 grafemas
- **THEN** el sistema responde 400

#### Scenario: Persona inexistente

- **WHEN** el agente consulta un `identification_number` bien formado que no existe
- **THEN** el sistema responde 404

### Requirement: Etiqueta del contacto y persona en las lecturas de conversación

La bandeja y el detalle de conversación SHALL exponer la etiqueta de presentación del contacto junto con su identificador interno y externo. El detalle SHALL exponer además la persona asociada al contacto como un objeto anidado `personal_information` con sus datos y fechas (`created_at`, `updated_at`), nulo cuando el contacto no tenga persona.

#### Scenario: Bandeja con etiqueta

- **WHEN** el agente lista la bandeja
- **THEN** cada elemento incluye la etiqueta de presentación de su contacto

#### Scenario: Detalle con persona asociada

- **WHEN** el agente abre una conversación cuyo contacto tiene persona
- **THEN** el detalle incluye el objeto anidado `personal_information` con sus datos y fechas

#### Scenario: Detalle sin persona asociada

- **WHEN** el agente abre una conversación cuyo contacto no tiene persona
- **THEN** el objeto `personal_information` del detalle es nulo

## MODIFIED Requirements

### Requirement: Consultar un contacto por identificador

El sistema SHALL exponer `GET /v1/contacts/{id}` para que un agente autenticado obtenga un contacto por su identificador. La respuesta exitosa SHALL ser 200 con el identificador, el identificador externo, la etiqueta de presentación, el `display_name`, la persona asociada como objeto anidado `personal_information` (nulo si no tiene) y la fecha de creación. Si el `id` no es válido SHALL responder 400 y si el contacto no existe SHALL responder 404.

#### Scenario: Contacto existente

- **WHEN** el agente consulta un contacto existente
- **THEN** el sistema responde 200 con su identificador, identificador externo, etiqueta de presentación, `display_name`, persona asociada o su ausencia, y fecha de creación

#### Scenario: Contacto inexistente o identificador inválido

- **WHEN** el agente consulta un contacto que no existe o con un `id` inválido
- **THEN** el sistema responde 404 o 400 respectivamente
