# Design

## Context

Ver `proposal.md`. El módulo `contacts` hoy tiene un solo agregado (`Contact`: `id uuid`, `external_contact_id` único, `created_at`) y su repositorio (`FindByID`, `FindByExternalContactID`, `Save` con upsert por `id`). La bandeja y el detalle de conversaciones leen contactos por `JOIN` directo desde SQLC. El patrón de escritura es: `handler -> app (caso de uso) -> domain -> repo pg`; las lecturas son `handler -> *sqlc.Queries -> DTO`, sin interfaces ni mocks. `domain`/`app` nunca importan `infra`. Las migraciones se editan en su lugar en dev y la base se recrea (`make reset`).

## Goals / Non-Goals

**Goals:**

- Introducir la persona como agregado compartible y asociarla a los contactos, con la etiqueta de presentación resuelta.
- Mantener el flujo de escritura `app/domain -> repo` y el de lectura `handler -> sqlc`.
- Resolver el guardado como upsert por documento sin transacción cruzada entre agregados.

**Non-Goals:**

- Borrar o desasociar personas.
- Editar el `identification_number` una vez creado.
- Comportamiento mobile (la consola es desktop-first).
- Endpoint o UI dedicado a setear `display_name` (lo setea el webhook).
- Cambiar la creación/resolución de contactos del webhook más allá de propagar `display_name`.

## Decisions

### 1. La persona es un agregado propio; el contacto la referencia

`personal_information` tiene identidad y ciclo de vida propios porque puede existir sin contactos (orfandad) y puede estar referenciada por varios contactos (propiedad compartida). Una entidad hija pertenece a un único agregado, así que no puede ser compartida. `contacts.personal_information_id` nullable apunta a la persona.

- Alternativa descartada: PI como entidad hija de `Contact`. Rompería el compartir (la misma persona por varios canales) o exigiría duplicar personas, y con `identification_number` único no se podrían tener dos contactos de la misma persona.

### 2. Identificador sustituto `uuid` + documento único, igual que `Contact`

`personal_information.id` es `uuid` generado por la aplicación; `identification_number` es `text` único e inmutable. Es exactamente la forma de `contacts` (`id uuid` + `external_contact_id` único): una entidad resuelta por una clave natural que además lleva identidad sustituta. Mantiene la uniformidad del dominio (todo agregado tiene `ID() uuid`, `FindByID`, `Rehydrate`).

- Alternativa descartada: `identification_number` como PK. Acopla la identidad al dato de negocio y rompe la uniformidad de repos/API.

### 3. Upsert por documento con pisado; el documento no cambia

Guardar persona ejecuta `INSERT ... ON CONFLICT (identification_number) DO UPDATE` sobre los campos de datos (nunca sobre `identification_number`) y devuelve el id de la fila resultante. Si el documento no existe, inserta con el `id` propuesto; si existe, reemplaza los datos (un dato ausente o vacío queda nulo) y conserva el `id` original y el `created_at`.

- Consecuencia con identificadores internos distintos en una carrera: el `id` que se conserva es el del primero y los datos quedan los del último guardado (last-write-wins). Es el comportamiento de "pisar" acordado.
- `DO UPDATE` (y no `DO NOTHING`) es necesario para que `RETURNING` devuelva la fila existente en el conflicto.

### 4. `Save` del repositorio de personas es `error` y reconcilia el id con `AssignID`

`PersonalInformationRepository.Save(ctx, *PersonalInformation) error` sigue el estilo de la casa (solo `error`). El repo hace el upsert `RETURNING id` y, si el id persistido difiere del propuesto (conflicto por documento), reconcilia la entidad con `PersonalInformation.AssignID(id)`. Así el caso de uso obtiene el id canónico sin que `Save` tenga que devolver la entidad.

### 5. Un único caso de uso con nombre de negocio y dos escrituras secuenciales

El guardado es **un solo caso de uso** de aplicación (módulo `contacts`, `app`) que expresa la acción de negocio "registrar la información personal de un contacto":

- `RegisterContactPersonalInformation`: busca la persona por `identification_number`; si existe, actualiza sus datos y la guarda; si no, la crea y la guarda; luego asocia el contacto a esa persona.

