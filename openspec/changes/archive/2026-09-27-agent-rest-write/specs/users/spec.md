# Spec Delta

## Purpose

Identidad de los usuarios del sistema, compartida por los módulos que necesitan validar la existencia, crear o reconstruir un usuario.

## ADDED Requirements

### Requirement: API pública del módulo de usuarios

El módulo de usuarios SHALL exponer en su raíz una interfaz `UsersAPI` con los métodos que necesitan otros módulos, devolviendo valores primitivos para minimizar el acoplamiento. Para el MVP se requiere la validación de existencia por identificador.

#### Scenario: Usuario existente

- **WHEN** un módulo solicita un usuario por su identificador y el usuario existe
- **THEN** la API retorna el identificador del usuario sin error

#### Scenario: Usuario inexistente

- **WHEN** un módulo solicita un usuario por su identificador y el usuario no existe
- **THEN** la API retorna error de usuario no encontrado

### Requirement: Crear usuario del sistema

El módulo de usuarios SHALL permitir crear un usuario del sistema con un identificador y una fecha de creación. La creación SHALL validar que el identificador no sea nulo y SHALL ser distinta de la rehidratación.

#### Scenario: Crear usuario válido

- **WHEN** se crea un usuario con identificador no nulo y fecha de creación
- **THEN** se obtiene el usuario con esos valores

#### Scenario: Fallar al crear usuario con identificador nulo

- **WHEN** se intenta crear un usuario con identificador nulo
- **THEN** el sistema retorna error de identificador inválido

### Requirement: Rehidratar usuario

El módulo de usuarios SHALL permitir reconstruir un usuario a partir de datos persistidos sin validar sus invariantes. La rehidratación SHALL ser distinta de la creación.

#### Scenario: Rehidratar usuario

- **WHEN** se rehidrata un usuario con identificador y fecha de creación
- **THEN** se obtiene el usuario con esos valores
