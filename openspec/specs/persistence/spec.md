# persistence Specification

## Purpose
Persiste agentes, contactos, conversaciones y mensajes en PostgreSQL con las restricciones que las búsquedas del MVP ya asumen.

## Requirements

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

### Requirement: Invariantes de conversación

El sistema SHALL exigir que una conversación tenga agente o contacto. El sistema SHALL exigir `finished_at` cuando el estado es `expired` o `resolved`. El sistema SHALL aceptar solo los estados de conversación `pending`, `assigned`, `expired` y `resolved`.

#### Scenario: Conversación sin agente ni contacto

- **WHEN** se intenta guardar una conversación con agente y contacto nulos
- **THEN** la base rechaza la escritura

#### Scenario: Conversación finalizada sin fecha

- **WHEN** se intenta guardar una conversación `expired` o `resolved` sin `finished_at`
- **THEN** la base rechaza la escritura

#### Scenario: Estado de conversación desconocido

- **WHEN** se intenta guardar una conversación con un estado fuera de `pending`, `assigned`, `expired` y `resolved`
- **THEN** la base rechaza la escritura

### Requirement: Invariantes de mensaje

El sistema SHALL exigir que un mensaje tenga exactamente un dueño, agente o contacto. El sistema SHALL aceptar solo los estados de mensaje `sent`, `read`, `deleted` y `failed`. El sistema SHALL aceptar solo el tipo de mensaje `text` en este esquema. El sistema SHALL NOT limitar la longitud del texto en la base: ese límite lo aplica el dominio por grafemas.

#### Scenario: Mensaje sin un único dueño

- **WHEN** se intenta guardar un mensaje con agente y contacto a la vez, o sin ninguno
- **THEN** la base rechaza la escritura

#### Scenario: Estado o tipo de mensaje desconocido

- **WHEN** se intenta guardar un mensaje con un estado fuera de `sent`, `read`, `deleted` y `failed`, o con un tipo distinto de `text`
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

El sistema SHALL poder localizar un contacto por su identificador externo, una conversación abierta por contacto, los mensajes de una conversación y un mensaje por su identificador externo, sin barrer la tabla completa para esas claves. El sistema SHALL disponer además de índices sobre `conversations` por `(user_id, last_message_at, id)` y por `(status, last_message_at, id)` para listar la bandeja ordenada por el último mensaje, y por `contact_id` para el filtro por contacto, y de un índice sobre `messages` por `(conversation_id, sent_at, id)` para ordenar y paginar los mensajes de una conversación. El sistema SHALL NOT crear un índice parcial de mensajes no leídos: el conteo de no leídos vive en `unread_count`.

#### Scenario: Búsqueda de conversación abierta

- **WHEN** se busca la conversación `pending` o `assigned` de un contacto
- **THEN** la búsqueda usa el índice de una conversación abierta por contacto

#### Scenario: Mensajes de una conversación

- **WHEN** se cargan y paginan los mensajes de una conversación
- **THEN** la búsqueda usa el índice por conversación y `sent_at`, sin barrer la tabla completa

#### Scenario: Mensaje por identificador externo

- **WHEN** se busca un mensaje por su identificador externo
- **THEN** la búsqueda usa el índice único de ese identificador

#### Scenario: Filtrar conversaciones de la bandeja

- **WHEN** se listan las conversaciones visibles para un agente con su filtro de estado
- **THEN** la búsqueda usa los índices por agente y por estado junto con `last_message_at`, sin barrer la tabla completa

#### Scenario: Filtrar conversaciones por contacto

- **WHEN** se listan las conversaciones de un contacto
- **THEN** la búsqueda usa el índice por contacto sin barrer la tabla completa

#### Scenario: Ordenar por último mensaje

- **WHEN** se ordenan las conversaciones por el `sent_at` de su último mensaje
- **THEN** la búsqueda usa el índice por agente o por estado junto con `last_message_at` de la conversación

### Requirement: Persistencia de conversaciones

El repositorio de conversaciones SHALL persistir la conversación y únicamente los mensajes marcados como modificados, dentro de una transacción, mediante inserción con actualización en conflicto por identificador. Las lecturas SHALL indicar ausencia cuando la fila no exista. La búsqueda de la conversación activa SHALL incluir los identificadores externos de sus mensajes. Los conflictos de unicidad de la base SHALL propagarse como error.

