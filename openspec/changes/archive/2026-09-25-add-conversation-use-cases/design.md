# Design

## Context

El módulo `conversations` hoy solo tiene dominio (aggregate `Conversation`, entidad `Message`, value objects). No existe capa de aplicación, ni los módulos `agents`/`contacts`. Ver `proposal.md`.

Restricciones que moldean el diseño:

- DDD: la intención de negocio vive en el dominio; los casos de uso orquestan y no reimplementan reglas.
- Bajo acoplamiento: los módulos se comunican por interfaces en la raíz (`ModuleNameAPI`), no por implementaciones.
- Go 1.27 con el paquete `uuid` de la stdlib (`uuid.UUID`, `uuid.NewV7`).
- La persistencia/infraestructura está fuera de este cambio; los finders se definen como interfaces y se implementan después.

## Goals / Non-Goals

**Goals:**

- Definir los 7 casos de uso de conversaciones con una convención uniforme.
- Definir las fronteras públicas `AgentsAPI` y `ContactsAPI` mínimas que consumen los casos de uso.
- Definir el puerto outbound `AgentMessageSender` y el flujo de envío.
- Ajustar el dominio para soportar rehidratación parcial y `externalID` en construcción.

**Non-Goals:**

- Persistencia SQLC/Postgres, handlers HTTP y adaptadores del canal externo.
- Casos de uso `AgentEditMessage`, `AgentDeleteMessage`, `MarkAgentMessageFailed` (métodos de dominio se conservan).
- Reportes asíncronos de estado de entrega de la plataforma.
- Validación de existencia de agente en los casos de contacto.

## Decisions

### 1. Estructura de módulos y frontera API

```
internal/modules/
  conversations/
    api.go                    # ConversationsAPI (frontera pública)
    app/                      # un archivo por caso de uso
    domain/
      conversation.go
      conversation_repository.go
      agent_message_sender.go # puerto outbound
  agents/
    api.go                    # AgentsAPI
    app/  domain/
  contacts/
    api.go                    # ContactsAPI
    app/  domain/
```

Cada módulo expone en su raíz una interfaz con los métodos que otros módulos necesitan, devolviendo valores primitivos. Alternativa descartada: importar entidades de otros módulos (acoplamiento alto).

### 2. Convención de casos de uso

- Interface con un solo método `Execute(ctx context.Context, input <Caso>Input) error`.
- Implementación privada (`agentSendMessage`), constructor `New<Caso>(...) <Caso>` que devuelve la interface.
- Un error base por archivo, en gerundio y en el orden del nombre del método (`SendAgentMessage` → `ErrSendingAgentMessage`).
- Método `joinErr(errs ...error) error` en el struct que une los errores de la orquestación con el error base: `errors.Join(ErrBase, errors.Join(errs...))`.
- Idempotencia inline en `Execute` con `errors.Is`: si el error de dominio corresponde a una operación ya satisfecha, retorna `nil`. Sin helper (se descartó `isIdempotent` por no aportar al negocio).

### 3. Finders dedicados y rehidratación

El repositorio expone búsquedas por caso (no un `FindByID` genérico):

| Finder | Casos de uso | Trae |
|---|---|---|
| `FindWithoutMessages` | `SendAgentMessage`, `ExpireConversation`, `ResolveConversation` | solo metadatos |
| `FindWithContactUnreadMessagesByID` | `AgentReadConversation` | metadatos + mensajes del contacto sin leer |
| `FindWithMessageByExternalID` | `ReceiveContactMessageEdit`, `ReceiveContactMessageDelete` | metadatos + el mensaje |
| `FindOpenByContactID` | `ReceiveContactMessage` | metadatos + mensajes del contacto (índice dedup) |

`FindOpenByContactID` asume **una sola conversación activa por contacto** (no recibe timestamp).

Para poder hidratar subconjuntos (incluido 0 mensajes) se separa la rehidratación de la creación:

- `NewConversation` / `NewMessage`: creación estricta con validaciones (la conversación exige ≥1 mensaje).
- `RehydrateConversation` / `RehydrateMessage`: rehidratación directa desde datos persistidos, **sin validar invariantes** y permitiendo 0..n mensajes. Delegan en constructores privados `newConversation` / `newMessage`, que inicializan los mapas si vienen nil.
- Alternativa descartada: cargar siempre el agregado completo (memoria innecesaria con conversaciones largas).

