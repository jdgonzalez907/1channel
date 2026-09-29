# agent-console Specification

## Purpose

Pantalla web del agente que lista las conversaciones visibles para él y opera el detalle de una conversación (responder, marcar como leída y resolver), consumiendo el contrato REST `/v1` con identidad `Authorization: Bearer <id>`.

## Requirements

### Requirement: Identidad del agente por texto

La consola SHALL obtener el identificador del agente desde un campo de texto, SHALL persistirlo en el almacenamiento del navegador y SHALL enviarlo como `Authorization: Bearer <id>` en cada solicitud. Al entrar un identificador, la consola SHALL validarlo contra `GET /v1/users/{id}` antes de operar la bandeja. Cuando el identificador falte, no sea un UUID válido o el usuario no exista, la consola SHALL mostrar el error y SHALL NOT listar conversaciones.

#### Scenario: Identificador válido

- **WHEN** el agente escribe un identificador de un usuario existente y lo confirma
- **THEN** la consola valida el usuario, conserva el identificador y habilita la bandeja

#### Scenario: Identificador inválido o inexistente

- **WHEN** el agente escribe un identificador que no es un UUID o que no corresponde a ningún usuario
- **THEN** la consola muestra el error de validación y no lista conversaciones

#### Scenario: Restaurar sesión al recargar

- **WHEN** la página se recarga y existe un identificador guardado
- **THEN** la consola lo restaura desde el almacenamiento del navegador y lo usa como token

### Requirement: Listar conversaciones por estado

La consola SHALL listar las conversaciones mediante `GET /v1/conversations` con el parámetro `status` obligatorio, ofreciendo una selección entre `open` y `finished`. La opción por defecto SHALL ser `open`. Al cambiar el estado, la consola SHALL repetir la consulta para ese estado.

#### Scenario: Listado por defecto

- **WHEN** el agente entra a la bandeja con un identificador válido
- **THEN** la consola lista las conversaciones con `status=open`

#### Scenario: Cambio de estado

- **WHEN** el agente selecciona `finished`
- **THEN** la consola vuelve a consultar con `status=finished` y muestra ese listado

### Requirement: Filtro por identificador externo de contacto

La consola SHALL ofrecer un campo opcional para el `external_contact_id` con una acción de búsqueda y una de limpieza. La búsqueda SHALL consultar `GET /v1/conversations` incluyendo el `external_contact_id` y el estado vigente. Cuando el contacto no exista, la consola SHALL mostrar el error y SHALL NOT dejar el listado anterior como si fuera el resultado. La limpieza SHALL vaciar el campo y repetir la consulta sin el filtro de contacto.

#### Scenario: Buscar un contacto existente

- **WHEN** el agente escribe el `external_contact_id` de un contacto existente y busca
- **THEN** la consola muestra solo las conversaciones de ese contacto para el estado vigente

#### Scenario: Buscar un contacto inexistente

- **WHEN** el agente busca un `external_contact_id` que no existe
- **THEN** la consola muestra el error del servidor y no presenta un listado filtrado

#### Scenario: Limpiar el filtro

- **WHEN** el agente limpia el campo de contacto
- **THEN** la consola repite la consulta sin `external_contact_id` y muestra el listado completo del estado vigente

### Requirement: Contenido de cada item de la lista

Cada item de la lista SHALL mostrar en negrita el `external_contact_id`, la vista previa del último mensaje junto con su dueño (`agent` o `contact`), el estado de la conversación y la fecha del último mensaje. La cantidad de mensajes sin leer SHALL mostrarse únicamente cuando sea mayor a cero. La cantidad SHALL mostrarse de forma atenuada en una conversación `expired` y destacada en el resto.

#### Scenario: Item con mensajes sin leer

- **WHEN** una conversación tiene `unread_count` mayor a cero
- **THEN** su item muestra la cantidad de mensajes sin leer

#### Scenario: Item sin mensajes sin leer

- **WHEN** una conversación tiene `unread_count` igual a cero
- **THEN** su item no muestra la cantidad de mensajes sin leer

#### Scenario: Item de conversación expirada

- **WHEN** una conversación `expired` tiene `unread_count` mayor a cero
- **THEN** su item muestra la cantidad de mensajes sin leer de forma atenuada

#### Scenario: Item de conversación no expirada

- **WHEN** una conversación no `expired` tiene `unread_count` mayor a cero
- **THEN** su item muestra la cantidad de mensajes sin leer de forma destacada

