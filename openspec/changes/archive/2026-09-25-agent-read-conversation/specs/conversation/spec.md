# Spec Delta

## ADDED Requirements

### Requirement: Agente puede leer conversacion
El sistema SHALL permitir a un agente marcar mensajes del contacto como leídos en una conversacion existente. El agente solo puede marcar como leídos mensajes que el contacto envió.

#### Scenario: Marcar mensajes como leídos exitosamente
- **WHEN** un agente lee una conversacion asignada a el con mensajes del contacto sin leer
- **THEN** los mensajes del contacto se marcan como leídos con timestamp y status actualizados

#### Scenario: Fallar si el agente no es el asignado
- **WHEN** un agente intenta leer una conversacion asignada a otro agente
- **THEN** el sistema retorna ErrConversationAgentNotOwner

#### Scenario: No marcar mensajes del agente como leídos
- **WHEN** un agente lee una conversacion con mensajes del agente y del contacto
- **THEN** solo los mensajes del contacto se marcan como leídos, los mensajes del agente no cambian

#### Scenario: Marcar mensajes en conversacion finalizada
- **WHEN** un agente lee una conversacion en estado expired o resolved
- **THEN** los mensajes del contacto se marcan como leídos normalmente

#### Scenario: No marcar mensajes ya leídos
- **WHEN** un agente lee una conversacion con mensajes del contacto ya leídos
- **THEN** los mensajes ya leídos no cambian, solo se marcan los no leídos

### Requirement: Mensaje tiene exactamente un dueño
El sistema SHALL exigir que todo mensaje tenga exactamente un dueño: o el agente o el contacto, nunca ambos ni ninguno.

#### Scenario: Crear mensaje del agente
- **WHEN** se crea un mensaje con agentID y sin contactID
- **THEN** el mensaje se crea exitosamente

#### Scenario: Crear mensaje del contacto
- **WHEN** se crea un mensaje con contactID y sin agentID
- **THEN** el mensaje se crea exitosamente

#### Scenario: Fallar si mensaje no tiene dueño
- **WHEN** se intenta crear un mensaje sin agentID y sin contactID
- **THEN** el sistema retorna ErrMessageInvalidOwner

#### Scenario: Fallar si mensaje tiene dos dueños
- **WHEN** se intenta crear un mensaje con agentID y contactID
- **THEN** el sistema retorna ErrMessageInvalidOwner

### Requirement: Marcar mensaje como leído
El sistema SHALL permitir marcar un mensaje como leído, cambiando su status a read y asignando el timestamp de lectura.

#### Scenario: Marcar mensaje como leído
- **WHEN** se marca un mensaje como leído con un timestamp
- **THEN** el mensaje queda con status read y readAt asignado