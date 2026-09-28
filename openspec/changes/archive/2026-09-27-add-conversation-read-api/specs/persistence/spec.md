# Spec Delta

## ADDED Requirements

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

## MODIFIED Requirements

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