### 4. Contratos de módulos para el MVP

- `AgentsAPI.FindAgentByID(ctx, id) (uuid.UUID, error)` — usado por los 6 casos de agente.
- `ContactsAPI.GetOrCreateContactIDByExternalID(ctx, externalContactID) (uuid.UUID, error)`.
- `ContactsAPI.FindExternalContactIDByContactID(ctx, contactID) (string, error)`.

### 5. Puerto `AgentMessageSender` y flujo de envío

```
type AgentMessageSender interface {
    Send(ctx context.Context, to string, message *Message) (externalMessageID string, err error)
}
```

- El caso `SendAgentMessage` persiste primero (save 1), luego resuelve el destino vía `ContactsAPI`, envía por el puerto y asigna el `externalID` (save 2).
- Si el envío falla: `MarkAgentMessageFailed` + save y se retorna el error envuelto (la UI lo verá `failed`).
- No hay worker ni caso de uso de recuperación para el MVP.

### 6. Cambios de dominio

- `NewMessage` recibe `externalID *string`; valida que, si no es nil, no sea vacío. Los mensajes de agente pasan `nil` y el constructor nunca fabrica punteros a locales.
- `RehydrateConversation` y `RehydrateMessage`: rehidratación directa sin validación, apoyadas en los constructores privados nil-safe `newConversation` / `newMessage`.
- Nuevo `ErrConversationNotFound`.
- Nuevo método de agregado `AssignAgentMessageExternalID(msgID, externalMessageID, at)` (éxito del envío).

### 7. Idempotencia por caso

| Caso de uso | Error silenciado |
|---|---|
| `ReceiveContactMessage` | `ErrConversationDuplicateMessage` |
| `ReceiveContactMessageEdit` | — (la edición stale la absorbe el dominio) |
| `ReceiveContactMessageDelete` | `ErrMessageAlreadyDeleted` |
| `SendAgentMessage` | — |
| `AgentReadConversation` | — |
| `ExpireConversation` | `ErrConversationFinished` |
| `ResolveConversation` | `ErrConversationFinished` |

## Risks / Trade-offs

- **[Riesgo] Duplicado sin índice único** → Sin el índice único de `messages.external_id` (infra), un webhook repetido cuando no existe conversación activa podría crear una conversación nueva duplicada. Mitigación futura: el índice único en infraestructura.
- **[Riesgo] `failed` no bloquea `SendAgentMessage`** → El dominio no consulta `failed` al enviar; un reintento con el mismo `MessageID` podría resucitar el mensaje. Aceptado para MVP.
- **[Trade-off] Doble `Save` en `SendAgentMessage`** → Necesario para registrar localmente antes de enviar y actualizar el `externalID` después. Se acepta por consistencia.
- **[Trade-off] Rehidratación 0..n** → Habilita finders parciales pero relaja el invariante de construcción; se mantiene `NewConversation` estricto para creación.
- **[Trade-off] Sin validación de agente en receive** → Los casos de contacto solo validan contacto; el agente no aplica.
- **[Decisión] Mocks en producción** → Los dobles de prueba viven en archivos `*_mock.go` de los paquetes de producción (decisión aceptada). Efecto: `go test -cover` reporta esos paquetes por debajo de 100% porque el código de mock no se ejecuta; la cobertura del código real se mide excluyéndolos.
- **[Trade-off] Cobertura 98.5% en `conversations/app`** → Dos ramas defensivas son inalcanzables por contrato del dominio (`ExpireConversation` solo devuelve `ErrConversationFinished`; `MarkAgentMessageFailed` no puede fallar tras un `SendAgentMessage` exitoso). Se conservan como defensa y se aceptan sin cubrir.

## Migration Plan

No aplica: este cambio solo define la capa de aplicación y contratos; no toca base de datos ni despliegue.

## Open Questions

- Manejo de reportes asíncronos de entrega (`MarkAgentMessageFailed` como caso de uso) y del índice único de dedup: se difieren a un cambio posterior.
