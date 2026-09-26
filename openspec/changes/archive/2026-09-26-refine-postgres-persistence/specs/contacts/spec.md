# Spec Delta

## ADDED Requirements

### Requirement: Rehidratar contacto

El módulo de contactos SHALL permitir reconstruir un contacto a partir de datos persistidos sin validar sus invariantes, aceptando un identificador externo vacío. La rehidratación SHALL ser distinta de la creación, que sí valida el identificador externo.

#### Scenario: Rehidratar contacto con identificador externo

- **WHEN** se rehidrata un contacto con identificador, identificador externo y fecha de creación
- **THEN** se obtiene el contacto con esos valores

#### Scenario: Rehidratar contacto con identificador externo vacío

- **WHEN** se rehidrata un contacto con identificador externo vacío
- **THEN** se obtiene el contacto sin validar el identificador externo
