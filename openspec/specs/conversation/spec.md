# Conversation Module

## Purpose

El modulo de conversaciones permite a los contactos iniciar conversaciones con la empresa para soporte, ventas, cambios, envios y otras interacciones cortas. Cada conversacion tiene un solo agente activo y puede expirar segun la configuracion de la plataforma.

## Requirements

### Requirement: Crear conversacion
El sistema SHALL permitir crear una nueva conversacion con todos sus campos requeridos. Si el estado es expired o resolved, finishedAt es obligatorio.

#### Scenario: Crear conversacion con mensajes
- **WHEN** se crea una conversacion con id, status, mensajes, agentID, contactID, createdAt, updatedAt, finishedAt
- **THEN** se retorna una conversacion con todos los campos asignados y los mensajes almacenados internamente

#### Scenario: Fallar al crear conversacion con uuid invalido
- **WHEN** se intenta crear una conversacion con un id Nil
- **THEN** el sistema retorna ErrConversationInvalidID

#### Scenario: Fallar al crear conversacion sin contactID ni agentID
- **WHEN** se intenta crear una conversacion con contactID y agentID ambos nil
- **THEN** el sistema retorna ErrConversationMissingContactAndAgent

#### Scenario: Fallar al crear conversacion con status invalido
- **WHEN** se intenta crear una conversacion con un status que no es pending, assigned, expired o resolved
- **THEN** el sistema retorna ErrConversationStatusInvalid

#### Scenario: Fallar al crear conversacion finalizada sin finishedAt
- **WHEN** se intenta crear una conversacion con estado expired o resolved y finishedAt es nil
- **THEN** el sistema retorna ErrConversationFinishedAtMissing

### Requirement: Obtener mensajes de conversacion
El sistema SHALL retornar los mensajes de una conversacion como un slice ordenado por sentAt ascendente.

#### Scenario: Obtener mensajes
- **WHEN** se solicitan los mensajes de una conversacion
- **THEN** se retorna un slice con todos los mensajes ordenados por sentAt ascendente

#### Scenario: Desempatar mensajes con el mismo sentAt
- **WHEN** dos mensajes tienen el mismo sentAt
- **THEN** el orden se desempata por id de forma determinista

### Requirement: Estado de conversacion
El sistema SHALL soportar los estados: pending, assigned, expired, resolved.

#### Scenario: Estado inicial
- **WHEN** se crea una conversacion
- **THEN** su estado es pending o assigned segun el agente asignado

### Requirement: Estado de mensaje
El sistema SHALL soportar los estados de mensaje: sent, read, deleted, failed.

#### Scenario: Estado de mensaje nuevo
- **WHEN** se crea un mensaje
- **THEN** su estado es sent

#### Scenario: Estado failed es terminal
- **WHEN** un mensaje queda en estado failed
- **THEN** el mensaje no acepta ediciones ni eliminaciones

### Requirement: Tipo de mensaje
El sistema SHALL soportar el tipo de mensaje text para MVP.

#### Scenario: Mensaje de texto
- **WHEN** se crea un mensaje de tipo text
- **THEN** el mensaje tiene un campo text con contenido

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

### Requirement: Recibir mensaje de contacto
El sistema SHALL permitir recibir un mensaje de un contacto en una conversacion existente, validando que la conversacion acepte mensajes, que el contacto pertenezca a la conversacion y que el mensaje no este duplicado.

#### Scenario: Recibir mensaje exitosamente
- **WHEN** se recibe un mensaje de un contacto que pertenece a la conversacion
- **THEN** el mensaje se agrega a la conversacion y el timestamp de actualizacion se actualiza

#### Scenario: Fallar si la conversacion no acepta mensajes
- **WHEN** se recibe un mensaje en una conversacion que no acepta mensajes en ese momento
- **THEN** el sistema retorna ErrConversationNotAcceptingMessages

#### Scenario: Fallar si el contacto no pertenece a la conversacion
- **WHEN** se recibe un mensaje de un contacto que no es el asignado a la conversacion
- **THEN** el sistema retorna ErrConversationContactNotOwner

#### Scenario: Fallar si la conversacion no tiene contacto
- **WHEN** se recibe un mensaje pero la conversacion no tiene un contacto asignado
- **THEN** el sistema retorna ErrConversationHasNoContact

#### Scenario: Fallar si el mensaje ya existe (duplicado por externalID)
- **WHEN** se recibe un mensaje con un externalID que ya existe en la conversacion
- **THEN** el sistema retorna ErrConversationDuplicateMessage

