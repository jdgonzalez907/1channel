# Spec Delta

## ADDED Requirements

### Requirement: Caso de uso: marcar como leído un mensaje del agente por el contacto

El sistema SHALL marcar como leído un mensaje del agente a partir de un acuse de lectura del contacto, localizándolo por su identificador externo dentro de la conversación. El sistema no SHALL rechazar el acuse por el estado finalizado de la conversación. Los acuses con timestamp anterior o igual a la última lectura registrada SHALL descartarse sin error. El sistema SHALL rechazar el acuse si el contacto no pertenece a la conversación o si el mensaje no fue enviado por un agente. El acuse no SHALL reactivar un mensaje eliminado: SHALL registrar la lectura conservando su estado eliminado. Un mensaje fallido SHALL permanecer sin cambios.

#### Scenario: Lectura aplicada

- **WHEN** el contacto envía un acuse de lectura de un mensaje del agente no leído
- **THEN** el mensaje queda con su timestamp de lectura y estado read

#### Scenario: Acuse obsoleto

- **WHEN** el contacto envía un acuse con timestamp anterior o igual al último registrado
- **THEN** el mensaje conserva su lectura más reciente y no se retorna error

#### Scenario: Identificador externo inexistente

- **WHEN** el contacto envía un acuse con un identificador externo que no pertenece a la conversación
- **THEN** el sistema retorna ErrMessageNotFound

#### Scenario: Mensaje no pertenece al agente

- **WHEN** el contacto envía un acuse sobre un mensaje enviado por el propio contacto
- **THEN** el sistema retorna ErrConversationContactNotOwner

#### Scenario: Lectura en conversación finalizada

- **WHEN** el contacto envía un acuse sobre una conversación finalizada
- **THEN** la lectura se aplica igualmente

#### Scenario: Acuse sobre mensaje eliminado

- **WHEN** el contacto envía un acuse sobre un mensaje del agente eliminado
- **THEN** la lectura se registra y el mensaje conserva su estado eliminado

#### Scenario: Acuse sobre mensaje fallido

- **WHEN** el contacto envía un acuse sobre un mensaje del agente fallido
- **THEN** el mensaje permanece sin cambios y sin error
