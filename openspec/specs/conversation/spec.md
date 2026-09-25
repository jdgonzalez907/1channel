# Conversation Module

## Purpose

El modulo de conversaciones permite a los contactos iniciar conversaciones con la empresa para soporte, ventas, cambios, envios y otras interacciones cortas. Cada conversacion tiene un solo agente activo y puede expirar segun la configuracion de la plataforma.

## Requirements

### Requirement: Crear conversacion
El sistema SHALL permitir crear una nueva conversacion con todos sus campos requeridos.

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
