# http-read-api Specification

## Purpose

Contrato REST de lectura para que un agente autenticado consulte la bandeja de conversaciones, abra una conversación con su historial y consulte contactos y usuarios por identificador.

## Requirements

### Requirement: Identidad del solicitante en lecturas

Las solicitudes a los endpoints de lectura SHALL obtener la identidad del usuario desde el encabezado `Authorization: Bearer <id>`. El sistema SHALL rechazar con 401 las solicitudes sin encabezado, con un valor no parseable o cuyo usuario no exista.

#### Scenario: Lectura autenticada

- **WHEN** llega una lectura con `Authorization: Bearer <uuid>` de un usuario existente
- **THEN** la operación se ejecuta con la identidad de ese usuario

#### Scenario: Falta el encabezado

- **WHEN** llega una lectura sin encabezado `Authorization`
- **THEN** el sistema responde 401

#### Scenario: Bearer no parseable o inexistente

- **WHEN** llega una lectura con un Bearer que no es un identificador válido o cuyo usuario no existe
- **THEN** el sistema responde 401

### Requirement: Listar la bandeja de conversaciones

El sistema SHALL exponer `GET /v1/conversations` para listar las conversaciones visibles para el agente autenticado. Una conversación SHALL ser visible cuando su estado es `pending` o cuando su agente asignado es el solicitante. El parámetro `status` SHALL ser obligatorio y SHALL aceptar únicamente `open` y `finished`: `open` SHALL incluir los estados `pending` y `assigned`, y `finished` SHALL incluir `resolved` y `expired`. Cuando `status` falte o no sea `open` ni `finished`, el sistema SHALL responder 422. El sistema SHALL aceptar además el parámetro opcional `external_contact_id`; cuando esté presente, SHALL resolver el contacto por su identificador externo y, si no existe, SHALL responder 404 sin consultar conversaciones; si existe, SHALL restringir el listado a las conversaciones de ese contacto. El sistema SHALL ordenar los resultados por el `sent_at` del último mensaje de la conversación de forma descendente y, en empate, por el identificador de la conversación de forma descendente, y SHALL devolver como máximo 20 conversaciones por página junto con los valores `next_before_sent_at` y `next_before_id` para pedir la página siguiente. Cada elemento SHALL exponer: el identificador de la conversación, su estado, el identificador interno y el identificador externo de su contacto, una vista previa del último mensaje (texto, `sent_at` y dueño) y la cantidad de mensajes del contacto sin leer.

#### Scenario: Conversaciones visibles

- **WHEN** el agente lista la bandeja
- **THEN** aparecen las conversaciones `pending` y las conversaciones cuyo agente asignado es el solicitante
- **AND** no aparecen las conversaciones asignadas a otro agente

#### Scenario: Filtro open

- **WHEN** el agente envía `status=open`
- **THEN** el sistema lista solo las conversaciones `pending` o `assigned` visibles para el agente

#### Scenario: Filtro finished

- **WHEN** el agente envía `status=finished`
- **THEN** el sistema lista solo las conversaciones `resolved` o `expired` del agente

#### Scenario: Status ausente o inválido

- **WHEN** el agente omite `status` o envía un valor fuera de `open` y `finished`
- **THEN** el sistema responde 422

#### Scenario: Filtro por contacto

- **WHEN** el agente envía un `external_contact_id` de un contacto existente
- **THEN** el sistema lista solo las conversaciones de ese contacto que sean visibles para el agente

#### Scenario: Contacto inexistente

- **WHEN** el agente envía un `external_contact_id` que no corresponde a ningún contacto
- **THEN** el sistema responde 404 sin consultar conversaciones

#### Scenario: Orden por último mensaje

- **WHEN** una conversación recibe un mensaje nuevo, sea del contacto o del agente
- **THEN** esa conversación pasa al inicio de la bandeja según el `sent_at` de ese mensaje

#### Scenario: Paginación de la bandeja

- **WHEN** el agente pide la página siguiente usando `before_sent_at` y `before_id` devueltos
- **THEN** el sistema devuelve las siguientes 20 conversaciones sin repetir las anteriores
- **AND** cuando no quedan más, los valores siguientes son nulos

#### Scenario: Vista previa y no leídos

- **WHEN** una conversación aparece en la bandeja
- **THEN** el elemento incluye el texto, `sent_at` y dueño de su último mensaje
- **AND** incluye el identificador interno y externo de su contacto
- **AND** incluye la cantidad de mensajes del contacto sin leer

### Requirement: Abrir una conversación con su historial