#### Scenario: Recibir mensaje sin externalID
- **WHEN** se recibe un mensaje que no tiene externalID asignado
- **THEN** el mensaje se agrega a la conversacion sin validar duplicados por externalID

### Requirement: Identificador externo de mensaje
El sistema SHALL permitir asignar un identificador externo a un mensaje que aun no tiene uno. Una vez asignado, el identificador es inmutable.

#### Scenario: Asignar externalID exitosamente
- **WHEN** se asigna un externalID a un mensaje que no tiene uno
- **THEN** el mensaje queda con el externalID asignado

#### Scenario: Fallar si el externalID ya fue asignado
- **WHEN** se intenta asignar un externalID a un mensaje que ya tiene uno
- **THEN** el sistema retorna ErrMessageExternalIDAlreadySet

#### Scenario: Fallar si el externalID es vacio
- **WHEN** se intenta asignar un externalID con valor vacio
- **THEN** el sistema retorna ErrMessageExternalIDInvalid

### Requirement: Conversacion requiere al menos un mensaje
El sistema SHALL exigir que toda conversacion tenga al menos un mensaje. No es posible crear una conversacion sin mensajes.

#### Scenario: Fallar al crear conversacion sin mensajes
- **WHEN** se intenta crear una conversacion con un slice de mensajes vacio o nil
- **THEN** el sistema retorna ErrConversationEmptyMessages

#### Scenario: Crear conversacion con un mensaje
- **WHEN** se crea una conversacion con al menos un mensaje
- **THEN** la conversacion se crea exitosamente con el mensaje almacenado

### Requirement: Enviar mensaje de agente
El sistema SHALL permitir a un agente enviar un mensaje a una conversacion existente. Si la conversacion esta pendiente sin agente asignado, el agente se asigna automaticamente y la conversacion cambia a estado assigned.

#### Scenario: Enviar mensaje exitosamente
- **WHEN** un agente envia un mensaje a una conversacion asignada a el
- **THEN** el mensaje se agrega a la conversacion y el timestamp de actualizacion se actualiza

#### Scenario: Reclamar conversacion pendiente
- **WHEN** un agente envia una mensaje a una conversacion en estado pending sin agente asignado
- **THEN** el agente se asigna a la conversacion, el estado cambia a assigned y el mensaje se agrega

#### Scenario: Fallar si el agente no es el asignado
- **WHEN** un agente intenta enviar un mensaje a una conversacion asignada a otro agente
- **THEN** el sistema retorna ErrConversationAgentNotOwner

#### Scenario: Sincronizar externalID si existe
- **WHEN** un agente envia un mensaje con externalID
- **THEN** el indice de externalID se actualiza con el nuevo mensaje

### Requirement: Validar que conversacion acepta mensajes
El sistema SHALL validar que una conversacion acepta mensajes en un momento dado, considerando su estado y fecha de finalizacion.

#### Scenario: Conversacion abierta acepta mensajes
- **WHEN** se intenta enviar o recibir un mensaje en una conversacion en estado pending o assigned
- **THEN** el sistema acepta el mensaje

#### Scenario: Conversacion cerrada rechaza mensajes tardios
- **WHEN** se intenta enviar o recibir un mensaje en una conversacion expired o resolved con timestamp posterior a finishedAt
- **THEN** el sistema retorna ErrConversationNotAcceptingMessages

#### Scenario: Conversacion cerrada acepta mensajes enviados antes de finalizar
- **WHEN** se intenta recibir un mensaje en una conversacion expired o resolved con timestamp anterior a finishedAt
- **THEN** el sistema acepta el mensaje (mensaje tardio valido)

#### Scenario: Fallar si conversacion finalizada no tiene finishedAt
- **WHEN** una conversacion tiene estado expired o resolved pero finishedAt es nil
- **THEN** el sistema retorna ErrConversationFinishedAtMissing

### Requirement: Mensajes son punteros
El sistema SHALL usar punteros a mensajes (*Message) en vez de valores (Message) para mantener la identidad de la entidad.

#### Scenario: Crear conversacion con mensajes como punteros
- **WHEN** se crea una conversacion con un slice de punteros a mensajes
- **THEN** la conversacion almacena los mensajes como punteros

#### Scenario: Recibir mensaje como puntero
- **WHEN** se recibe un mensaje de contacto como puntero
- **THEN** el mensaje se almacena como puntero en la conversacion

