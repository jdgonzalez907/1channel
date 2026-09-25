# Spec Delta

## Purpose

Expone la API pública del módulo de contactos que otros módulos consumen para resolver o crear un contacto por su identificador externo y para obtener el identificador externo de un contacto.

## ADDED Requirements

### Requirement: Resolver o crear contacto por identificador externo

El módulo de contactos SHALL exponer en su raíz una interfaz `ContactsAPI` con una operación que, dado un identificador externo de contacto, retorne el contacto existente o cree uno nuevo.

#### Scenario: Contacto existente

- **WHEN** se solicita el contacto por un identificador externo ya registrado
- **THEN** la API retorna el contacto existente sin crear uno nuevo

#### Scenario: Contacto nuevo

- **WHEN** se solicita el contacto por un identificador externo no registrado
- **THEN** la API crea el contacto y retorna su identificador

#### Scenario: Identificador externo vacío

- **WHEN** se solicita el contacto con un identificador externo vacío
- **THEN** la API retorna error de identificador externo inválido

### Requirement: Obtener identificador externo de un contacto

El módulo de contactos SHALL permitir obtener el identificador externo de un contacto a partir de su identificador interno.

#### Scenario: Contacto existente

- **WHEN** se solicita el identificador externo de un contacto registrado
- **THEN** la API retorna el identificador externo asociado

#### Scenario: Contacto inexistente

- **WHEN** se solicita el identificador externo de un contacto no registrado
- **THEN** la API retorna error de contacto no encontrado
