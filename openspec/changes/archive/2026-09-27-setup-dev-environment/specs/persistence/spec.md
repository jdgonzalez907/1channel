# Spec Delta

## ADDED Requirements

### Requirement: Timestamps como instantes UTC

El sistema SHALL almacenar y recuperar los timestamps de todas las entidades como instantes UTC, con independencia de la zona horaria del host donde corre la aplicación y de la configuración regional del proceso. Al recuperar una fila, el sistema SHALL exponer sus timestamps en UTC.

#### Scenario: Lectura en UTC desde un host en otra zona

- **WHEN** el proceso corre con una zona horaria distinta de UTC
- **AND** se recupera de la base una entidad con timestamps
- **THEN** los timestamps expuestos quedan expresados en UTC

#### Scenario: Escritura que preserva el instante

- **WHEN** se persiste una entidad con un timestamp
- **THEN** la base conserva ese instante
- **AND** al releerlo se expone en UTC

#### Scenario: Fecha nula permanece nula

- **WHEN** se recupera una entidad cuyo timestamp opcional no tiene valor
- **THEN** el sistema no asigna un instante por defecto y lo expone como ausencia
