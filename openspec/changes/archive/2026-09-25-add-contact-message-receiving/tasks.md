# Tasks

## 1. Message: identificador externo

- [x] 1.1 Agregar campo `externalID *string` al struct Message y verificar que compila
- [x] 1.2 Agregar errores `ErrMessageExternalIDInvalid` y `ErrMessageExternalIDAlreadySet` con mensajes de negocio
- [x] 1.3 Agregar getter `ExternalID() *string` y verificar que compila
- [x] 1.4 Agregar metodo `AssignExternalID(id string) error` con validacion set-once y verificar que compila
- [x] 1.5 Agregar tests para `AssignExternalID`: exito, ya asignado, valor vacio (AAA pattern)

## 2. Conversation: invariante de mensajes minimos

- [x] 2.1 Agregar error `ErrConversationEmptyMessages` con mensaje de negocio
- [x] 2.2 Modificar `NewConversation` para validar `len(messages) >= 1` y verificar que compila
- [x] 2.3 Actualizar tests existentes de `TestNewConversation` para incluir mensajes validos y agregar casos: mensajes vacios, mensajes nil

## 3. Conversation: indice de mensajes externos

- [x] 3.1 Agregar campo `externalMsgIdx map[string]uuid.UUID` al struct Conversation
- [x] 3.2 Modificar `NewConversation` para construir el indice desde mensajes con externalID
- [x] 3.3 Verificar busqueda por externalID directo en el map (sin helper)
- [x] 3.4 Verificar que el indice se construye correctamente en tests de NewConversation

## 4. Conversation: recibir mensaje de contacto

- [x] 4.1 Agregar errores `ErrConversationHasNoContact`, `ErrConversationContactNotOwner`, `ErrConversationDuplicateMessage` con mensajes de negocio
- [x] 4.2 Agregar metodo `ReceiveContactMessage(contactID uuid.UUID, msg Message, at time.Time) error` con validaciones de pertenencia y duplicados
- [x] 4.3 Agregar tests para `ReceiveContactMessage`: exito, contacto no pertenece, sin contacto, duplicado, sin externalID, actualiza timestamp (AAA pattern)

## 5. Quality gates

- [x] 5.1 Ejecutar `go mod verify` y verificar que pasa
- [x] 5.2 Ejecutar `gofmt -l .` y verificar que no hay archivos sin formatear
- [x] 5.3 Ejecutar `go vet ./...` y verificar que pasa
- [x] 5.4 Ejecutar `go build ./...` y verificar que compila
- [x] 5.5 Ejecutar `go test -race -count=1 ./...` y verificar que todos los tests pasan
