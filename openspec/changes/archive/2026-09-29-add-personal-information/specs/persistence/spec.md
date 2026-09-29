# Spec Delta

## ADDED Requirements

### Requirement: Tabla y persistencia de la persona

El sistema SHALL persistir la persona en una tabla `personal_information` con identificador propio provisto por la aplicación, sin valor por defecto en la base, `identification_number` no vacío y único, los campos opcionales `first_name`, `last_name`, `phone_number`, `email` y `address` que admiten nulo, y `created_at`/`updated_at`. El sistema SHALL NOT limitar la longitud de esos campos en la base ni validar su juego de caracteres en la base: la longitud por grafemas y el alfabeto de letras y números los aplica el dominio. La persona SHALL poder existir sin filas que la referencien.

#### Scenario: Persistir una persona

- **WHEN** se guarda una persona con identificador, documento y datos
- **THEN** la fila queda con esos valores
- **AND** los datos opcionales ausentes quedan nulos

#### Scenario: Documento duplicado

- **WHEN** se intenta guardar una segunda persona con un `identification_number` ya existente
- **THEN** la base rechaza la escritura

### Requirement: Persistencia de personas

El repositorio de personas SHALL permitir buscar por identificador y por documento, y guardar mediante inserción con actualización en conflicto por documento, asignando a la entidad el identificador de la fila persistida. Las lecturas SHALL indicar ausencia cuando la fila no exista. Los conflictos de unicidad distintos del documento SHALL propagarse como error.

#### Scenario: Buscar persona inexistente

- **WHEN** se busca por identificador o por documento una persona que no existe
- **THEN** el repositorio devuelve ausencia sin error

#### Scenario: Guardar persona nueva

- **WHEN** se guarda una persona con un documento que no existe
- **THEN** la fila se inserta con el identificador provisto

#### Scenario: Guardar persona existente

- **WHEN** se guarda una persona con un documento que ya existe
- **THEN** se actualizan sus datos y su `updated_at`, y se conservan el identificador y el `created_at` de la fila existente

### Requirement: Referencia restrictiva del contacto a la persona

El contacto SHALL tener `display_name` opcional y una referencia nullable `personal_information_id` hacia la persona. La base SHALL rechazar guardar un contacto que referencie una persona inexistente, y SHALL rechazar borrar o actualizar el identificador de una persona mientras existan contactos que la referencien, sin anular ni borrar en cascada.

#### Scenario: Contacto con persona inexistente

- **WHEN** se intenta guardar un contacto cuyo `personal_information_id` no existe
- **THEN** la base rechaza la escritura

#### Scenario: Borrar una persona referenciada

- **WHEN** se intenta borrar una persona que tiene contactos asociados
- **THEN** la base rechaza el borrado y los contactos permanecen

### Requirement: Acceso a la persona por su documento

El sistema SHALL poder localizar una persona por su `identification_number` sin barrer la tabla completa, mediante una restricción de unicidad sobre ese documento.

#### Scenario: Búsqueda por documento

- **WHEN** se busca una persona por su `identification_number`
- **THEN** la búsqueda usa la restricción de unicidad de ese documento