#### Scenario: Vista previa y dueño

- **WHEN** una conversación aparece en la lista
- **THEN** su item muestra el texto y el dueño del último mensaje, además del estado y la fecha

### Requirement: Paginación de la lista por cursor

La consola SHALL paginar la lista usando el par `before_sent_at` y `before_id` devuelto como `next_before_sent_at` y `next_before_id`. La consola SHALL ofrecer una acción de cargar más que anexe la página siguiente al listado existente. Cuando no queden más páginas, la acción SHALL quedar deshabilitada.

#### Scenario: Cargar la página siguiente

- **WHEN** el agente usa la acción de cargar más y la última respuesta tiene `next_before_sent_at` y `next_before_id`
- **THEN** la consola consulta con ese par de posición y anexa los nuevos items al final del listado

#### Scenario: No hay más páginas

- **WHEN** la última respuesta tiene `next_before_sent_at` y `next_before_id` nulos
- **THEN** la acción de cargar más no está disponible

### Requirement: Abrir el detalle de una conversación

Al seleccionar un item, la consola SHALL solicitar `GET /v1/conversations/{id}` y SHALL mostrar en el encabezado del detalle el `external_contact_id` y el estado de la conversación. Cuando la conversación no exista o no sea accesible, la consola SHALL mostrar el error correspondiente sin dejar el detalle anterior visible.

#### Scenario: Selección de una conversación

- **WHEN** el agente selecciona una conversación de la lista
- **THEN** la consola carga su detalle y muestra en el encabezado el identificador externo del contacto y el estado

#### Scenario: Conversación no accesible

- **WHEN** el detalle responde 403 o 404
- **THEN** la consola muestra el error y no presenta un detalle desactualizado

### Requirement: Historial de mensajes del detalle

El detalle SHALL mostrar los mensajes de la conversación con su dueño, su texto y su fecha de envío. Los mensajes con estado `deleted` SHALL mostrarse como mensajes eliminados, sin texto. La consola SHALL ofrecer una acción de cargar más para traer los mensajes más antiguos usando `before_sent_at` y `before_id`, deshabilitada cuando no queden más.

#### Scenario: Mensajes con dueño

- **WHEN** el historial contiene mensajes del agente y del contacto
- **THEN** cada mensaje se muestra indicando si lo envió el agente o el contacto, junto con su texto y su fecha

#### Scenario: Mensaje eliminado

- **WHEN** el historial incluye un mensaje con estado `deleted` y texto nulo
- **THEN** la consola lo muestra identificado como eliminado y sin texto

#### Scenario: Cargar mensajes más antiguos

- **WHEN** el agente usa la acción de cargar más del historial y hay página siguiente
- **THEN** la consola anexa los mensajes más antiguos y mantiene el orden cronológico

### Requirement: Responder solo en conversaciones abiertas

La consola SHALL mostrar el campo de respuesta únicamente cuando la conversación esté abierta (`pending` o `assigned`). Al enviar `POST /v1/conversations/{id}/messages` con el texto, la consola SHALL refrescar el detalle y la lista. Cuando la conversación esté finalizada, la consola SHALL NOT mostrar el campo.

#### Scenario: Campo visible en conversación abierta

- **WHEN** el agente abre una conversación `pending` o `assigned`
- **THEN** el detalle muestra el campo para escribir y enviar una respuesta

#### Scenario: Campo oculto en conversación finalizada

- **WHEN** el agente abre una conversación `resolved` o `expired`
- **THEN** el detalle no muestra el campo de respuesta

#### Scenario: Envío de respuesta

- **WHEN** el agente envía un texto no vacío en una conversación abierta
- **THEN** la consola crea el mensaje, refresca el historial y la lista, y refleja el nuevo estado si la conversación `pending` pasó a `assigned`

#### Scenario: Error al responder

- **WHEN** el envío responde 409, 422 u otro error
- **THEN** la consola muestra el error y conserva el detalle visible

### Requirement: Resolver solo conversaciones asignadas

La consola SHALL mostrar la acción de resolver únicamente cuando la conversación tenga estado `assigned`. Al resolver con `PATCH /v1/conversations/{id}` con `{"status":"resolved"}`, la consola SHALL refrescar el detalle y la lista.

#### Scenario: Resolver disponible en asignada

