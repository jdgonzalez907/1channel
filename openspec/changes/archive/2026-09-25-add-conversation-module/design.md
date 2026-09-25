# Design

## Context

1Channel es un CRM conversacional multi-canal construido en Go como un modulo monolitico siguiendo Clean Architecture y DDD. El proyecto esta fresco - `internal/` esta vacio. Necesitamos establecer el primer modulo: conversaciones.

El modulo sigue la estructura definida en AGENTS.md:
```
internal/module/conversations/
├── api.go              # Interfaz publica
├── app/                # Casos de uso
├── domain/             # Agregados, entidades, value objects
└── infra/              # Implementaciones
```

Para este cambio, nos enfocamos solo en el dominio.

## Goals / Non-Goals

**Goals:**
- Definir el aggregate root `Conversation` con su estructura interna
- Definir la entity `Message` con todos sus campos
- Definir los Value Objects para estados y tipos con validacion
- Proporcionar constructores globales con validaciones de negocio
- Definir errores de dominio en archivo centralizado

**Non-Goals:**
- Persistencia (infraestructura)
- Casos de uso (aplicacion)
- Handlers HTTP
- Jobs de expiracion
- Logica de negocio compleja (asignacion, expiracion) - futuras iteracciones

## Decisions

### 1. UUID v7 de Go 1.27 stdlib

**Decision:** Usar el paquete `uuid` de la libreria estandar de Go 1.27.

**Razon:** Go 1.27 incluye soporte nativo para UUIDs. El tipo `uuid.UUID` es `[16]byte` y `uuid.NewV7()` genera UUIDs ordenados por tiempo, ideal para primary keys en PostgreSQL.

**Validacion:** Los constructores validan que el UUID no sea Nil.

**Alternativas consideradas:**
- `github.com/google/uuid`: Paquete popular pero innecesario con Go 1.27

### 2. Estructura interna de Conversation

**Decision:** `found` y `dirty` son maps a nivel de Conversation, no dentro de un VO separado.

```
type Conversation struct {
    id         uuid.UUID
    status     ConversationStatus
    found      map[uuid.UUID]Message
    dirty      map[uuid.UUID]Message
    agentID    *uuid.UUID
    contactID  uuid.UUID
    createdAt  time.Time
    updatedAt  *time.Time
    finishedAt *time.Time
}
```

**Razon:** 
- `found` almacena mensajes persistidos (cargados de DB)
- `dirty` almacena mensajes modificados/nuevos (pendientes de guardar)
- Getter `Messages()` itera `found` y construye slice
- Constructor recibe `[]Message` y llena `found` con `make(map[uuid.UUID]Message, len(messages))`
- `dirty` se inicializa vacio con `make(map[uuid.UUID]Message)`

**Alternativas consideradas:**
- VO Messages separado: Rechazado - complejidad innecesaria
- Slice de mensajes: Rechazado - dificulta tracking de modificaciones

### 3. Nombres de tipos

**Decision:** 
- `ConversationStatus` (no `Status`) para evitar ambigüedad con `MessageStatus`
- `MessageStatus` para el estado de mensajes
- `MessageType` para el tipo de mensaje

**Razon:** Ambos tipos viven en el mismo package `conversation`, necesitan nombres unicos.

### 4. Value Objects como tipos string con validacion

**Decision:** Los Value Objects son tipos `string` con constantes y constructores que validan.

```go
type ConversationStatus string

const (
    StatusPending  ConversationStatus = "pending"
    StatusAssigned ConversationStatus = "assigned"
    StatusExpired  ConversationStatus = "expired"
    StatusResolved ConversationStatus = "resolved"
)

func NewConversationStatus(s string) (ConversationStatus, error) // valida contra constantes
func (s ConversationStatus) String() string
```

**Razon:** Simples, serializables, comparables. Constructores validan que el valor sea del enum.

### 5. Constructores globales con validaciones

**Decision:** `NewConversation(...)` y `NewMessage(...)` son funciones globales con validaciones de negocio.

**Validaciones NewConversation:**
- `id` no debe ser uuid Nil
- `contactID` o `agentID` debe estar presente (al menos uno)

**Validaciones NewMessage:**
- `id` no debe ser uuid Nil
- Si `type` es "text", `text` debe tener entre 1 y 1000 runas (usando `utf8.RuneCountInString`)

**Razon:** Validaciones tempranas fallan rapido con errores de dominio claros.

### 6. Errores de dominio scoped por archivo

**Decision:** Cada archivo define sus propios errores con prefijo descriptivo.

```go
// conversation_status.go
var ErrConversationStatusInvalid = errors.New("conversation status is invalid")

// message_status.go
var ErrMessageStatusInvalid = errors.New("message status is invalid")

// message_type.go
var ErrMessageTypeInvalid = errors.New("message type is invalid")

// conversation.go
var (
    ErrConversationInvalidID              = errors.New("conversation identifier is invalid")
    ErrConversationMissingContactAndAgent = errors.New("conversation must have a contact or an agent")
)

// message.go
var (
    ErrMessageInvalidID   = errors.New("message identifier is invalid")
    ErrMessageEmptyText   = errors.New("message text cannot be empty")
    ErrMessageTextTooLong = errors.New("message text exceeds maximum length")
)
```

**Razon:** Arquitectura que grita. Cada archivo es autonomo con sus errores. Facilita encontrar el origen del error.

### 7. Constantes sin valores magicos

**Decision:** Todos los valores literales como constantes con nombre.

```go
const (
    MinTextLength = 1
    MaxTextLength = 1000
)
```

**Razon:** Claridad, mantenibilidad, documentacion del dominio.

## Risks / Trade-offs

- **Trade-off:** Sin logica de negocio (asignacion, expiracion) - se implementara despues
- **Risco:** `Messages()` retorna slice desde `found` - si se modifica el slice, no afecta al map (seguro)
- **Trade-off:** Usamos `utf8.RuneCountInString` para longitud de texto - no cuenta grapheme clusters perfectamente pero es suficiente para el dominio
