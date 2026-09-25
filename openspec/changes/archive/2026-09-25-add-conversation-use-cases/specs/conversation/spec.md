# Spec Delta

## ADDED Requirements

### Requirement: Caso de uso: recibir mensaje de contacto

El sistema SHALL adjuntar un mensaje entrante del contacto a la conversación activa del contacto; si no existe una conversación activa, SHALL iniciar una conversación nueva que contenga ese mensaje. El contacto SHALL resolverse o crearse por su identificador externo.

#### Scenario: Mensaje a conversación activa

- **WHEN** llega un mensaje de un contacto que tiene una conversación activa
- **THEN** el mensaje se adjunta a esa conversación y se actualiza su timestamp de actualización

#### Scenario: Contacto sin conversación activa

- **WHEN** llega un mensaje de un contacto que no tiene conversación activa
- **THEN** el sistema inicia una conversación nueva en estado pending con ese mensaje

#### Scenario: Mensaje duplicado es idempotente

- **WHEN** llega un mensaje cuyo identificador externo ya existe en la conversación
- **THEN** el sistema no altera la conversación y no retorna error

#### Scenario: Contacto nuevo

- **WHEN** llega un mensaje de un identificador externo sin contacto registrado
- **THEN** el sistema crea el contacto antes de adjuntar el mensaje

#### Scenario: Conversación de otro contacto

- **WHEN** el mensaje llega para una conversación cuyo contacto no coincide
- **THEN** el sistema retorna ErrConversationContactNotOwner

### Requirement: Caso de uso: editar mensaje de contacto

El sistema SHALL aplicar la edición de un mensaje enviada por el contacto, localizándolo por su identificador externo dentro de la conversación. Las ediciones con timestamp anterior o igual a la última edición aplicada SHALL descartarse sin error. El sistema no SHALL rechazar la edición por el estado finalizado de la conversación.

#### Scenario: Edición aplicada

- **WHEN** el contacto edita un mensaje suyo y el timestamp es más reciente que la última edición
- **THEN** el texto y el timestamp de edición del mensaje se actualizan

#### Scenario: Edición obsoleta

- **WHEN** el contacto edita un mensaje con timestamp anterior o igual a la última edición aplicada
- **THEN** el mensaje conserva su edición más reciente y no se retorna error

#### Scenario: Identificador externo inexistente

- **WHEN** el contacto edita un mensaje con un identificador externo que no pertenece a la conversación
- **THEN** el sistema retorna ErrMessageNotFound

#### Scenario: Mensaje no pertenece al contacto

- **WHEN** el contacto intenta editar un mensaje enviado por el agente
- **THEN** el sistema retorna ErrConversationContactNotOwner

#### Scenario: Edición en conversación finalizada

- **WHEN** el contacto edita un mensaje en una conversación finalizada
- **THEN** la edición se aplica igualmente

### Requirement: Caso de uso: eliminar mensaje de contacto

El sistema SHALL aplicar la eliminación (soft delete) de un mensaje enviada por el contacto, localizándolo por su identificador externo. Un mensaje ya eliminado SHALL tratarse como idempotente (sin error). El sistema no SHALL rechazar la eliminación por el estado finalizado de la conversación.

#### Scenario: Eliminación aplicada

- **WHEN** el contacto elimina un mensaje suyo no eliminado
- **THEN** el mensaje queda con timestamp de eliminación y estado deleted

#### Scenario: Eliminación repetida es idempotente

- **WHEN** el contacto elimina un mensaje ya eliminado
- **THEN** el sistema no altera el mensaje y no retorna error

#### Scenario: Identificador externo inexistente

- **WHEN** el contacto elimina un mensaje con un identificador externo que no pertenece a la conversación
- **THEN** el sistema retorna ErrMessageNotFound

#### Scenario: Mensaje no pertenece al contacto

- **WHEN** el contacto intenta eliminar un mensaje enviado por el agente
- **THEN** el sistema retorna ErrConversationContactNotOwner

#### Scenario: Eliminación en conversación finalizada

- **WHEN** el contacto elimina un mensaje en una conversación finalizada
- **THEN** la eliminación se aplica igualmente

### Requirement: Caso de uso: enviar mensaje de agente

El sistema SHALL registrar y enviar un mensaje de un agente a una conversación existente. Si la conversación está pendiente sin agente, SHALL asignar al agente y cambiar el estado a assigned. El sistema SHALL validar la existencia del agente, persistir el mensaje y enviarlo por el canal externo. Si el envío externo falla, SHALL marcar el mensaje como failed y retornar error.

#### Scenario: Envío exitoso

- **WHEN** un agente envía un mensaje a una conversación asignada a él y el canal externo responde con un identificador
- **THEN** el mensaje queda registrado y con su identificador externo asignado