- **WHEN** el agente abre una conversación en estado `assigned`
- **THEN** el detalle muestra la acción de resolver

#### Scenario: Resolver no disponible en pendiente ni finalizada

- **WHEN** el agente abre una conversación `pending`, `resolved` o `expired`
- **THEN** el detalle no muestra la acción de resolver

#### Scenario: Resolución exitosa

- **WHEN** el agente resuelve una conversación `assigned` y el servidor responde 204
- **THEN** la consola refresca el detalle y la lista mostrando el nuevo estado

### Requirement: Marcar como leída al abrir una conversación asignada al agente

Al abrir una conversación asignada al agente solicitante (`agent_id` igual al identificador del agente) con `unread_count` mayor a cero, la consola SHALL marcar los mensajes del contacto como leídos con `PATCH /v1/conversations/{id}/messages` y SHALL refrescar la lista para actualizar el contador, sin importar el estado de la conversación (incluye `assigned`, `resolved` y `expired` asignadas). Cuando la conversación no esté asignada al agente, la consola SHALL NOT intentar marcarla como leída.

#### Scenario: Marcar leída al abrir una asignada

- **WHEN** el agente abre una conversación `assigned` con `unread_count` mayor a cero
- **THEN** la consola marca los mensajes como leídos y actualiza el contador en la lista

#### Scenario: Marcar leída en conversación finalizada asignada

- **WHEN** el agente abre una conversación `resolved` o `expired` asignada a él con `unread_count` mayor a cero
- **THEN** la consola marca los mensajes como leídos y actualiza el contador en la lista

#### Scenario: No marcar en conversación sin agente

- **WHEN** el agente abre una conversación `pending` o `expired` sin agente asignado
- **THEN** la consola no intenta marcar mensajes como leídos

#### Scenario: Sin mensajes sin leer

- **WHEN** el agente abre una conversación asignada a él con `unread_count` igual a cero
- **THEN** la consola no realiza el marcado

### Requirement: Estados vacíos, de carga y de error

La consola SHALL mostrar un estado de carga mientras espera respuestas, un mensaje cuando no hay conversaciones ni filtros aplicados, y el error del servidor (título y detalle de `application/problem+json`) cuando una operación falla.

#### Scenario: Listado vacío

- **WHEN** la consulta no devuelve conversaciones
- **THEN** la consola muestra un mensaje de lista vacía

#### Scenario: Error del servidor

- **WHEN** una consulta responde con un error
- **THEN** la consola muestra el título y el detalle del error

#### Scenario: Carga en curso

- **WHEN** una consulta está en curso
- **THEN** la consola indica que está cargando

### Requirement: Indicadores de lectura y edición en los mensajes

El historial SHALL indicar cuáles mensajes fueron leídos (`read_at` no nulo) y cuáles fueron editados (`edited_at` no nulo), con marcas visualmente distinguibles entre sí. Un mensaje eliminado SHALL mostrarse sin esas marcas.

#### Scenario: Mensaje leído

- **WHEN** un mensaje del historial tiene `read_at` no nulo y no está eliminado
- **THEN** el mensaje muestra la marca de leído

#### Scenario: Mensaje editado

- **WHEN** un mensaje del historial tiene `edited_at` no nulo y no está eliminado
- **THEN** el mensaje muestra la marca de editado, distinguible de la marca de leído

#### Scenario: Mensaje eliminado

- **WHEN** un mensaje del historial está en estado `deleted`
- **THEN** el mensaje no muestra ni la marca de leído ni la de editado

### Requirement: Desplazamiento del historial de mensajes

Al abrir una conversación, el sistema SHALL posicionar el historial en el último mensaje. Al cargar mensajes más antiguos, el sistema SHALL conservar la posición de lectura previa del historial.

#### Scenario: Abrir en el último mensaje

- **WHEN** el agente abre una conversación
- **THEN** el historial queda posicionado en su último mensaje

#### Scenario: Conservar la posición al cargar más

- **WHEN** el agente carga mensajes más antiguos con el historial desplazado en una posición intermedia
- **THEN** el historial mantiene a la vista el mismo punto que antes de la carga

### Requirement: Identificador de la conversación en el encabezado

El detalle SHALL mostrar el identificador de la conversación junto al identificador externo de su contacto.

#### Scenario: Encabezado con el id de la conversación

- **WHEN** el agente abre una conversación
- **THEN** el encabezado muestra el identificador externo del contacto y el identificador de la conversación
