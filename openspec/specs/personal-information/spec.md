# personal-information Specification

## Purpose

Modela a la persona detrás de los contactos: un registro de datos personales identificado por un documento único, compartible por varios contactos de distintos canales y con ciclo de vida propio.

## Requirements

### Requirement: Persona con documento único e inmutable

El sistema SHALL modelar una persona con `identification_number` no vacío, compuesto únicamente por letras y números, único en todo el sistema e inmutable una vez creada, con los datos opcionales `first_name`, `last_name`, `phone_number`, `email` y `address`, y con `created_at` y `updated_at`. La persona SHALL poder existir sin ningún contacto asociado.

#### Scenario: Documento duplicado

- **WHEN** se intenta crear una persona cuyo `identification_number` ya existe
- **THEN** el sistema no crea una segunda persona con ese documento

#### Scenario: Documento inmutable

- **WHEN** se modifican los datos de una persona existente
- **THEN** su `identification_number` permanece igual

#### Scenario: Persona huérfana

- **WHEN** una persona no tiene ningún contacto asociado
- **THEN** la persona existe y es consultable por su documento

#### Scenario: Documento ausente o inválido

- **WHEN** se intenta guardar una persona con `identification_number` vacío o con un carácter distinto de letras y números
- **THEN** el sistema rechaza la operación por validación

#### Scenario: Solo con documento

- **WHEN** se guarda una persona con `identification_number` y todos los datos opcionales ausentes
- **THEN** el sistema acepta la operación

### Requirement: Longitud de los datos de la persona

`identification_number` SHALL tener entre 1 y 100 grafemas. Los datos opcionales, cuando estén presentes y no vacíos, SHALL tener como máximo 100 grafemas (`first_name`, `last_name`, `phone_number`) o 254 grafemas (`email`, `address`). Un dato ausente o vacío SHALL tratarse como ausente. El límite SHALL aplicarse por grafemas.

#### Scenario: Documento vacío

- **WHEN** se intenta guardar una persona con `identification_number` sin grafemas
- **THEN** el sistema rechaza la operación por validación

#### Scenario: Documento fuera de rango

- **WHEN** se guarda una persona cuyo `identification_number` supera 100 grafemas
- **THEN** el sistema rechaza la operación por validación

#### Scenario: Dato opcional dentro del límite

- **WHEN** se guarda una persona con uno o más datos opcionales dentro de su máximo
- **THEN** el sistema acepta la operación

#### Scenario: Dato opcional que excede el máximo

- **WHEN** se guarda una persona con un dato opcional que supera su máximo
- **THEN** el sistema rechaza la operación por validación

#### Scenario: Conteo por grafemas

- **WHEN** un campo contiene acentos o emojis
- **THEN** cada grafema cuenta como una unidad de longitud, no cada byte

### Requirement: Guardar persona como upsert por documento

Guardar una persona SHALL insertar una fila nueva cuando el documento no existe, o actualizar los datos de la fila existente cuando el documento ya existe, reemplazando los datos previos por los provistos (un dato ausente o vacío queda nulo), sin crear una fila duplicada. Al actualizar, `updated_at` SHALL reflejar el nuevo guardado y `created_at` SHALL conservar el instante original. La operación SHALL devolver el identificador de la fila persistida, que es el de la fila creada primero cuando hubo conflicto.

#### Scenario: Documento nuevo

- **WHEN** se guarda una persona con un documento que no existe
- **THEN** se crea la fila y se devuelve su identificador

#### Scenario: Documento existente

- **WHEN** se guarda una persona con un documento que ya existe
- **THEN** los datos de la fila existente se reemplazan con los nuevos valores
- **AND** se devuelve el identificador de esa fila existente
- **AND** no se crea una segunda fila

#### Scenario: Dos guardados concurrentes del mismo documento

- **WHEN** dos guardados del mismo documento llegan a la vez con identificadores internos distintos
- **THEN** queda una sola fila
- **AND** el identificador conservado es el de la fila creada primero
- **AND** los datos de la fila quedan con los del último guardado

### Requirement: Asociación de la persona con los contactos

Un contacto SHALL poder quedar asociado a lo sumo a una persona. La asociación SHALL fijar la referencia del contacto hacia la persona, y una misma persona SHALL poder estar referenciada por varios contactos. El sistema SHALL NOT exponer operaciones de desasociar ni de borrar personas.

#### Scenario: Asociar un contacto a una persona

- **WHEN** se asocia un contacto a una persona
- **THEN** el contacto queda referenciando a esa persona

#### Scenario: Varios contactos a una misma persona

- **WHEN** dos contactos distintos se asocian al mismo `identification_number`
- **THEN** ambos quedan referenciando a la misma persona

#### Scenario: Reasociar un contacto

- **WHEN** se asocia a un contacto una persona distinta de la que tenía
- **THEN** la referencia del contacto apunta a la nueva persona

#### Scenario: Sin desasociar ni borrar

- **WHEN** un operador busca separar un contacto de su persona o eliminar una persona
- **THEN** el sistema no ofrece esa operación
