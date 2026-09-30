# delivery Specification

## Purpose

Define cómo se construyen, sirven y despliegan de forma independiente el cliente web (`web`) y la API (`api`) en un mismo repositorio y un mismo host, garantizando un borde único de entrada y releases desacoplados por componente.

## Requirements

### Requirement: Hosting del cliente web en la raíz

El sistema SHALL servir el cliente web compilado en `/` desde el mismo origen que la API. Las rutas que no correspondan a un archivo estático SHALL resolverse contra `index.html` (fallback de history), de modo que el enrutador del cliente maneje rutas profundas.

#### Scenario: Carga del cliente

- **WHEN** un navegador solicita `GET /`
- **THEN** el sistema responde 200 con el documento HTML del cliente

#### Scenario: Ruta profunda del enrutador del cliente

- **WHEN** un navegador solicita `GET /cualquier/ruta/del/cliente` que no es un archivo estático
- **THEN** el sistema responde 200 con `index.html` para que el cliente resuelva la ruta

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

### Requirement: Construcción independiente por componente

Cada componente SHALL construirse y publicarse únicamente cuando cambian sus propias fuentes. Un cambio limitado a `frontend/**` SHALL NOT reconstruir ni republicar la imagen `api`, y un cambio limitado a `backend/**` SHALL NOT reconstruir ni republicar la imagen `web`.

#### Scenario: Cambio solo en el front

- **WHEN** se integra un commit que solo modifica `frontend/**`
- **THEN** se construye y publica una nueva imagen `web` y no se reconstruye `api`

#### Scenario: Cambio solo en el back

- **WHEN** se integra un commit que solo modifica `backend/**`
- **THEN** se construye y publica una nueva imagen `api` y no se reconstruye `web`

#### Scenario: Cambio que no toca ningún componente

- **WHEN** se integra un commit que no modifica `frontend/**` ni `backend/**`
- **THEN** no se reconstruye ninguna de las dos imágenes

### Requirement: Identidad de release por sha por componente

Cada imagen SHALL etiquetarse con el sha del commit que la construyó, y cada componente SHALL mantener un puntero a su último build exitoso. El despliegue SHALL poder fijar un sha distinto por componente, de modo que `web` y `api` corran versiones de commits diferentes a la vez.

#### Scenario: Shas independientes

- **WHEN** `api` no ha cambiado desde un commit anterior y `web` sí
- **THEN** el despliegue puede usar el sha previo de `api` junto con el sha nuevo de `web`

#### Scenario: Puntero por componente

- **WHEN** se publica una nueva imagen de un componente en su rama principal
- **THEN** su etiqueta de "último build" apunta a esa imagen sin afectar la del otro componente

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

### Requirement: Compatibilidad del contrato bajo `/v1`

El cliente web SHALL consumir la API bajo el prefijo `/v1`. Un cambio incompatible de la API SHALL exponerse bajo un prefijo de versión nuevo y SHALL NOT retirar `/v1` mientras exista un cliente desplegado que lo consuma.

#### Scenario: Cambio incompatible

- **WHEN** se introduce un cambio incompatible en el contrato de la API
- **THEN** se publica bajo un prefijo nuevo y `/v1` sigue respondiendo para los clientes existentes
