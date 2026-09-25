# Design

## Context

El modulo de conversaciones actual tiene un aggregate `Conversation` con un mapa `found map[uuid.UUID]Message` que almacena mensajes cargados desde la base de datos, y un mapa `dirty` para mensajes pendientes de persistir. No existe capacidad de recibir mensajes de contactos ni identificadores externos de plataformas.

Ver `proposal.md` para la motivacion del cambio.

## Goals / Non-Goals

**Goals:**
- Permitir recibir mensajes de contactos en conversaciones existentes
- Almacenar identificadores externos de plataformas en mensajes
- Garantizar que toda conversacion tenga al menos un mensaje
- Busquedas O(1) por UUID y por externalID en el aggregate

**Non-Goals:**
- Persistencia (infra, repositorios, migraciones, SQLC) — solo dominio
- Creacion automatica de conversaciones al recibir primer mensaje (use case)
- Identificador externo de conversacion (campo separado, otro cambio)
- Handlers HTTP o adaptadores de plataforma

## Decisions

### 1. externalID como `*string` en Message

El identificador externo es un campo nullable porque no todos los mensajes provienen de una plataforma (ej: mensajes del agente creados internamente). Es `string` porque las plataformas usan strings, no UUIDs.

**Alternativas consideradas:**
- `uuid.UUID` — Rechazado: las plataformas no usan UUIDs estandar
- Value object `ExternalID` con platform+id — Rechazado: complejidad prematura para MVP, se puede agregar despues

### 2. Asignacion set-once via `AssignExternalID`

El metodo `AssignExternalID(id string) error` en Message permite asignar el externalID solo si es nil. Una vez asignado, es inmutable. Esto refleja la realidad: el ID de plataforma no cambia.

**Por que no un setter generico:**
- En DDD, los metodos hablan lenguaje de negocio. "Set" no dice nada. "Assign" refleja la intencion: asignar un identificador que la plataforma otorgo.
- La inmutabilidad es una regla de negocio, no una restriccion tecnica.

### 3. Mapa + indice para busquedas

Estructura en Conversation:
```
found          map[uuid.UUID]Message    // fuente de verdad
externalMsgIdx map[string]uuid.UUID     // indice: externalID → UUID
```

El indice `externalMsgIdx` es derivado de `found`. Buscar por externalID es O(1): `idx[ext] → found[uuid]`. Al agregar un mensaje con externalID, se actualiza el indice.

**Alternativas consideradas:**
- Un solo mapa, iterar para buscar por externalID — O(n), rechazado por performance
- Dos mapas como fuentes de verdad — Riesgo de inconsistencia, rechazado
- Indice separado para dirty — Complejidad innecesaria, un solo indice cubre ambos

### 4. Invariante: NUNCA 0 mensajes

Validado en `NewConversation`. Una conversacion sin mensajes es un estado invalido del dominio. Toda conversacion nace con al menos un mensaje.

**Por que en el constructor y no solo en el use case:**
- Es un invariante del aggregate, no una regla del caso de uso
- El constructor tambien se usa para reconstitucion desde DB, donde el invariante tambien aplica
- Si la DB tiene una conversacion con 0 mensajes, es un error de datos que debe fallar al cargar

### 5. `ReceiveContactMessage` como metodo del aggregate

El metodo valida:
1. Que la conversacion tenga contacto asignado
2. Que el contacto del mensaje sea el de la conversacion
3. Que el externalID no este duplicado (si tiene uno)

Agrega el mensaje a `found` directamente (no a `dirty`) porque el mensaje ya es un hecho recibido.

**Por que no `dirty`:**
- `dirty` es para cambios pendientes del agente (crear, editar)
- Un mensaje recibido del contacto es un hecho que ya ocurrio, no una operacion pendiente

### 6. Lenguaje DDD en errores

Los mensajes de error reflejan intencion de negocio, no restricciones tecnicas:
- `ErrConversationHasNoContact` — "cannot receive message: conversation has no contact"
- `ErrConversationContactNotOwner` — "cannot receive message: contact does not belong to this conversation"
- `ErrConversationDuplicateMessage` — "cannot receive message: message already exists in conversation"
- `ErrConversationEmptyMessages` — "conversation must have at least one message"
- `ErrMessageExternalIDAlreadySet` — "message external identifier is already assigned"
- `ErrMessageExternalIDInvalid` — "message external identifier is invalid"

## Risks / Trade-offs

- **Indice en memoria** — El `externalMsgIdx` se construye al cargar la conversacion. Para conversaciones con muchos mensajes, el costo de construccion inicial es O(n). Mitigacion: es un costo una vez por carga, y las busquedas subsecuentes son O(1).
- **Mensaje sin externalID** — Un mensaje recibido sin externalID no se puede buscar por ese indice. Mitigacion: es un caso valido (plataformas sin ID), se busca por UUID.
- **externalID como string** — Sin validacion de formato. Mitigacion: la validacion de formato pertenece a la capa de infraestructura (adaptadores de plataforma), no al dominio.
