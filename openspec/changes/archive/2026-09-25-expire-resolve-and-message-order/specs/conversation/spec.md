# Spec Delta

## ADDED Requirements

### Requirement: Expirar conversacion
El sistema SHALL permitir cerrar una conversacion por expiracion, asignando su fecha de finalizacion. Puede expirar una conversacion en estado pending o assigned. Una conversacion ya finalizada no puede expirarse.

#### Scenario: Expirar conversacion pending
- **WHEN** se expira una conversacion en estado pending
- **THEN** el estado cambia a expired y finishedAt se asigna

#### Scenario: Expirar conversacion assigned
- **WHEN** se expira una conversacion en estado assigned
- **THEN** el estado cambia a expired y finishedAt se asigna

#### Scenario: Fallar si la conversacion ya esta finalizada
- **WHEN** se intenta expirar una conversacion en estado expired o resolved
- **THEN** el sistema retorna ErrConversationFinished

#### Scenario: Expirar registra actividad
- **WHEN** se expira una conversacion
- **THEN** updatedAt se actualiza si el timestamp es mas reciente

### Requirement: Resolver conversacion
El sistema SHALL permitir cerrar una conversacion porque el agente la resolvio, asignando su fecha de finalizacion. Solo el agente asignado puede resolverla. Una conversacion ya finalizada no puede resolverse.

#### Scenario: Resolver conversacion exitosamente
- **WHEN** el agente asignado resuelve una conversacion en estado assigned
- **THEN** el estado cambia a resolved y finishedAt se asigna

#### Scenario: Fallar si el agente no es el asignado
- **WHEN** un agente intenta resolver una conversacion asignada a otro agente
- **THEN** el sistema retorna ErrConversationAgentNotOwner

#### Scenario: Fallar si la conversacion no tiene agente
- **WHEN** se intenta resolver una conversacion en estado pending sin agente asignado
- **THEN** el sistema retorna ErrConversationAgentNotOwner

#### Scenario: Fallar si la conversacion ya esta finalizada
- **WHEN** se intenta resolver una conversacion en estado expired o resolved
- **THEN** el sistema retorna ErrConversationFinished

#### Scenario: Resolver registra actividad
- **WHEN** se resuelve una conversacion
- **THEN** updatedAt se actualiza si el timestamp es mas reciente

## MODIFIED Requirements

### Requirement: Obtener mensajes de conversacion
El sistema SHALL retornar los mensajes de una conversacion como un slice ordenado por sentAt ascendente.

#### Scenario: Obtener mensajes
- **WHEN** se solicitan los mensajes de una conversacion
- **THEN** se retorna un slice con todos los mensajes ordenados por sentAt ascendente

#### Scenario: Desempatar mensajes con el mismo sentAt
- **WHEN** dos mensajes tienen el mismo sentAt
- **THEN** el orden se desempata por id de forma determinista