#### Scenario: Enviar mensaje como puntero
- **WHEN** un agente envia un mensaje como puntero
- **THEN** el mensaje se almacena como puntero en la conversacion

### Requirement: Agente puede leer conversacion
El sistema SHALL permitir a un agente marcar mensajes del contacto como leídos en una conversacion existente. El agente solo puede marcar como leídos mensajes que el contacto envió.

#### Scenario: Marcar mensajes como leídos exitosamente
- **WHEN** un agente lee una conversacion asignada a el con mensajes del contacto sin leer
- **THEN** los mensajes del contacto se marcan como leídos con timestamp y status actualizados

#### Scenario: Fallar si el agente no es el asignado
- **WHEN** un agente intenta leer una conversacion asignada a otro agente
- **THEN** el sistema retorna ErrConversationAgentNotOwner

#### Scenario: No marcar mensajes del agente como leídos
- **WHEN** un agente lee una conversacion con mensajes del agente y del contacto
- **THEN** solo los mensajes del contacto se marcan como leídos, los mensajes del agente no cambian

#### Scenario: Marcar mensajes en conversacion finalizada
- **WHEN** un agente lee una conversacion en estado expired o resolved
- **THEN** los mensajes del contacto se marcan como leídos normalmente

#### Scenario: No marcar mensajes ya leídos
- **WHEN** un agente lee una conversacion con mensajes del contacto ya leídos
- **THEN** los mensajes ya leídos no cambian, solo se marcan los no leídos

### Requirement: Mensaje tiene exactamente un dueño
El sistema SHALL exigir que todo mensaje tenga exactamente un dueño: o el agente o el contacto, nunca ambos ni ninguno.

#### Scenario: Crear mensaje del agente
- **WHEN** se crea un mensaje con agentID y sin contactID
- **THEN** el mensaje se crea exitosamente

#### Scenario: Crear mensaje del contacto
- **WHEN** se crea un mensaje con contactID y sin agentID
- **THEN** el mensaje se crea exitosamente

#### Scenario: Fallar si mensaje no tiene dueño
- **WHEN** se intenta crear un mensaje sin agentID y sin contactID
- **THEN** el sistema retorna ErrMessageInvalidOwner

#### Scenario: Fallar si mensaje tiene dos dueños
- **WHEN** se intenta crear un mensaje con agentID y contactID
- **THEN** el sistema retorna ErrMessageInvalidOwner

### Requirement: Marcar mensaje como leído
El sistema SHALL permitir marcar un mensaje como leído, cambiando su status a read y asignando el timestamp de lectura.

#### Scenario: Marcar mensaje como leído
- **WHEN** se marca un mensaje como leído con un timestamp
- **THEN** el mensaje queda con status read y readAt asignado

### Requirement: Marcar mensaje como fallido
El sistema SHALL permitir marcar como fallido un mensaje enviado por el agente cuando la infraestructura no pudo entregarlo. El estado failed es terminal y bloquea todas las modificaciones.

#### Scenario: Marcar mensaje de agente como fallido
- **WHEN** la infraestructura reporta que un mensaje enviado por el agente no se pudo entregar
- **THEN** el mensaje queda con status failed

#### Scenario: Marcar como fallido es idempotente
- **WHEN** se marca como fallido un mensaje que ya tiene status failed
- **THEN** el estado no cambia y no se retorna error

#### Scenario: Fallar si el mensaje no fue enviado por el agente
- **WHEN** se intenta marcar como fallido un mensaje enviado por el contacto
- **THEN** el sistema retorna ErrMessageNotFromAgent

#### Scenario: Fallar si el agente no es el asignado
- **WHEN** un agente intenta reportar el fallo de un mensaje de una conversación asignada a otro agente
- **THEN** el sistema retorna ErrConversationAgentNotOwner

#### Scenario: Fallar si el mensaje no existe
- **WHEN** se intenta marcar como fallido un mensaje con un id que no existe en la conversación
- **THEN** el sistema retorna ErrMessageNotFound

### Requirement: Timestamp de actualizacion monotono
El sistema SHALL actualizar el timestamp de actualización de la conversación solo cuando el nuevo timestamp sea más reciente que el actual.

#### Scenario: Actualizar con timestamp más reciente
- **WHEN** ocurre una operación con un timestamp posterior a updatedAt
- **THEN** updatedAt se actualiza al nuevo timestamp

#### Scenario: No retroceder con timestamp anterior
- **WHEN** llega un mensaje tardío con un timestamp anterior a updatedAt
- **THEN** updatedAt conserva el valor más reciente

