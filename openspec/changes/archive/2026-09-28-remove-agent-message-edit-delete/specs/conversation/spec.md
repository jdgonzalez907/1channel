# Spec Delta

## REMOVED Requirements

### Requirement: Editar mensaje de texto

**Reason**: La edición por parte del agente se elimina; solo el contacto edita a través del canal. Las reglas de validación de la edición quedan en el caso de uso del contacto.

**Migration**: Reemplazado por `Caso de uso: editar mensaje de contacto`, que concentra las reglas de edición y la validación del mensaje.

### Requirement: Eliminar mensaje

**Reason**: La eliminación por parte del agente se elimina; solo el contacto elimina a través del canal. Las reglas de eliminación quedan en el caso de uso del contacto.

**Migration**: Reemplazado por `Caso de uso: eliminar mensaje de contacto`, que concentra las reglas de eliminación.

## MODIFIED Requirements

### Requirement: Caso de uso: editar mensaje de contacto

El sistema SHALL aplicar la edición de un mensaje enviada por el contacto, localizándolo por su identificador externo dentro de la conversación. Un agente no SHALL poder editar mensajes. Las ediciones con timestamp anterior o igual a la última edición aplicada SHALL descartarse sin error. El sistema no SHALL rechazar la edición por el estado finalizado de la conversación. El sistema SHALL validar el contenido: solo mensajes de tipo texto, texto no vacío, como máximo 1000 caracteres, y mensajes no fallidos.

#### Scenario: Edición aplicada

- **WHEN** el contacto edita un mensaje suyo y el timestamp es más reciente que la última edición
- **THEN** el texto y el timestamp de edición del mensaje se actualizan

#### Scenario: Edición obsoleta

- **WHEN** el contacto edita un mensaje con timestamp anterior o igual a la última edición aplicada
- **THEN** el mensaje conserva su edición más reciente y no se retorna error

#### Scenario: Identificador externo inexistente

- **WHEN** el contacto edita un mensaje con un identificador externo que no pertenece a la conversación
- **THEN** el sistema retorna ErrMessageNotFound

#### Scenario: Mensaje no pertenece al contacto

- **WHEN** el contacto intenta editar un mensaje enviado por el agente
- **THEN** el sistema retorna ErrConversationContactNotOwner

#### Scenario: Edición en conversación finalizada

- **WHEN** el contacto edita un mensaje en una conversación finalizada
- **THEN** la edición se aplica igualmente

#### Scenario: Editar un mensaje sin texto

- **WHEN** el contacto edita un mensaje que no tiene texto
- **THEN** el sistema retorna ErrMessageNotText

#### Scenario: Editar con texto vacío

- **WHEN** el contacto edita un mensaje con texto vacío
- **THEN** el sistema retorna ErrMessageEmptyText

#### Scenario: Editar con texto que excede el máximo

- **WHEN** el contacto edita un mensaje con texto de más de 1000 caracteres
- **THEN** el sistema retorna ErrMessageTextTooLong

#### Scenario: Editar un mensaje eliminado con timestamp posterior

- **WHEN** el contacto edita un mensaje eliminado con timestamp posterior a deletedAt
- **THEN** el sistema retorna ErrMessageAlreadyDeleted

#### Scenario: Editar un mensaje eliminado con timestamp anterior

- **WHEN** el contacto edita un mensaje eliminado con timestamp anterior a deletedAt
- **THEN** la edición se aplica como traza del estado anterior

#### Scenario: Editar un mensaje fallido

- **WHEN** el contacto edita un mensaje en estado failed
- **THEN** el sistema retorna ErrMessageFailed

### Requirement: Caso de uso: eliminar mensaje de contacto

El sistema SHALL aplicar la eliminación (soft delete) de un mensaje enviada por el contacto, localizándolo por su identificador externo. Un agente no SHALL poder eliminar mensajes. Un mensaje ya eliminado SHALL tratarse como idempotente (sin error). El sistema no SHALL rechazar la eliminación por el estado finalizado de la conversación. El sistema SHALL rechazar eliminar un mensaje en estado failed.

#### Scenario: Eliminación aplicada

- **WHEN** el contacto elimina un mensaje suyo no eliminado
- **THEN** el mensaje queda con timestamp de eliminación y estado deleted

#### Scenario: Eliminación repetida es idempotente

- **WHEN** el contacto elimina un mensaje ya eliminado
- **THEN** el sistema no altera el mensaje y no retorna error

#### Scenario: Identificador externo inexistente

- **WHEN** el contacto elimina un mensaje con un identificador externo que no pertenece a la conversación
- **THEN** el sistema retorna ErrMessageNotFound

#### Scenario: Mensaje no pertenece al contacto

- **WHEN** el contacto intenta eliminar un mensaje enviado por el agente
- **THEN** el sistema retorna ErrConversationContactNotOwner

#### Scenario: Eliminación en conversación finalizada

- **WHEN** el contacto elimina un mensaje en una conversación finalizada
- **THEN** la eliminación se aplica igualmente

#### Scenario: Eliminar un mensaje fallido

- **WHEN** el contacto elimina un mensaje en estado failed
- **THEN** el sistema retorna ErrMessageFailed
