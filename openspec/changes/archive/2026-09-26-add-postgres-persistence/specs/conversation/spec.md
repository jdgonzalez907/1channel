# Spec Delta

## MODIFIED Requirements

### Requirement: Búsquedas dedicadas del repositorio de conversaciones

El repositorio de conversaciones SHALL exponer búsquedas dedicadas por caso de uso que recuperan únicamente los datos que la operación necesita, incluyendo: conversación sin mensajes, conversación con solo los mensajes no leídos del contacto, conversación con un mensaje por identificador externo, y conversación activa por contacto. La búsqueda de la conversación activa SHALL incluir los identificadores externos de sus mensajes para permitir la detección de duplicados.

#### Scenario: Carga sin mensajes

- **WHEN** una operación solo necesita los metadatos de la conversación
- **THEN** el repositorio recupera la conversación sin cargar sus mensajes

#### Scenario: Carga de no leídos

- **WHEN** el agente lee una conversación
- **THEN** el repositorio recupera únicamente los mensajes del contacto sin leer

#### Scenario: Búsqueda por identificador externo

- **WHEN** un caso de uso opera sobre un mensaje por su identificador externo
- **THEN** el repositorio recupera la conversación y ese mensaje

#### Scenario: Carga de identificadores externos de la conversación activa

- **WHEN** un caso de uso necesita detectar mensajes duplicados en la conversación activa
- **THEN** el repositorio recupera la conversación activa junto con los identificadores externos de sus mensajes

## ADDED Requirements

### Requirement: Seguimiento de mensajes modificados

El sistema SHALL rastrear los mensajes modificados desde que la conversación fue cargada. La creación SHALL marcar todos sus mensajes iniciales como modificados. La rehidratación SHALL NOT marcar ninguno. Cada mutación SHALL marcar el mensaje afectado. El sistema SHALL exponer los mensajes modificados en orden determinista por `sentAt` ascendente y, en empate, por `id`.

#### Scenario: Creación marca los mensajes iniciales

- **WHEN** se crea una conversación con mensajes
- **THEN** todos sus mensajes quedan marcados como modificados

#### Scenario: Rehidratación no marca mensajes

- **WHEN** se rehidrata una conversación con mensajes
- **THEN** ningún mensaje queda marcado como modificado

#### Scenario: Mutación marca el mensaje afectado

- **WHEN** una operación modifica un mensaje
- **THEN** solo ese mensaje queda marcado como modificado

#### Scenario: Orden determinista de los modificados

- **WHEN** se exponen los mensajes modificados
- **THEN** vienen ordenados por `sentAt` ascendente y, en empate, por `id`
