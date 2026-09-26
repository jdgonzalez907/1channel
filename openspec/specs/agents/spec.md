# agents Specification

## Purpose
Expone la API pública del módulo de agentes que otros módulos consumen para validar la existencia de un agente sin acoplarse a su implementación interna.

## Requirements

### Requirement: API pública del módulo de agentes

El módulo de agentes SHALL exponer en su raíz una interfaz `AgentsAPI` con los métodos que necesitan otros módulos, devolviendo valores primitivos para minimizar el acoplamiento. Para el MVP solo se requiere la validación de existencia por identificador.

#### Scenario: Agente existente

- **WHEN** un módulo solicita un agente por su identificador y el agente existe
- **THEN** la API retorna el identificador del agente sin error

#### Scenario: Agente inexistente

- **WHEN** un módulo solicita un agente por su identificador y el agente no existe
- **THEN** la API retorna error de agente no encontrado

### Requirement: Rehidratar agente

El módulo de agentes SHALL permitir reconstruir un agente a partir de datos persistidos sin validar sus invariantes. La rehidratación SHALL ser distinta de la creación.

#### Scenario: Rehidratar agente

- **WHEN** se rehidrata un agente con identificador y fecha de creación
- **THEN** se obtiene el agente con esos valores
