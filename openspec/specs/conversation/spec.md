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
El sistema SHALL permitir crear un nuevo mensaje con todos sus campos requeridos.

#### Scenario: Crear mensaje completo
- **WHEN** se crea un mensaje con id, status, type, text, agentID, contactID, sentAt, readAt, editedAt, deletedAt
- **THEN** se retorna un mensaje con todos los campos asignados

#### Scenario: Fallar al crear mensaje con uuid invalido
- **WHEN** se intenta crear un mensaje con un id Nil
- **THEN** el sistema retorna ErrMessageInvalidID

#### Scenario: Fallar al crear mensaje con status invalido
- **WHEN** se intenta crear un mensaje con un status que no es sent, read o deleted
- **THEN** el sistema retorna ErrMessageStatusInvalid

#### Scenario: Fallar al crear mensaje con type invalido
- **WHEN** se intenta crear un mensaje con un type que no es text
- **THEN** el sistema retorna ErrMessageTypeInvalid

#### Scenario: Fallar al crear mensaje con texto vacio para tipo text
- **WHEN** se intenta crear un mensaje de tipo text con texto nil o vacio
- **THEN** el sistema retorna ErrMessageEmptyText

#### Scenario: Fallar al crear mensaje con texto muy largo
- **WHEN** se intenta crear un mensaje de tipo text con mas de 1000 runas
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

### Requirement: Editar mensaje de texto
El sistema SHALL permitir editar el texto de un mensaje existente de tipo texto. Solo el dueño del mensaje puede editarlo. Si llega una edición con timestamp anterior o igual a la última edición aplicada, el sistema SHALL conservar la edición más reciente.

#### Scenario: Agente edita mensaje exitosamente
- **WHEN** un agente edita un mensaje de texto que le pertenece en una conversación activa
- **THEN** el texto del mensaje se actualiza y editedAt se asigna

#### Scenario: Contacto edita mensaje exitosamente
- **WHEN** un contacto edita un mensaje de texto que le pertenece por externalID
- **THEN** el texto del mensaje se actualiza y editedAt se asigna

#### Scenario: Descartar edición anterior a la última edición
- **WHEN** se recibe una edición con timestamp anterior o igual a la última edición aplicada
- **THEN** el mensaje conserva el texto y editedAt de la edición más reciente, sin retornar error

#### Scenario: Fallar si mensaje no es de texto
- **WHEN** se intenta editar un mensaje que no tiene texto
- **THEN** el sistema retorna ErrMessageNotText

#### Scenario: Fallar si el nuevo texto está vacío
- **WHEN** se intenta editar un mensaje con texto vacío
- **THEN** el sistema retorna ErrMessageEmptyText

#### Scenario: Fallar si el nuevo texto excede el máximo
- **WHEN** se intenta editar un mensaje con texto de más de 1000 caracteres
- **THEN** el sistema retorna ErrMessageTextTooLong

#### Scenario: Fallar si mensaje está eliminado y timestamp es posterior
- **WHEN** se intenta editar un mensaje eliminado con timestamp posterior a deletedAt
- **THEN** el sistema retorna ErrMessageAlreadyDeleted

#### Scenario: Permitir edit con timestamp anterior a eliminación
- **WHEN** se intenta editar un mensaje eliminado con timestamp anterior a deletedAt
- **THEN** la edición se aplica como traza del estado anterior

#### Scenario: Fallar si mensaje está en estado failed
- **WHEN** se intenta editar un mensaje en estado failed
- **THEN** el sistema retorna ErrMessageFailed

#### Scenario: Fallar si el mensaje no existe
- **WHEN** un agente intenta editar un mensaje con un id que no existe en la conversación
- **THEN** el sistema retorna ErrMessageNotFound

#### Scenario: Fallar si el externalID no existe
- **WHEN** un contacto intenta editar un mensaje con un externalID que no existe en la conversación
- **THEN** el sistema retorna ErrMessageNotFound

#### Scenario: Fallar si el mensaje no pertenece al agente
- **WHEN** un agente intenta editar un mensaje enviado por el contacto
- **THEN** el sistema retorna ErrConversationAgentNotOwner

#### Scenario: Fallar si el mensaje no pertenece al contacto
- **WHEN** un contacto intenta editar por externalID un mensaje enviado por el agente
- **THEN** el sistema retorna ErrConversationContactNotOwner

#### Scenario: Fallar si agente intenta editar en conversación finalizada
- **WHEN** un agente intenta editar un mensaje en una conversación expired o resolved
- **THEN** el sistema retorna ErrConversationFinished

#### Scenario: Contacto puede editar en conversación finalizada
- **WHEN** un contacto edita un mensaje en una conversación expired o resolved
- **THEN** la edición se aplica normalmente

### Requirement: Eliminar mensaje
El sistema SHALL permitir eliminar un mensaje existente (soft delete). Solo el dueño del mensaje puede eliminarlo.

#### Scenario: Agente elimina mensaje exitosamente
- **WHEN** un agente elimina un mensaje que le pertenece en una conversación activa
- **THEN** el mensaje queda con status deleted y deletedAt asignado

#### Scenario: Contacto elimina mensaje exitosamente
- **WHEN** un contacto elimina un mensaje que le pertenece por externalID
- **THEN** el mensaje queda con status deleted y deletedAt asignado

#### Scenario: Fallar si mensaje ya está eliminado
- **WHEN** se intenta eliminar un mensaje que ya tiene deletedAt, sin importar el timestamp
- **THEN** el sistema retorna ErrMessageAlreadyDeleted

#### Scenario: Fallar si mensaje está en estado failed
- **WHEN** se intenta eliminar un mensaje en estado failed
- **THEN** el sistema retorna ErrMessageFailed

#### Scenario: Fallar si el mensaje no existe
- **WHEN** un agente intenta eliminar un mensaje con un id que no existe en la conversación
- **THEN** el sistema retorna ErrMessageNotFound

#### Scenario: Fallar si el externalID no existe
- **WHEN** un contacto intenta eliminar un mensaje con un externalID que no existe en la conversación
- **THEN** el sistema retorna ErrMessageNotFound

#### Scenario: Fallar si el mensaje no pertenece al agente
- **WHEN** un agente intenta eliminar un mensaje enviado por el contacto
- **THEN** el sistema retorna ErrConversationAgentNotOwner

#### Scenario: Fallar si el mensaje no pertenece al contacto
- **WHEN** un contacto intenta eliminar por externalID un mensaje enviado por el agente
- **THEN** el sistema retorna ErrConversationContactNotOwner

#### Scenario: Fallar si agente intenta eliminar en conversación finalizada
- **WHEN** un agente intenta eliminar un mensaje en una conversación expired o resolved
- **THEN** el sistema retorna ErrConversationFinished

#### Scenario: Contacto puede eliminar en conversación finalizada
- **WHEN** un contacto elimina un mensaje en una conversación expired o resolved
- **THEN** la eliminación se aplica normalmente

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
