# Spec Delta

## ADDED Requirements

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

## MODIFIED Requirements

### Requirement: Estado de mensaje
El sistema SHALL soportar los estados de mensaje: sent, read, deleted, failed.

#### Scenario: Estado de mensaje nuevo
- **WHEN** se crea un mensaje
- **THEN** su estado es sent

#### Scenario: Estado failed es terminal
- **WHEN** un mensaje queda en estado failed
- **THEN** el mensaje no acepta ediciones ni eliminaciones