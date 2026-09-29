# Tasks

## 1. Esquema, SQLC y seed

- [x] 1.1 Editar las migraciones en su lugar: crear `personal_information` (`id uuid PK`, `identification_number text NOT NULL UNIQUE`, `first_name`/`last_name`/`phone_number`/`email`/`address` que admiten nulo, `created_at timestamptz NOT NULL`, `updated_at timestamptz NOT NULL`) y agregar `display_name text` y `personal_information_id uuid` nullable con FK restrictiva a `contacts`; verificar con `make reset` (o `migrate down`/`up`) que la base queda con la tabla, columnas y constraints esperados.
- [x] 1.2 Agregar las queries de personas en `db/queries/contacts` (`FindPersonalInformationByID`, `FindPersonalInformationByIdentificationNumber`, `UpsertPersonalInformation` con `ON CONFLICT (identification_number) DO UPDATE` —pisando datos y `updated_at`, conservando `created_at`— y `RETURNING`), extender `UpsertContact`/`FindContactByID` con `display_name` y `personal_information_id`, y agregar `LEFT JOIN personal_information` en `ListConversationsForAgent` y `FindConversationWithContactByID`; verificar que `sqlc generate` corre sin errores y `go -C backend build ./...` compila.
- [x] 1.3 Actualizar `db/seed/dev_seed.sql`: truncar también `personal_information`; contactos con `display_name` (la minoría sin ninguno) y otros con persona asociada (incluida compartida); read receipts en mensajes del agente cuando exista un mensaje posterior del contacto, manteniendo consistentes `status`, `unread_count`, `last_message_at` y `last_message_id`; verificar con `make reset && make seed` que corre y que los invariantes se cumplen (conteos y no leídos).

## 2. Dominio

- [x] 2.1 Crear el agregado `PersonalInformation` (solo `identification_number` obligatorio: 1..100 grafemas y solo letras/números; datos opcionales `first_name`/`last_name`/`phone_number`/`email`/`address` que admiten nulo con máximos 100/100/100/254/254; `Rehydrate`, getters, `UpdatePersonalData`, `AssignID`, `created_at`/`updated_at` y errores de validación) y su puerto `PersonalInformationRepository` (`FindByID`, `FindByIdentificationNumber`, `Save` error que reconcilia el id); verificar con tests unitarios table-driven de creación/rehidratación, documento ausente/inválido/fuera de rango, dato opcional ausente, exceso de máximo y conteo por grafemas (acentos/emoji).
- [x] 2.2 Extender el agregado `Contact` con `display_name` (con método para setearlo/actualizarlo desde el canal) y la referencia a persona (`AssociatePersonalInformation`, getters, `Rehydrate` con los nuevos campos), y agregar la resolución pura de la etiqueta (persona -> `display_name` -> `external_contact_id`); verificar con tests unitarios de asociación, rehidratación, asignación de `display_name` (`AssignDisplayName`) y los tres casos de etiqueta.

## 3. Aplicación

- [x] 3.1 Extender `GetOrCreateContactByExternalID` con un `displayName *string` (cuando viene no vacío, setea/actualiza el `display_name` del contacto) y actualizar `ContactsAPI.GetOrCreateContactIDByExternalID` con el nuevo parámetro; actualizar `ReceiveContactMessage` para propagarlo y `ReceiveContactMessageEdit`/`Delete`/`Read` para pasar `nil`, con sus mocks y tests; verificar con tests unitarios de los cuatro casos de uso y de la API.
- [x] 3.2 Crear el caso de uso `RegisterContactPersonalInformation` (busca por documento: si existe actualiza y guarda, si no crea y guarda; luego asocia el contacto) en dos escrituras secuenciales sin transacción cruzada; verificar con tests unitarios usando `MockPersonalInformationRepository` y `MockContactRepository` (persona nueva, persona existente que pisa, contacto inexistente, error de guardado y el orden de llamadas).
- [x] 3.3 Generar/actualizar los mocks de los puertos (`PersonalInformationRepository`, `ContactRepository`) y verificar que los tests del módulo corren con `go -C backend test ./internal/modules/contacts/...`.

## 4. Persistencia