El sistema SHALL exponer `GET /v1/conversations/{id}` para obtener una conversación y sus mensajes más recientes. El sistema SHALL permitir el acceso cuando la conversación es `pending` o cuando su agente asignado es el solicitante; en caso contrario SHALL responder 403. Si la conversación no existe SHALL responder 404; si el `id` no es válido SHALL responder 400. La respuesta SHALL incluir el identificador de la conversación, su estado, el identificador interno y el identificador externo de su contacto, el agente asignado (si lo hay), la cantidad de mensajes del contacto sin leer y como máximo los últimos 20 mensajes. Los mensajes SHALL exponerse en orden `sent_at` ascendente y, en empate, por identificador ascendente. El sistema SHALL paginar hacia los mensajes más antiguos mediante los parámetros `before_sent_at` y `before_id`, y SHALL devolver `next_before_sent_at` y `next_before_id`, nulos cuando no quedan mensajes más antiguos. Los mensajes eliminados SHALL incluirse con estado `deleted` y su texto nulo.

#### Scenario: Abrir una conversación propia

- **WHEN** el agente abre una conversación asignada a él
- **THEN** recibe su información y sus últimos 20 mensajes en orden ascendente

#### Scenario: Abrir una conversación pendiente

- **WHEN** el agente abre una conversación `pending`
- **THEN** recibe su información y sus mensajes

#### Scenario: Conversación ajena

- **WHEN** el agente abre una conversación asignada a otro agente
- **THEN** el sistema responde 403

#### Scenario: Conversación inexistente o identificador inválido

- **WHEN** el agente abre una conversación que no existe o cuyo `id` no es válido
- **THEN** el sistema responde 404 o 400 respectivamente

#### Scenario: Paginar hacia mensajes antiguos

- **WHEN** el agente pide el historial anterior usando `before_sent_at` y `before_id` devueltos
- **THEN** el sistema devuelve los siguientes 20 mensajes más antiguos
- **AND** cuando no quedan, los valores siguientes son nulos

#### Scenario: Mensaje eliminado

- **WHEN** el historial incluye un mensaje eliminado
- **THEN** el mensaje aparece con estado `deleted` y su texto es nulo

### Requirement: Consultar un contacto por identificador

El sistema SHALL exponer `GET /v1/contacts/{id}` para que un agente autenticado obtenga un contacto por su identificador. La respuesta exitosa SHALL ser 200 con el identificador, el identificador externo y la fecha de creación. Si el `id` no es válido SHALL responder 400 y si el contacto no existe SHALL responder 404.

#### Scenario: Contacto existente

- **WHEN** el agente consulta un contacto existente
- **THEN** el sistema responde 200 con su identificador, identificador externo y fecha de creación

#### Scenario: Contacto inexistente o identificador inválido

- **WHEN** el agente consulta un contacto que no existe o con un `id` inválido
- **THEN** el sistema responde 404 o 400 respectivamente

### Requirement: Consultar un usuario por identificador

El sistema SHALL exponer `GET /v1/users/{id}` para que un agente autenticado obtenga un usuario por su identificador. La respuesta exitosa SHALL ser 200 con el identificador y la fecha de creación. Si el `id` no es válido SHALL responder 400 y si el usuario no existe SHALL responder 404.

#### Scenario: Usuario existente

- **WHEN** el agente consulta un usuario existente
- **THEN** el sistema responde 200 con su identificador y fecha de creación

#### Scenario: Usuario inexistente o identificador inválido

- **WHEN** el agente consulta un usuario que no existe o con un `id` inválido
- **THEN** el sistema responde 404 o 400 respectivamente

### Requirement: Paginación por posición

El sistema SHALL paginar las lecturas de conversaciones y de mensajes mediante los parámetros `before_sent_at` (instante RFC 3339) y `before_id` (identificador). Ambos SHALL ser opcionales y SHALL enviarse juntos; enviar solo uno SHALL rechazarse con 400, y un valor no parseable SHALL rechazarse con 400. El sistema SHALL exponer en cada respuesta paginada `next_before_sent_at` y `next_before_id`, nulos cuando no quedan elementos.

#### Scenario: Primera página sin posición

- **WHEN** una lectura paginada se solicita sin `before_sent_at` ni `before_id`
- **THEN** el sistema devuelve la primera página

#### Scenario: Posición inválida

- **WHEN** una lectura paginada recibe `before_sent_at` o `before_id` no parseables, o recibe solo uno de los dos
- **THEN** el sistema responde 400

### Requirement: Respuestas de error uniformes en lecturas

El sistema SHALL responder los errores de las lecturas con `Content-Type: application/problem+json` y un cuerpo con al menos `title`, `status` y `detail`, e `instance` con el identificador del request. Los errores SHALL mapearse a los códigos: 400 para entrada malformada o posición inválida, 401 para autenticación, 403 para falta de propiedad, 404 para recurso inexistente, 422 para validación, 500 para fallos internos y 504 para timeout del request.

#### Scenario: Formato de error

- **WHEN** una lectura falla
- **THEN** la respuesta usa `application/problem+json` con `title`, `status` y `detail`

#### Scenario: Falta de propiedad

- **WHEN** el agente solicita una conversación que no es `pending` ni suya
- **THEN** el sistema responde 403

#### Scenario: Timeout del request

- **WHEN** la lectura excede el tiempo máximo del request
- **THEN** el sistema responde 504