```
pi, err := registerContactPersonalInformation.Execute(contactID, datos) // 1 caso de uso, 2 escrituras secuenciales (commit propio cada una)
```

No se abre una transacción que abarque ambos agregados (eso requeriría pasar `*sqlc.Queries`/tx a los repositorios: complejidad no justificada). El estado intermedio es una persona huérfana, que es un estado válido, y la asociación es idempotente/reintentable; por eso la consistencia eventual es benigna. Esto respeta la regla "una transacción por agregado".

- Alternativa descartada: separar en dos casos de uso (`RegisterPersonalInformation` + `AssignPersonalInformationToContact`). La asociación no es una acción de negocio independiente del guardado, así que partirlo no aportaba nada.
- Alternativa descartada: UnitOfWork o inyección de `*sqlc.Queries` en los repos para atomicidad. Sobreeningeniería para un estado intermedio legal.

### 6. Capas

- Escritura: `infra/http (handler) -> app (caso de uso) -> domain (PersonalInformation, Contact) -> domain repo port -> infra/pg`. `domain`/`app` no importan `infra`.
- Lectura: `infra/http (read handler) -> *sqlc.Queries -> DTO`. Sin app/domain, sin interfaces ni mocks. Incluye el `LEFT JOIN` a `personal_information` en la lista y el detalle, y la búsqueda por documento.

### 7. Contrato HTTP

- `PUT /v1/contacts/{id}/personal-information` (escritura): upsert por documento + asociación. Cuerpo con `identification_number` (obligatorio) y los datos opcionales. Responde 200 con la persona.
- `GET /v1/personal-information/{identification_number}` (lectura, único handler de PI): búsqueda por documento; 400 si el documento es vacío/inválido, 200 si existe, 404 si no existe.
- Ambos dentro del grupo autenticado (`Auth`) de `router.go`. La persona no tiene rutas propias de escritura: todo entra por contacto.

### 8. Etiqueta de presentación

Se proyectan `personal_information.first_name/last_name` y `contacts.display_name` en las consultas de lectura y la etiqueta se resuelve en el mapeo `row -> DTO` (código Go): persona -> `display_name` -> `external_contact_id`. Se evita armar la concatenación en SQL para poder testear el mapeo.

### 9. Migraciones

En dev se editan en su lugar: se agrega `display_name` y `personal_information_id` a `contacts` y se crea `personal_information` (con `id uuid PK`, `identification_number text NOT NULL UNIQUE`, datos opcionales que admiten nulo, `created_at timestamptz NOT NULL` y `updated_at timestamptz NOT NULL`), con la FK `contacts.personal_information_id -> personal_information.id` restrictiva (sin cascade). Se recrea con `make reset` y se regenera con `sqlc generate`.

### 10. Validación de los campos: solo el documento es obligatorio

Solo `identification_number` es obligatorio: no vacío, compuesto únicamente por letras y números (no se asume el formato de ningún país) y entre 1 y 100 grafemas. Los otros cinco campos son opcionales y admiten nulo, porque no siempre se tiene toda la información a mano; al normalizar, un valor ausente o vacío se guarda como nulo. Cuando están presentes, `first_name`, `last_name` y `phone_number` admiten hasta 100 grafemas, y `email` y `address` hasta 254.

La longitud se mide en **grafemas**, reusando `uniseg.GraphemeClusterCount` como ya hace el texto de mensajes (`message.go:74`), para que acentos y emojis cuenten como una unidad. El límite vive en el **dominio** (error de validación → 422), y la base guarda `text` sin límite de longitud y admite nulos. El front refleja la validación, pero la fuente de verdad es el servidor.

### 11. Disposición de la consola

El shell pasa de dos a tres paneles: `lista | detalle | persona`. El panel de persona se monta junto al detalle cuando hay una conversación seleccionada y se oculta sin selección. La consola es desktop-first; no se trabaja el comportamiento mobile en este cambio.

### 12. `display_name` provisto por el webhook