#### Scenario: Reclamar conversación pendiente

- **WHEN** un agente envía un mensaje a una conversación pendiente sin agente
- **THEN** el agente queda asignado, el estado cambia a assigned y el mensaje se registra

#### Scenario: Agente no asignado

- **WHEN** un agente intenta enviar un mensaje a una conversación asignada a otro agente
- **THEN** el sistema retorna ErrConversationAgentNotOwner

#### Scenario: Envío externo fallido

- **WHEN** el canal externo no puede entregar el mensaje
- **THEN** el mensaje queda en estado failed y el sistema retorna el error del envío

#### Scenario: Agente inexistente

- **WHEN** un identificador de agente no registrado intenta enviar un mensaje
- **THEN** el sistema retorna error de agente no encontrado

#### Scenario: Conversación no acepta mensajes

- **WHEN** un agente envía un mensaje a una conversación finalizada con timestamp posterior a su finalización
- **THEN** el sistema retorna ErrConversationNotAcceptingMessages

### Requirement: Caso de uso: agente lee conversación

El sistema SHALL marcar como leídos los mensajes del contacto sin leer en una conversación asignada al agente. El sistema SHALL considerar únicamente mensajes del contacto, y SHALL operar también sobre conversaciones finalizadas.

#### Scenario: Marcado de mensajes no leídos

- **WHEN** el agente asignado lee una conversación con mensajes del contacto sin leer
- **THEN** esos mensajes quedan con estado read y timestamp de lectura

#### Scenario: Sin mensajes no leídos

- **WHEN** el agente asignado lee una conversación sin mensajes del contacto sin leer
- **THEN** no cambia ningún mensaje y no se retorna error

#### Scenario: Mensajes del agente no se marcan

- **WHEN** el agente lee una conversación con mensajes suyos y del contacto
- **THEN** solo los mensajes del contacto se marcan como leídos

#### Scenario: Lectura en conversación finalizada

- **WHEN** el agente asignado lee una conversación finalizada
- **THEN** los mensajes del contacto sin leer se marcan igualmente

#### Scenario: Agente no asignado

- **WHEN** un agente intenta leer una conversación asignada a otro agente
- **THEN** el sistema retorna ErrConversationAgentNotOwner

### Requirement: Caso de uso: expirar conversación

El sistema SHALL cerrar por expiración una conversación no finalizada, asignando su fecha de finalización. Expirar una conversación ya finalizada SHALL ser idempotente (sin error).

#### Scenario: Expiración de conversación abierta

- **WHEN** se expira una conversación pending o assigned
- **THEN** el estado cambia a expired y se asigna finishedAt

#### Scenario: Expiración idempotente

- **WHEN** se intenta expirar una conversación ya expired o resolved
- **THEN** la conversación no cambia y no se retorna error

#### Scenario: Conversación inexistente

- **WHEN** se intenta expirar una conversación que no existe
- **THEN** el sistema retorna ErrConversationNotFound

### Requirement: Caso de uso: resolver conversación

El sistema SHALL cerrar una conversación porque el agente asignado la resolvió, validando la existencia del agente y asignando su fecha de finalización. Resolver una conversación ya finalizada SHALL ser idempotente (sin error).

#### Scenario: Resolución exitosa

- **WHEN** el agente asignado resuelve una conversación no finalizada
- **THEN** el estado cambia a resolved y se asigna finishedAt

#### Scenario: Agente no asignado

- **WHEN** un agente intenta resolver una conversación asignada a otro agente o sin agente
- **THEN** el sistema retorna ErrConversationAgentNotOwner

#### Scenario: Resolución idempotente

- **WHEN** se intenta resolver una conversación ya expired o resolved
- **THEN** la conversación no cambia y no se retorna error

#### Scenario: Agente inexistente

- **WHEN** un identificador de agente no registrado intenta resolver una conversación
- **THEN** el sistema retorna error de agente no encontrado

### Requirement: Errores de casos de uso

Cada caso de uso de conversaciones SHALL declarar un error base y todo error retornado SHALL poder identificarse con ese error base mediante `errors.Is`.

#### Scenario: Error de caso de uso identificable

- **WHEN** un caso de uso falla por un error de orquestación o de dominio no idempotente
- **THEN** el error retornado satisface `errors.Is` contra el error base del caso de uso

#### Scenario: Errores idempotentes no se propagan

- **WHEN** el error de dominio corresponde a una operación ya satisfecha del caso de uso
- **THEN** el caso de uso retorna nil

### Requirement: Búsquedas dedicadas del repositorio de conversaciones

El repositorio de conversaciones SHALL exponer búsquedas dedicadas por caso de uso que recuperan únicamente los datos que la operación necesita, incluyendo: conversación sin mensajes, conversación con solo los mensajes no leídos del contacto, conversación con un mensaje por identificador externo, y conversación activa por contacto.

