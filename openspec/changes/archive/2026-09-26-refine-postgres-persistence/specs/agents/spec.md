# Spec Delta

## ADDED Requirements

### Requirement: Rehidratar agente

El módulo de agentes SHALL permitir reconstruir un agente a partir de datos persistidos sin validar sus invariantes. La rehidratación SHALL ser distinta de la creación.

#### Scenario: Rehidratar agente

- **WHEN** se rehidrata un agente con identificador y fecha de creación
- **THEN** se obtiene el agente con esos valores
