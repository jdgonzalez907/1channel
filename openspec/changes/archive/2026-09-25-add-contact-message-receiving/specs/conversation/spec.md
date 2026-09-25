# Spec Delta

## ADDED Requirements

### Requirement: Recibir mensaje de contacto
El sistema SHALL permitir recibir un mensaje de un contacto en una conversacion existente, validando que el contacto pertenezca a la conversacion y que el mensaje no este duplicado.

#### Scenario: Recibir mensaje exitosamente
- **WHEN** se recibe un mensaje de un contacto que pertenece a la conversacion
- **THEN** el mensaje se agrega a la conversacion y el timestamp de actualizacion se actualiza

#### Scenario: Fallar si el contacto no pertenece a la conversacion
- **WHEN** se recibe un mensaje de un contacto que no es el asignado a la conversacion
- **THEN** el sistema retorna ErrConversationContactNotOwner

#### Scenario: Fallar si la conversacion no tiene contacto
- **WHEN** se recibe un mensaje pero la conversacion no tiene un contacto asignado
- **THEN** el sistema retorna ErrConversationHasNoContact

#### Scenario: Fallar si el mensaje ya existe (duplicado por externalID)
- **WHEN** se recibe un mensaje con un externalID que ya existe en la conversacion
- **THEN** el sistema retorna ErrConversationDuplicateMessage

#### Scenario: Recibir mensaje sin externalID
- **WHEN** se recibe un mensaje que no tiene externalID asignado
- **THEN** el mensaje se agrega a la conversacion sin validar duplicados por externalID

### Requirement: Identificador externo de mensaje
El sistema SHALL permitir asignar un identificador externo a un mensaje que aun no tiene uno. Una vez asignado, el identificador es inmutable.

#### Scenario: Asignar externalID exitosamente
- **WHEN** se asigna un externalID a un mensaje que no tiene uno
- **THEN** el mensaje queda con el externalID asignado

#### Scenario: Fallar si el externalID ya fue asignado
- **WHEN** se intenta asignar un externalID a un mensaje que ya tiene uno
- **THEN** el sistema retorna ErrMessageExternalIDAlreadySet

#### Scenario: Fallar si el externalID es vacio
- **WHEN** se intenta asignar un externalID con valor vacio
- **THEN** el sistema retorna ErrMessageExternalIDInvalid

### Requirement: Conversacion requiere al menos un mensaje
El sistema SHALL exigir que toda conversacion tenga al menos un mensaje. No es posible crear una conversacion sin mensajes.

#### Scenario: Fallar al crear conversacion sin mensajes
- **WHEN** se intenta crear una conversacion con un slice de mensajes vacio o nil
- **THEN** el sistema retorna ErrConversationEmptyMessages

#### Scenario: Crear conversacion con un mensaje
- **WHEN** se crea una conversacion con al menos un mensaje
- **THEN** la conversacion se crea exitosamente con el mensaje almacenado
