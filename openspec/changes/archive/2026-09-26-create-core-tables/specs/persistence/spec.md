# Spec Delta

## Purpose

Persiste agentes, contactos, conversaciones y mensajes en PostgreSQL con las restricciones que las búsquedas del MVP ya asumen.

## ADDED Requirements

### Requirement: Tablas del agregado de conversación

El sistema SHALL persistir agentes, contactos, conversaciones y mensajes en tablas separadas. Una conversación SHALL referenciar como máximo un agente y como máximo un contacto. Un mensaje SHALL pertenecer a exactamente una conversación. El sistema SHALL NOT generar identificadores: los recibe de la aplicación y no asigna un valor por defecto en la base.

#### Scenario: Persistir un agente

- **WHEN** se guarda un agente con identificador y fecha de creación
- **THEN** la fila queda con ese identificador y esa fecha, sin otros atributos

#### Scenario: Persistir un contacto

- **WHEN** se guarda un contacto con identificador, identificador externo no vacío y fecha de creación
- **THEN** la fila queda con esos tres valores

#### Scenario: Persistir una conversación

- **WHEN** se guarda una conversación con identificador, estado, fechas y agente o contacto opcionales
- **THEN** la fila conserva el estado, `created_at`, `updated_at` nulo si no hubo actividad y `finished_at` nulo si no está finalizada

#### Scenario: Persistir un mensaje

- **WHEN** se guarda un mensaje de una conversación con estado, tipo, texto opcional, dueño, identificador externo opcional y timestamps de envío, lectura, edición y borrado
- **THEN** la fila queda asociada a esa conversación con esos valores, permitiendo nulos en texto, identificador externo y timestamps de lectura, edición y borrado

### Requirement: Integridad referencial restrictiva

El sistema SHALL rechazar una conversación o un mensaje cuyo agente, contacto o conversación referenciados no existan. El sistema SHALL rechazar borrar un agente, un contacto o una conversación mientras existan filas que los referencien. El sistema SHALL rechazar actualizar el identificador de un agente, un contacto o una conversación mientras existan filas que lo referencien. El sistema SHALL NOT anular ni borrar en cascada esas referencias.

#### Scenario: Conversación con agente inexistente

- **WHEN** se intenta guardar una conversación con un identificador de agente que no existe
- **THEN** la base rechaza la escritura

#### Scenario: Mensaje con conversación inexistente

- **WHEN** se intenta guardar un mensaje cuya conversación no existe
- **THEN** la base rechaza la escritura

#### Scenario: Borrar un contacto referenciado

- **WHEN** se intenta borrar un contacto que tiene una conversación o un mensaje
- **THEN** la base rechaza el borrado y las filas dependientes permanecen

#### Scenario: Actualizar un identificador referenciado

- **WHEN** se intenta actualizar el identificador de un agente, un contacto o una conversación que tiene filas asociadas
- **THEN** la base rechaza la actualización y las filas dependientes permanecen

### Requirement: Invariantes de conversación y mensaje

El sistema SHALL exigir que una conversación tenga agente o contacto. El sistema SHALL exigir `finished_at` cuando el estado es `expired` o `resolved`. El sistema SHALL exigir que un mensaje tenga exactamente un dueño, agente o contacto. El sistema SHALL aceptar solo los estados de conversación `pending`, `assigned`, `expired` y `resolved`, y solo los estados de mensaje `sent`, `read`, `deleted` y `failed`. El sistema SHALL aceptar solo el tipo de mensaje `text` en este esquema. El sistema SHALL NOT limitar la longitud del texto en la base: ese límite lo aplica el dominio por grafemas.

#### Scenario: Conversación sin agente ni contacto

- **WHEN** se intenta guardar una conversación con agente y contacto nulos
- **THEN** la base rechaza la escritura

#### Scenario: Conversación finalizada sin fecha

- **WHEN** se intenta guardar una conversación `expired` o `resolved` sin `finished_at`
- **THEN** la base rechaza la escritura

#### Scenario: Mensaje con dos dueños

- **WHEN** se intenta guardar un mensaje con agente y contacto a la vez, o sin ninguno
- **THEN** la base rechaza la escritura

#### Scenario: Estado o tipo desconocido

- **WHEN** se intenta guardar un estado o un tipo de mensaje fuera de los valores admitidos
- **THEN** la base rechaza la escritura

### Requirement: Una conversación abierta por contacto

El sistema SHALL impedir que un mismo contacto tenga más de una conversación en estado `pending` o `assigned`. El sistema SHALL permitir cero conversaciones abiertas y SHALL permitir varias conversaciones finalizadas del mismo contacto. Una conversación sin contacto SHALL NOT ocupar ese cupo.

#### Scenario: Segunda conversación abierta

- **WHEN** un contacto ya tiene una conversación `pending` o `assigned` y se intenta guardar otra en uno de esos estados
- **THEN** la base rechaza la escritura

#### Scenario: Nueva conversación tras finalizar

- **WHEN** la única conversación abierta de un contacto pasa a `expired` o `resolved` y llega otra en estado `pending`
- **THEN** la base acepta la nueva fila

### Requirement: Identificadores externos únicos

El sistema SHALL impedir dos contactos con el mismo identificador externo. El sistema SHALL impedir dos mensajes con el mismo identificador externo no nulo. El sistema SHALL permitir varios mensajes sin identificador externo.

#### Scenario: Contacto duplicado por identificador externo

- **WHEN** se intenta guardar un contacto cuyo identificador externo ya existe
- **THEN** la base rechaza la escritura

#### Scenario: Mensaje duplicado por identificador externo

- **WHEN** se intenta guardar un mensaje con un identificador externo no nulo que ya existe en cualquier conversación
- **THEN** la base rechaza la escritura

#### Scenario: Mensajes sin identificador externo

- **WHEN** se guardan varios mensajes con identificador externo nulo
- **THEN** la base acepta todas las filas

### Requirement: Acceso por las búsquedas del MVP

El sistema SHALL poder localizar un contacto por su identificador externo, una conversación abierta por contacto, los mensajes de una conversación y un mensaje por su identificador externo, sin barrer la tabla completa para esas claves. El sistema SHALL NOT crear índices para listar por agente, por `updated_at`, por `sent_at` ni un índice parcial de mensajes no leídos.

#### Scenario: Búsqueda de conversación abierta

- **WHEN** se busca la conversación `pending` o `assigned` de un contacto
- **THEN** la búsqueda usa el índice de una conversación abierta por contacto

#### Scenario: Mensajes de una conversación

- **WHEN** se cargan los mensajes de una conversación, incluidos los no leídos del contacto
- **THEN** la búsqueda usa el índice por conversación y filtra la lectura en la consulta

#### Scenario: Mensaje por identificador externo

- **WHEN** se busca un mensaje por su identificador externo
- **THEN** la búsqueda usa el índice único de ese identificador