El canal envía un `display_name` opcional en `POST /v1/webhooks/1channel`, y el servidor lo propaga al contacto **solo** con `message.received`; el resto de eventos lo ignora. La firma de `ContactsAPI.GetOrCreateContactIDByExternalID` gana un `displayName *string`: `ReceiveContactMessage` lo pasa, y `edit`/`delete`/`read` pasan `nil`. El caso de uso de contactos setea el valor al crear o actualizar; si viene ausente o vacío, conserva el existente.

### 13. Timestamps de la persona

`personal_information` lleva `created_at` y `updated_at`. La aplicación provee los instantes (como en `Contact`); el upsert conserva `created_at` y actualiza `updated_at` en cada guardado.

### 14. Alfabeto del documento

`identification_number` admite solo letras y números tras `trim`; cualquier otro carácter es error de validación. No se normaliza ni se asume formato por país.

### 15. Ubicación de módulo y consultas

El módulo `contacts` pasa a tener dos agregados (Contact y PersonalInformation); es válido, un módulo puede albergar más de un agregado. Las queries propias de la persona viven en `db/queries/contacts`. Las consultas de lista y detalle de conversación permanecen en `db/queries/conversations` y agregan `LEFT JOIN personal_information` (lectura infra-infra, igual que ya joinean `contacts`). No se expone una API cross-módulo para la persona: la etiqueta se resuelve en las lecturas.

### 16. Forma anidada en las respuestas

La persona se expone como objeto anidado `personal_information` (nulo cuando el contacto no tiene persona) en el detalle de conversación y en la consulta de contacto, con sus datos y fechas. Al abrir una conversación, el detalle trae la persona completa (una sola llamada); la búsqueda por documento es el único endpoint de lectura propio de la persona.

### 17. Panel editable en cualquier estado

El panel de persona es independiente del ciclo de vida de la conversación: se puede precargar, buscar y guardar aunque la conversación esté `expired` o `resolved`.

### 18. Seed de desarrollo

Se actualiza `db/seed/dev_seed.sql`: truncar también `personal_information`; contactos con `display_name` (la minoría sin ninguno) y otros con persona asociada (incluida compartida entre contactos); y acuses de lectura en mensajes del agente cuando exista un mensaje posterior del contacto.

### 19. Wiring

El repositorio, el caso de uso y los handlers nuevos se cablean en `cmd/api/app.go` (`dependencies` y `newContactsModule`) y se registran en el grupo `Auth` de `router.go`.

## Risks / Trade-offs

- [Pisado concurrente de datos de la persona] → Es el upsert acordado (last-write-wins); la consola busca antes de guardar.
- [Carrera de creación con `AssignID`] → Con `Save` error-only y `AssignID`, en una carrera de dos creaciones del mismo documento el `created_at` en memoria puede quedar con el valor propuesto en vez del original; es raro y se autocompone al reintentar.
- [Reasociar por error] → El documento del contacto es editable y guardar reasocia a la persona de ese documento; es intencional para corregir un dato mal cargado. La persona anterior no se borra (puede quedar huérfana).
- [Fallo entre las dos escrituras deja persona huérfana] → Estado válido y reintentable; no hay corrupción ni duplicados.
- [`display_name` solo lo setea el webhook] → Un contacto sin persona y sin `display_name` (nunca llegó por un canal con nombre) muestra su `external_contact_id`; es el fallback previsto.
- [Cambiar la firma de `GetOrCreateContactIDByExternalID`] → Toca los cuatro casos de uso de recepción y sus mocks/tests; se pasan `nil` donde no aplica y se cubre con los tests existentes.
- [Documento con alfabeto restringido] → Solo letras y números; documentos con guiones, puntos o espacios serán rechazados por validación y habrá que acordar su formato si aparece un caso real.
- [Cambios de esquema con migraciones editadas en su lugar] → Solo dev; la base se recrea con `make reset`.
- [Exposición de la persona en el detalle de conversación] → La persona es global, no propiedad del agente; el detalle la expone a cualquier agente que pueda ver la conversación, igual que el contacto.

## Migration Plan

1. Editar las migraciones en su lugar (columnas de `contacts` + tabla `personal_information`) en modo dev.
2. `make reset` (o `migrate down`/`up`) y `sqlc generate`.
3. Desplegar backend y frontend juntos; no hay datos en producción.
4. Rollback: revertir migraciones y `make reset`; revertir el código.
