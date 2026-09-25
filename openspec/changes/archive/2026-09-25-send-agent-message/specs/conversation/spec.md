# Spec Delta

## ADDED Requirements

### Requirement: Enviar mensaje de agente
El sistema SHALL permitir a un agente enviar un mensaje a una conversacion existente. Si la conversacion esta pendiente sin agente asignado, el agente se asigna automaticamente y la conversacion cambia a estado assigned.

#### Scenario: Enviar mensaje exitosamente
- **WHEN** un agente envia un mensaje a una conversacion asignada a el
- **THEN** el mensaje se agrega a la conversacion y el timestamp de actualizacion se actualiza

#### Scenario: Reclamar conversacion pendiente
- **WHEN** un agente envia una mensaje a una conversacion en estado pending sin agente asignado
- **THEN** el agente se asigna a la conversacion, el estado cambia a assigned y el mensaje se agrega

#### Scenario: Fallar si el agente no es el asignado
- **WHEN** un agente intenta enviar un mensaje a una conversacion asignada a otro agente
- **THEN** el sistema retorna ErrConversationAgentNotOwner

#### Scenario: Sincronizar externalID si existe
- **WHEN** un agente envia un mensaje con externalID
- **THEN** el indice de externalID se actualiza con el nuevo mensaje

### Requirement: Validar que conversacion acepta mensajes
El sistema SHALL validar que una conversacion acepta mensajes en un momento dado, considerando su estado y fecha de finalizacion.

#### Scenario: Conversacion abierta acepta mensajes
- **WHEN** se intenta enviar o recibir un mensaje en una conversacion en estado pending o assigned
- **THEN** el sistema acepta el mensaje

#### Scenario: Conversacion cerrada rechaza mensajes tardios
- **WHEN** se intenta enviar o recibir un mensaje en una conversacion expired o resolved con timestamp posterior a finishedAt
- **THEN** el sistema retorna ErrConversationNotAcceptingMessages

#### Scenario: Conversacion cerrada acepta mensajes enviados antes de finalizar
- **WHEN** se intenta recibir un mensaje en una conversacion expired o resolved con timestamp anterior a finishedAt
- **THEN** el sistema acepta el mensaje (mensaje tardio valido)

#### Scenario: Fallar si conversacion finalizada no tiene finishedAt
- **WHEN** una conversacion tiene estado expired o resolved pero finishedAt es nil
- **THEN** el sistema retorna ErrConversationFinishedAtMissing

### Requirement: Mensajes son punteros
El sistema SHALL usar punteros a mensajes (*Message) en vez de valores (Message) para mantener la identidad de la entidad.

#### Scenario: Crear conversacion con mensajes como punteros
- **WHEN** se crea una conversacion con un slice de punteros a mensajes
- **THEN** la conversacion almacena los mensajes como punteros

#### Scenario: Recibir mensaje como puntero
- **WHEN** se recibe un mensaje de contacto como puntero
- **THEN** el mensaje se almacena como puntero en la conversacion

#### Scenario: Enviar mensaje como puntero
- **WHEN** un agente envia un mensaje como puntero
- **THEN** el mensaje se almacena como puntero en la conversacion

## MODIFIED Requirements

### Requirement: Crear conversacion
El sistema SHALL permitir crear una nueva conversacion con todos sus campos requeridos. Si el estado es expired o resolved, finishedAt es obligatorio.

#### Scenario: Crear conversacion con mensajes
- **WHEN** se crea una conversacion con id, status, mensajes, agentID, contactID, createdAt, updatedAt, finishedAt
- **THEN** se retorna una conversacion con todos los campos asignados y los mensajes almacenados internamente

#### Scenario: Fallar al crear conversacion con uuid invalido
- **WHEN** se intenta crear una conversacion con un id Nil
- **THEN** el sistema retorna ErrConversationInvalidID

#### Scenario: Fallar al crear conversacion sin contactID ni agentID
- **WHEN** se intenta crear una conversacion con contactID y agentID ambos nil
- **THEN** el sistema retorna ErrConversationMissingContactAndAgent

#### Scenario: Fallar al crear conversacion con status invalido
- **WHEN** se intenta crear una conversacion con un status que no es pending, assigned, expired o resolved
- **THEN** el sistema retorna ErrConversationStatusInvalid

#### Scenario: Fallar al crear conversacion finalizada sin finishedAt
- **WHEN** se intenta crear una conversacion con estado expired o resolved y finishedAt es nil
- **THEN** el sistema retorna ErrConversationFinishedAtMissing

### Requirement: Recibir mensaje de contacto
El sistema SHALL permitir recibir un mensaje de un contacto en una conversacion existente, validando que la conversacion acepte mensajes, que el contacto pertenezca a la conversacion y que el mensaje no este duplicado.

#### Scenario: Recibir mensaje exitosamente
- **WHEN** se recibe un mensaje de un contacto que pertenece a la conversacion
- **THEN** el mensaje se agrega a la conversacion y el timestamp de actualizacion se actualiza

#### Scenario: Fallar si la conversacion no acepta mensajes
- **WHEN** se recibe un mensaje en una conversacion que no acepta mensajes en ese momento
- **THEN** el sistema retorna ErrConversationNotAcceptingMessages

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