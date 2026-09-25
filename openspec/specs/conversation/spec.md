# Conversation Module

## Purpose

El modulo de conversaciones permite a los contactos iniciar conversaciones con la empresa para soporte, ventas, cambios, envios y otras interacciones cortas. Cada conversacion tiene un solo agente activo y puede expirar segun la configuracion de la plataforma.

## Requirements

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

### Requirement: Obtener mensajes de conversacion
El sistema SHALL retornar los mensajes de una conversacion como un slice.

#### Scenario: Obtener mensajes
- **WHEN** se solicitan los mensajes de una conversacion
- **THEN** se retorna un slice con todos los mensajes encontrados

### Requirement: Estado de conversacion
El sistema SHALL soportar los estados: pending, assigned, expired, resolved.

#### Scenario: Estado inicial
- **WHEN** se crea una conversacion
- **THEN** su estado es pending o assigned segun el agente asignado

### Requirement: Estado de mensaje
El sistema SHALL soportar los estados de mensaje: sent, read, deleted.

#### Scenario: Estado de mensaje nuevo
- **WHEN** se crea un mensaje
- **THEN** su estado es sent

### Requirement: Tipo de mensaje
El sistema SHALL soportar el tipo de mensaje text para MVP.

#### Scenario: Mensaje de texto
- **WHEN** se crea un mensaje de tipo text
- **THEN** el mensaje tiene un campo text con contenido

### Requirement: Crear mensaje
El sistema SHALL permitir crear un nuevo mensaje con todos sus campos requeridos.

#### Scenario: Crear mensaje completo
- **WHEN** se crea un mensaje con id, status, type, text, agentID, contactID, sentAt, readAt, editedAt, deletedAt
- **THEN** se retorna un mensaje con todos los campos asignados

#### Scenario: Fallar al crear mensaje con uuid invalido
- **WHEN** se intenta crear un mensaje con un id Nil
- **THEN** el sistema retorna ErrMessageInvalidID

#### Scenario: Fallar al crear mensaje con status invalido
- **WHEN** se intenta crear un mensaje con un status que no es sent, read o deleted
- **THEN** el sistema retorna ErrMessageStatusInvalid

#### Scenario: Fallar al crear mensaje con type invalido
- **WHEN** se intenta crear un mensaje con un type que no es text
- **THEN** el sistema retorna ErrMessageTypeInvalid

#### Scenario: Fallar al crear mensaje con texto vacio para tipo text
- **WHEN** se intenta crear un mensaje de tipo text con texto nil o vacio
- **THEN** el sistema retorna ErrMessageEmptyText

#### Scenario: Fallar al crear mensaje con texto muy largo
- **WHEN** se intenta crear un mensaje de tipo text con mas de 1000 runas
- **THEN** el sistema retorna ErrMessageTextTooLong

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
