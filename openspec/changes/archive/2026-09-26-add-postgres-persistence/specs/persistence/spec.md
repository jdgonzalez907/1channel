# Spec Delta

## ADDED Requirements

### Requirement: Persistencia de conversaciones

El repositorio de conversaciones SHALL persistir la conversación y únicamente los mensajes marcados como modificados, dentro de una transacción, mediante inserción con actualización en conflicto por identificador. Las lecturas SHALL indicar ausencia cuando la fila no exista. La búsqueda de la conversación activa SHALL incluir los identificadores externos de sus mensajes. Los conflictos de unicidad de la base SHALL propagarse como error.

#### Scenario: Persistir conversación nueva

- **WHEN** se guarda una conversación recién creada
- **THEN** la conversación y sus mensajes iniciales quedan persistidos en una sola transacción

#### Scenario: Persistir solo lo modificado

- **WHEN** se guarda una conversación con mensajes cargados sin cambios
- **THEN** solo la conversación y los mensajes modificados se escriben

#### Scenario: Guardar sin mensajes modificados

- **WHEN** se guarda una conversación sin mensajes modificados
- **THEN** la conversación se escribe y no se escribe ningún mensaje

#### Scenario: Lectura inexistente

- **WHEN** se busca una conversación que no existe
- **THEN** el repositorio devuelve ausencia sin error

#### Scenario: Conflicto de unicidad

- **WHEN** una escritura viola una restricción de unicidad
- **THEN** el repositorio propaga el error de la base

### Requirement: Persistencia de contactos y agentes

El repositorio de contactos SHALL permitir buscar por identificador y por identificador externo, y guardar mediante inserción con actualización en conflicto por identificador. El repositorio de agentes SHALL permitir buscar por identificador. Las lecturas SHALL indicar ausencia cuando la fila no exista.

#### Scenario: Buscar contacto inexistente

- **WHEN** se busca un contacto que no existe
- **THEN** el repositorio devuelve ausencia sin error

#### Scenario: Guardar contacto

- **WHEN** se guarda un contacto
- **THEN** sus campos quedan persistidos

#### Scenario: Buscar agente inexistente

- **WHEN** se busca un agente que no existe
- **THEN** el repositorio devuelve ausencia sin error
