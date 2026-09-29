# Proposal

## Why

Hoy un contacto solo se identifica por su `external_contact_id` (la identidad de un canal). No hay forma de saber que varios contactos de canales distintos son la misma persona, ni de guardar sus datos personales (documento, nombre, teléfono, email, dirección). Esto impide asociar varios tipos de contacto a una sola persona y mostrar un nombre legible en la bandeja.

## What Changes

- Se introduce una entidad de persona (`personal_information`) con `identification_number` único e inmutable, los campos `first_name`, `last_name`, `phone_number`, `email`, `address`, y `created_at`/`updated_at`.
- El `identification_number` acepta **solo letras y números** (se aplica `trim`; cualquier otro carácter se rechaza). No se asume el formato de ningún país.
- Solo `identification_number` es obligatorio (1..100 grafemas, solo letras y números). Los demás datos (`first_name`, `last_name`, `phone_number`, `email`, `address`) son opcionales y admiten nulo, porque no siempre se tiene toda la información a mano; cuando están, permiten hasta 100 grafemas (nombres y teléfono) o 254 (email y dirección). El límite lo aplica el dominio; la base no impone longitud.
- `contacts` gana `display_name` y una referencia nullable (`personal_information_id`) a la persona. Una persona puede estar referenciada por varios contactos (N contactos : 1 persona).
- El `display_name` del contacto lo provee el canal a través del webhook (`POST /v1/webhooks/1channel`), y SHALL setearse únicamente con `message.received`; los demás eventos no lo tocan.
- La persona es un agregado propio (compartido y con ciclo de vida independiente); el contacto la referencia, no la posee. Puede existir sin contactos (huérfana).
- Guardar desde la consola hace **upsert por `identification_number`**: inserta si no existe, pisa los datos si existe, y devuelve el identificador canónico de la fila. Luego asocia el contacto a esa persona. Son dos escrituras secuenciales, sin transacción cruzada entre agregados.
- No hay borrado ni desasociación: solo crear, modificar y asociar. El `identification_number` nunca cambia.
- Se agrega una lectura para buscar la persona por `identification_number` (solo precarga, no escribe).
- La bandeja y el detalle exponen una etiqueta del contacto resuelta como: nombre completo de la persona si existe, si no `display_name`, si no `external_contact_id`. La persona se expone como un objeto anidado `personal_information`.
- La consola deja de ser dos containers y pasa a tres (`lista | detalle | persona`): el tercer panel, a la derecha, precarga la persona si el contacto la tiene, permite buscarla por documento y guardarla con la misma validación de longitud. El panel es editable con la conversación en cualquier estado (incluidas `expired` y `resolved`).
- La consola es **desktop-first**: no se trabaja el comportamiento mobile en este cambio.
- Se actualiza el seed de desarrollo: contactos con y sin `display_name`, personas asociadas (incluidas compartidas), y acuses de lectura también en mensajes del agente cuando corresponda.

## Capabilities

### New Capabilities

- `personal-information`: agregado Persona compartible: campos, `identification_number` único e inmutable, upsert con pisa, orfandad permitida, y asociación N:1 con contactos (crear, modificar, asociar; sin borrar ni desasociar).

### Modified Capabilities

- `contacts`: el contacto gana `display_name` (seteado por el webhook en `message.received`) y una referencia nullable a la persona; se define la etiqueta de presentación (persona -> `display_name` -> identificador externo).
- `http-api`: nuevo endpoint de escritura autenticado para guardar (upsert) la persona y asociarla a un contacto; el webhook acepta `display_name` y lo aplica solo en `message.received`.
- `http-read-api`: nueva lectura de persona por `identification_number`; la bandeja y el detalle exponen la etiqueta del contacto y la persona asociada como objeto anidado.
- `persistence`: nueva tabla `personal_information` con `identification_number` único y `created_at`/`updated_at`; `contacts` gana `display_name` y `personal_information_id` con integridad referencial restrictiva; repositorio de persona.
- `agent-console`: tercer panel de la persona en el detalle de la conversación (precarga, búsqueda por documento, guardado), editable en cualquier estado; desktop-first.
- `webhook-simulator`: campo `display_name` opcional para `message.received`.

## Impact

- **Backend (módulo `contacts`)**: nuevo agregado y repositorio `PersonalInformation` (domain/app/infra/pg), extensión del agregado `Contact` (nuevo campo, nuevo método de dominio para asociar, rehidratación), DTOs de lectura y handlers de escritura, y consultas SQLC nuevas. El módulo pasa a tener dos agregados (Contact y PersonalInformation); sus consultas viven en `db/queries/contacts`.
- **Backend (módulos `conversations`)**: el webhook y `ReceiveContactMessage` propagan `display_name` a `contacts` a través de su API; las consultas de lista y detalle de conversación agregan `LEFT JOIN personal_information` (lectura infra-infra, como ya hacen con `contacts`).
- **Wiring**: repositorio, caso de uso y handlers nuevos registrados en `cmd/api/app.go` y el grupo autenticado de `router.go`.
- **Base de datos**: migración con la tabla `personal_information`, columnas nuevas en `contacts` y sus constraints/índices; se editan las migraciones existentes en su lugar (modo dev, `make reset`). Se actualiza `db/seed/dev_seed.sql`.
- **API `/v1`**: nuevos endpoints bajo `Auth`; se extienden las respuestas de lectura de conversaciones y contactos; el webhook acepta `display_name`.
- **Frontend**: layout de tres paneles, nuevo componente de persona, campo de `display_name` en el simulador y llamadas de API.
- **Specs**: se sincronizan los deltas de las capacidades listadas.