### Requirement: Expirar conversacion
El sistema SHALL permitir cerrar una conversacion por expiracion, asignando su fecha de finalizacion. Puede expirar una conversacion en estado pending o assigned. Una conversacion ya finalizada no puede expirarse.

#### Scenario: Expirar conversacion pending
- **WHEN** se expira una conversacion en estado pending
- **THEN** el estado cambia a expired y finishedAt se asigna

#### Scenario: Expirar conversacion assigned
- **WHEN** se expira una conversacion en estado assigned
- **THEN** el estado cambia a expired y finishedAt se asigna

#### Scenario: Fallar si la conversacion ya esta finalizada
- **WHEN** se intenta expirar una conversacion en estado expired o resolved
- **THEN** el sistema retorna ErrConversationFinished

#### Scenario: Expirar registra actividad
- **WHEN** se expira una conversacion
- **THEN** updatedAt se actualiza si el timestamp es mas reciente

### Requirement: Resolver conversacion
El sistema SHALL permitir cerrar una conversacion porque el agente la resolvio, asignando su fecha de finalizacion. Solo el agente asignado puede resolverla. Una conversacion ya finalizada no puede resolverse.

#### Scenario: Resolver conversacion exitosamente
- **WHEN** el agente asignado resuelve una conversacion en estado assigned
- **THEN** el estado cambia a resolved y finishedAt se asigna

#### Scenario: Fallar si el agente no es el asignado
- **WHEN** un agente intenta resolver una conversacion asignada a otro agente
- **THEN** el sistema retorna ErrConversationAgentNotOwner

#### Scenario: Fallar si la conversacion no tiene agente
- **WHEN** se intenta resolver una conversacion en estado pending sin agente asignado
- **THEN** el sistema retorna ErrConversationAgentNotOwner

#### Scenario: Fallar si la conversacion ya esta finalizada
- **WHEN** se intenta resolver una conversacion en estado expired o resolved
- **THEN** el sistema retorna ErrConversationFinished

#### Scenario: Resolver registra actividad
- **WHEN** se resuelve una conversacion
- **THEN** updatedAt se actualiza si el timestamp es mas reciente

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

El sistema SHALL aplicar la edición de un mensaje enviada por el contacto, localizándolo por su identificador externo dentro de la conversación. Un agente no SHALL poder editar mensajes. Las ediciones con timestamp anterior o igual a la última edición aplicada SHALL descartarse sin error. El sistema no SHALL rechazar la edición por el estado finalizado de la conversación. El sistema SHALL validar el contenido: solo mensajes de tipo texto, texto no vacío, como máximo 1000 caracteres, y mensajes no fallidos.

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

#### Scenario: Editar un mensaje sin texto

- **WHEN** el contacto edita un mensaje que no tiene texto
- **THEN** el sistema retorna ErrMessageNotText

#### Scenario: Editar con texto vacío

- **WHEN** el contacto edita un mensaje con texto vacío
- **THEN** el sistema retorna ErrMessageEmptyText

#### Scenario: Editar con texto que excede el máximo

- **WHEN** el contacto edita un mensaje con texto de más de 1000 caracteres
- **THEN** el sistema retorna ErrMessageTextTooLong

#### Scenario: Editar un mensaje eliminado con timestamp posterior

- **WHEN** el contacto edita un mensaje eliminado con timestamp posterior a deletedAt
- **THEN** el sistema retorna ErrMessageAlreadyDeleted

#### Scenario: Editar un mensaje eliminado con timestamp anterior

- **WHEN** el contacto edita un mensaje eliminado con timestamp anterior a deletedAt
- **THEN** la edición se aplica como traza del estado anterior

#### Scenario: Editar un mensaje fallido

- **WHEN** el contacto edita un mensaje en estado failed
- **THEN** el sistema retorna ErrMessageFailed

### Requirement: Caso de uso: eliminar mensaje de contacto

El sistema SHALL aplicar la eliminación (soft delete) de un mensaje enviada por el contacto, localizándolo por su identificador externo. Un agente no SHALL poder eliminar mensajes. Un mensaje ya eliminado SHALL tratarse como idempotente (sin error). El sistema no SHALL rechazar la eliminación por el estado finalizado de la conversación. El sistema SHALL rechazar eliminar un mensaje en estado failed.

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

#### Scenario: Eliminar un mensaje fallido

- **WHEN** el contacto elimina un mensaje en estado failed
- **THEN** el sistema retorna ErrMessageFailed

