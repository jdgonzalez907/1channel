# Spec Delta

## Purpose

Expone la API pública del módulo de agentes que otros módulos consumen para validar la existencia de un agente sin acoplarse a su implementación interna.

## ADDED Requirements

### Requirement: API pública del módulo de agentes

El módulo de agentes SHALL exponer en su raíz una interfaz `AgentsAPI` con los métodos que necesitan otros módulos, devolviendo valores primitivos para minimizar el acoplamiento. Para el MVP solo se requiere la validación de existencia por identificador.

#### Scenario: Agente existente

- **WHEN** un módulo solicita un agente por su identificador y el agente existe
- **THEN** la API retorna el identificador del agente sin error

#### Scenario: Agente inexistente

- **WHEN** un módulo solicita un agente por su identificador y el agente no existe
- **THEN** la API retorna error de agente no encontrado