#### Scenario: Carga sin mensajes

- **WHEN** una operación solo necesita los metadatos de la conversación
- **THEN** el repositorio recupera la conversación sin cargar sus mensajes

#### Scenario: Carga de no leídos

- **WHEN** el agente lee una conversación
- **THEN** el repositorio recupera únicamente los mensajes del contacto sin leer

#### Scenario: Búsqueda por identificador externo

- **WHEN** un caso de uso opera sobre un mensaje por su identificador externo
- **THEN** el repositorio recupera la conversación y ese mensaje

### Requirement: Rehidratación de conversación y mensaje

El sistema SHALL poder rehidratar una conversación y un mensaje directamente desde datos persistidos, sin validar sus invariantes y permitiendo cero o más mensajes. La rehidratación SHALL ser distinta de la creación, que sí valida y exige al menos un mensaje en la conversación.

#### Scenario: Rehidratar conversación sin mensajes

- **WHEN** se rehidrata una conversación existente a partir de solo sus metadatos
- **THEN** se obtiene una conversación válida sin mensajes cargados

#### Scenario: Rehidratar conversación con mensajes

- **WHEN** se rehidrata una conversación con un subconjunto de mensajes
- **THEN** los mensajes se cargan y su índice de identificadores externos se reconstruye

#### Scenario: Rehidratar mensaje

- **WHEN** se rehidrata un mensaje desde datos persistidos
- **THEN** el mensaje se obtiene con sus campos tal cual, sin validar invariantes

### Requirement: Asignar identificador externo a mensaje de agente

El sistema SHALL permitir asignar el identificador externo devuelto por el canal a un mensaje de agente ya registrado en la conversación, actualizando su índice.

#### Scenario: Asignación exitosa

- **WHEN** el canal devuelve un identificador externo tras enviar un mensaje de agente
- **THEN** el mensaje queda con ese identificador externo y el índice de la conversación se actualiza

#### Scenario: Mensaje inexistente

- **WHEN** se intenta asignar un identificador externo a un mensaje que no existe en la conversación
- **THEN** el sistema retorna ErrMessageNotFound

### Requirement: Conversación no encontrada

El sistema SHALL exponer un error de dominio específico para indicar que una conversación no existe.

#### Scenario: Uso del error

- **WHEN** un caso de uso no encuentra la conversación que necesita
- **THEN** el error retornado satisface `errors.Is` contra ErrConversationNotFound

### Requirement: Puerto de envío saliente de agente

El dominio de conversaciones SHALL declarar un puerto de salida `AgentMessageSender` que abstrae el envío de un mensaje de agente al canal externo y devuelve su identificador externo.

#### Scenario: Envío a través del puerto

- **WHEN** el caso de uso de envío de agente invoca el puerto
- **THEN** el puerto entrega el mensaje al canal y retorna su identificador externo o un error

## MODIFIED Requirements

### Requirement: Crear mensaje

El sistema SHALL permitir crear un nuevo mensaje con todos sus campos requeridos, incluyendo un identificador externo opcional.

#### Scenario: Crear mensaje completo

- **WHEN** se crea un mensaje con id, status, type, text, agentID, contactID, externalID, sentAt, readAt, editedAt, deletedAt
- **THEN** se retorna un mensaje con todos los campos asignados

#### Scenario: Crear mensaje sin identificador externo

- **WHEN** se crea un mensaje con identificador externo nil
- **THEN** el mensaje se crea sin identificador externo

#### Scenario: Fallar al crear mensaje con identificador externo vacío

- **WHEN** se intenta crear un mensaje con un identificador externo que no es nil pero está vacío
- **THEN** el sistema retorna ErrMessageExternalIDInvalid

#### Scenario: Fallar al crear mensaje con uuid invalido

- **WHEN** se intenta crear un mensaje con un id Nil
- **THEN** el sistema retorna ErrMessageInvalidID

#### Scenario: Fallar al crear mensaje con status invalido

- **WHEN** se intenta crear un mensaje con un status que no es sent, read, deleted o failed
- **THEN** el sistema retorna ErrMessageStatusInvalid

#### Scenario: Fallar al crear mensaje con type invalido

- **WHEN** se intenta crear un mensaje con un type que no es text
- **THEN** el sistema retorna ErrMessageTypeInvalid

#### Scenario: Fallar al crear mensaje con texto vacio para tipo text

- **WHEN** se intenta crear un mensaje de tipo text con texto nil o vacio
- **THEN** el sistema retorna ErrMessageEmptyText

#### Scenario: Fallar al crear mensaje con texto muy largo

- **WHEN** se intenta crear un mensaje de tipo text con más de 1000 caracteres visuales
- **THEN** el sistema retorna ErrMessageTextTooLong