### Requirement: Caso de uso: marcar como leído un mensaje del agente por el contacto

El sistema SHALL marcar como leído un mensaje del agente a partir de un acuse de lectura del contacto, localizándolo por su identificador externo dentro de la conversación. El sistema no SHALL rechazar el acuse por el estado finalizado de la conversación. Los acuses con timestamp anterior o igual a la última lectura registrada SHALL descartarse sin error. El sistema SHALL rechazar el acuse si el contacto no pertenece a la conversación o si el mensaje no fue enviado por un agente. El acuse no SHALL reactivar un mensaje eliminado: SHALL registrar la lectura conservando su estado eliminado. Un mensaje fallido SHALL permanecer sin cambios.

#### Scenario: Lectura aplicada

- **WHEN** el contacto envía un acuse de lectura de un mensaje del agente no leído
- **THEN** el mensaje queda con su timestamp de lectura y estado read

#### Scenario: Acuse obsoleto

- **WHEN** el contacto envía un acuse con timestamp anterior o igual al último registrado
- **THEN** el mensaje conserva su lectura más reciente y no se retorna error

#### Scenario: Identificador externo inexistente

- **WHEN** el contacto envía un acuse con un identificador externo que no pertenece a la conversación
- **THEN** el sistema retorna ErrMessageNotFound

#### Scenario: Mensaje no pertenece al agente

- **WHEN** el contacto envía un acuse sobre un mensaje enviado por el propio contacto
- **THEN** el sistema retorna ErrConversationContactNotOwner

#### Scenario: Lectura en conversación finalizada

- **WHEN** el contacto envía un acuse sobre una conversación finalizada
- **THEN** la lectura se aplica igualmente

#### Scenario: Acuse sobre mensaje eliminado

- **WHEN** el contacto envía un acuse sobre un mensaje del agente eliminado
- **THEN** la lectura se registra y el mensaje conserva su estado eliminado

#### Scenario: Acuse sobre mensaje fallido

- **WHEN** el contacto envía un acuse sobre un mensaje del agente fallido
- **THEN** el mensaje permanece sin cambios y sin error

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

El repositorio de conversaciones SHALL exponer búsquedas dedicadas por caso de uso que recuperan únicamente los datos que la operación necesita, incluyendo: conversación sin mensajes, conversación con solo los mensajes no leídos del contacto, conversación con un mensaje por identificador externo, y conversación activa por contacto. La búsqueda de la conversación activa SHALL incluir los identificadores externos de sus mensajes para permitir la detección de duplicados.

#### Scenario: Carga sin mensajes

- **WHEN** una operación solo necesita los metadatos de la conversación
- **THEN** el repositorio recupera la conversación sin cargar sus mensajes

#### Scenario: Carga de no leídos

- **WHEN** el agente lee una conversación
- **THEN** el repositorio recupera únicamente los mensajes del contacto sin leer

#### Scenario: Búsqueda por identificador externo

- **WHEN** un caso de uso opera sobre un mensaje por su identificador externo
- **THEN** el repositorio recupera la conversación y ese mensaje

#### Scenario: Carga de identificadores externos de la conversación activa

- **WHEN** un caso de uso necesita detectar mensajes duplicados en la conversación activa
- **THEN** el repositorio recupera la conversación activa junto con los identificadores externos de sus mensajes

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

### Requirement: Seguimiento de mensajes modificados

El sistema SHALL rastrear los mensajes modificados desde que la conversación fue cargada. La creación SHALL marcar todos sus mensajes iniciales como modificados. La rehidratación SHALL NOT marcar ninguno. Cada mutación SHALL marcar el mensaje afectado. El sistema SHALL exponer los mensajes modificados en orden determinista por `sentAt` ascendente y, en empate, por `id`.

#### Scenario: Creación marca los mensajes iniciales

- **WHEN** se crea una conversación con mensajes
- **THEN** todos sus mensajes quedan marcados como modificados

#### Scenario: Rehidratación no marca mensajes

- **WHEN** se rehidrata una conversación con mensajes
- **THEN** ningún mensaje queda marcado como modificado

#### Scenario: Mutación marca el mensaje afectado

- **WHEN** una operación modifica un mensaje
- **THEN** solo ese mensaje queda marcado como modificado

#### Scenario: Orden determinista de los modificados

- **WHEN** se exponen los mensajes modificados
- **THEN** vienen ordenados por `sentAt` ascendente y, en empate, por `id`