#### Scenario: Persistir conversación nueva

- **WHEN** se guarda una conversación recién creada
- **THEN** la conversación y sus mensajes iniciales quedan persistidos en una sola transacción

#### Scenario: Persistir solo lo modificado

- **WHEN** se guarda una conversación con mensajes cargados sin cambios
- **THEN** solo la conversación y los mensajes modificados se escriben

#### Scenario: Guardar sin mensajes modificados

- **WHEN** se guarda una conversación sin mensajes modificados
- **THEN** la conversación se escribe y no se escribe ningún mensaje

#### Scenario: Lectura inexistente

- **WHEN** se busca una conversación que no existe
- **THEN** el repositorio devuelve ausencia sin error

#### Scenario: Conflicto de unicidad

- **WHEN** una escritura viola una restricción de unicidad
- **THEN** el repositorio propaga el error de la base

### Requirement: Persistencia de contactos y agentes

El repositorio de contactos SHALL permitir buscar por identificador y por identificador externo, y guardar mediante inserción con actualización en conflicto por identificador. El repositorio de agentes SHALL permitir buscar por identificador. Las lecturas SHALL indicar ausencia cuando la fila no exista.

#### Scenario: Buscar contacto inexistente

- **WHEN** se busca un contacto que no existe
- **THEN** el repositorio devuelve ausencia sin error

#### Scenario: Guardar contacto

- **WHEN** se guarda un contacto
- **THEN** sus campos quedan persistidos

#### Scenario: Buscar agente inexistente

- **WHEN** se busca un agente que no existe
- **THEN** el repositorio devuelve ausencia sin error

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

### Requirement: Modelo de lectura denormalizado de conversación

Cada conversación SHALL almacenar `last_message_at`, `last_message_id` y `unread_count`, mantenidos en la misma transacción que persiste la conversación y sus mensajes modificados. `last_message_id` SHALL referenciar el mensaje de mayor `(sent_at, id)` de la conversación y `last_message_at` SHALL ser el `sent_at` de ese mensaje; ambos SHALL quedar establecidos en toda conversación persistida, dado que toda conversación tiene al menos un mensaje. `unread_count` SHALL ser la cantidad de mensajes del contacto de la conversación cuyo `read_at` es nulo, incluidos los mensajes eliminados; el conteo SHALL NOT filtrar por estado. Los valores SHALL recomputarse desde los mensajes persistidos después de escribirlos, de modo que sean correctos sin importar cuánto de la conversación se haya cargado en memoria, y SHALL recomputarse únicamente cuando la transacción modificó al menos un mensaje. El sistema SHALL NOT declarar una clave foránea desde `conversations.last_message_id` hacia `messages`.

#### Scenario: Conversación recién creada

- **WHEN** se persiste una conversación nueva con sus mensajes
- **THEN** `last_message_id` y `last_message_at` corresponden al mensaje más reciente
- **AND** `unread_count` refleja los mensajes del contacto sin leer

#### Scenario: Mensaje entrante

- **WHEN** se persiste un mensaje nuevo del contacto
- **THEN** `last_message_id` y `last_message_at` apuntan a ese mensaje
- **AND** `unread_count` aumenta en uno

#### Scenario: Marcar como leídos

- **WHEN** se marcan como leídos los mensajes del contacto
- **THEN** `unread_count` pasa a reflejar los mensajes del contacto aún sin leer

#### Scenario: Carga parcial

- **WHEN** la conversación se guarda habiendo cargado solo una parte de sus mensajes
- **THEN** los valores se recomputan desde los mensajes persistidos y no desde lo cargado en memoria

#### Scenario: Último mensaje eliminado

- **WHEN** el mensaje más reciente está marcado como `deleted`
- **THEN** `last_message_id` y `last_message_at` siguen apuntando a él

#### Scenario: Mensaje eliminado sin leer

- **WHEN** un mensaje del contacto marcado como `deleted` todavía tiene `read_at` nulo
- **THEN** `unread_count` lo cuenta igual

#### Scenario: Save sin mensajes modificados

- **WHEN** se guarda una conversación sin mensajes modificados
- **THEN** `last_message_id`, `last_message_at` y `unread_count` no se recomputan y conservan su valor