- [x] 4.1 Implementar el repositorio pg de personas (`FindByID`, `FindByIdentificationNumber` con ausencia `(nil, nil)`, `Save` error con upsert por documento que conserva `created_at`, actualiza `updated_at` y reconcilia el id vía `AssignID`) y actualizar el mapeo del repositorio de contactos con los nuevos campos; verificar con `go -C backend build ./...` y `go -C backend vet ./...` (sin tests de base).

## 5. HTTP de escritura y wiring

- [x] 5.1 Implementar `PUT /v1/contacts/{id}/personal-information` (upsert de la persona por documento + asociación, body con `identification_number` y datos opcionales, 200 con la persona) con request/response DTO, parseo de ruta/cuerpo y mapeo de errores de dominio a 400/404/422/401; verificar con tests unitarios de handler con el caso de uso mockeado para alta, pisado, solo documento, documento inválido (422), dato opcional demasiado largo (422), cuerpo inválido (400), contacto inexistente (404) e id de ruta inválido (400).
- [x] 5.2 Agregar `display_name` al `ContactWebhookRequest` y propagarlo solo en el evento `message.received` (los demás eventos lo ignoran); verificar con tests unitarios del handler de webhook para recepción con y sin `display_name` y para otro evento con `display_name`.
- [x] 5.3 Cablear repositorio, casos de uso y handlers nuevos en `cmd/api/app.go` (`dependencies`, `newContactsModule`) y registrar los handlers en el grupo `Auth` de `router.go`; verificar que `go -C backend build ./...` compila y que las rutas quedan detrás de `Auth` (revisión del router).

## 6. HTTP de lectura

- [x] 6.1 Implementar `GET /v1/personal-information/{identification_number}` como read handler sobre `*sqlc.Queries` (400 si el documento es vacío/inválido, 200 con la persona y sus fechas, 404 si no existe); verificar con tests unitarios del mapeo a DTO.
- [x] 6.2 Extender los read handlers de conversaciones (lista y detalle) y de contacto para exponer la etiqueta de presentación y la persona asociada como objeto anidado `personal_information` (nulo si no hay), incluyendo el helper puro de etiqueta; verificar con tests unitarios del mapeo `row -> DTO` para los casos con persona, con `display_name` y sin ninguno.

## 7. Frontend

- [x] 7.1 Agregar los tipos y las funciones de API para buscar la persona por documento (`GET /v1/personal-information/{documento}`) y guardar en un solo `PUT /v1/contacts/{id}/personal-information` (body con documento + datos); verificar con `pnpm --dir frontend type-check`.
- [x] 7.2 Crear el componente del panel de persona (precarga si existe, búsqueda por documento, campo editable para permitir reasociar, campos vacíos + guardar, sin borrar/desasociar, validación de longitud por campo con 100/254, editable en cualquier estado de la conversación) e integrarlo como tercer panel `lista | detalle | persona` en `App.vue` (desktop-first), usando la etiqueta resuelta en lista y encabezado; verificar con `pnpm --dir frontend build` y una prueba manual con `make seed` y `make run`.
- [x] 7.3 Mostrar los estados de carga y error del panel con el título y detalle del `application/problem+json`; verificar manualmente forzando un error de búsqueda y de guardado.
- [x] 7.4 Agregar el campo opcional `display_name` al simulador de webhook, visible y enviado solo para `message.received`; verificar con `pnpm --dir frontend build` y una prueba manual que setea el nombre del contacto.

## 8. Verificación de integración

- [x] 8.1 Correr las quality gates completas (`gofmt -l backend`, `go -C backend mod verify`, `go -C backend vet ./...`, `go -C backend build ./...`, `go -C backend test -race -count=1 ./...`, `pnpm --dir frontend type-check`, `pnpm --dir frontend build`) y verificar que pasan.
- [x] 8.2 Con `make reset` y `make seed`, verificar manualmente el flujo extremo a extremo: abrir una conversación sin persona, buscar un documento inexistente, guardar, comprobar la etiqueta en lista y detalle; repetir el guardado con el mismo documento desde otro contacto y comprobar que comparten la persona y que los datos se pisan; enviar `message.received` con `display_name` desde el simulador y comprobar que el nombre del contacto se actualiza.
