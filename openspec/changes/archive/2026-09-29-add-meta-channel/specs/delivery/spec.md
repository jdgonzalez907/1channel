# Spec Delta

## RENAMED Requirements

- FROM: `### Requirement: Compose de producción sin base de datos`
- TO: `### Requirement: Compose de producción`

## MODIFIED Requirements

### Requirement: Enrutado del API por el mismo origen

Las solicitudes a `/v1/*` y `/webhooks/*` SHALL enrutarse al servicio `api` interno desde el mismo origen que sirve el cliente web, sin CORS. El servicio `api` SHALL NOT publicar puerto en el host: solo el contenedor `web` SHALL ser alcanzable desde internet.

#### Scenario: Llamada del cliente al API

- **WHEN** un navegador solicita `GET /v1/...` al origen del cliente
- **THEN** la solicitud llega al servicio `api` y su respuesta se devuelve por el mismo origen

#### Scenario: Callback de un proveedor externo

- **WHEN** un proveedor (Meta) solicita `GET` o `POST /webhooks/...` al origen del cliente
- **THEN** la solicitud llega al servicio `api` y su respuesta se devuelve por el mismo origen

#### Scenario: API no expuesta directamente

- **WHEN** se inspecciona el puerto publicado por el servicio `api` en producción
- **THEN** no hay puerto publicado al host; solo `web` expone `:80`

### Requirement: Compose de producción

El despliegue de producción SHALL componerse de los servicios `web`, `api` y `postgres`. La base de datos SHALL ser un servicio del propio compose con almacenamiento persistente. La conexión del `api` a la base de datos SHALL resolverse por variables de entorno, apuntando al servicio `postgres` del compose.

#### Scenario: Servicios del compose de producción

- **WHEN** se levanta el compose de producción
- **THEN** se inician `web`, `api` y `postgres`

#### Scenario: Base de datos por entorno

- **WHEN** `api` arranca en producción
- **THEN** resuelve la conexión a la base de datos a partir de las variables de entorno, apuntando al servicio `postgres` del compose

#### Scenario: Base de datos persistente

- **WHEN** se recrea el compose de producción
- **THEN** los datos de `postgres` persisten en el volumen del compose
