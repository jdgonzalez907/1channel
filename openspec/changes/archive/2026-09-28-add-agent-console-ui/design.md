# Design

## Context

Ver `proposal.md` - Why para la motivación. Estado observado que condiciona el diseño:

- `frontend/` es el shell de create-vue: dependencia única `vue`, `src/main.ts` monta `App.vue`, alias `@` -> `src`, sin router, sin store y sin cliente HTTP.
- El borde ya está resuelto: `vite.config.ts` proxea `/v1` a `http://localhost:8080` en dev y `nginx.conf` proxea `/v1/` al `api` en prod. La consola consume `/v1` en el mismo origen, sin CORS ni URL de API configurable.
- La API de lectura (`openspec/specs/http-read-api/spec.md`) devuelve la lista con `unread_count`, `last_message` y `next_*`; el detalle con `contact`, `agent_id`, `status`, `unread_count` y hasta 20 mensajes con `next_*`.
- El dominio (`openspec/specs/conversation/spec.md`) define: marcar leído es un caso de uso explícito del agente asignado; resolver es solo para `assigned` (una `pending` sin agente devuelve `ErrConversationAgentNotOwner`); responder a una `pending` asigna al agente y la pasa a `assigned`.
- La API de lectura **no** expone el `external_id` de los mensajes, solo su `id` interno.
- CI del frontend son `pnpm type-check` y `pnpm build`; no hay framework de tests.

## Goals / Non-Goals

**Goals:**

- Una consola de dos paneles (lista + detalle) con la menor superficie técnica posible: solo `vue`, `fetch` nativo y `localStorage`.
- Reflejar fielmente las reglas del dominio en qué acciones se muestran y cuándo (resolver, responder, marcar leído).
- Aprovechar el cursor `before_sent_at`/`before_id` en lista y detalle con una acción "Cargar más".
- Un simulador de webhook que permita generar y modificar conversaciones desde la misma pantalla.

**Non-Goals:**

- Enrutador, store global, librería HTTP o framework de UI/CSS.
- Tiempo real (WebSocket) o refresco automático por sondeo.
- Login/sesión real: el identificador del agente es un token pegado a mano.
- Cambios en el backend, en el contrato `/v1`, SQLC, migraciones o base de datos.
- Tests automatizados del frontend (fuera de los gates de type-check y build).

## Decisions

### D1 - Sin router, sin store, sin librería HTTP

Composición con `App.vue` como orquestador y componentes de presentación hijos. Las llamadas usan `fetch` nativo envueltas en un módulo `api/`. El estado de la sesión (`agentId`) y de la navegación (conversación seleccionada, estado, filtro) vive en `App.vue`; los hijos reciben props y emiten eventos.

- Razón: la pantalla es un caso único sin rutas ni estado compartido complejo; añadir `vue-router` o `pinia` no aporta. El proyecto pide "simplicidad" y "sin dependencias nuevas".
- Alternativa descartada: `vue-router` con la conversación seleccionada en la URL. Se puede adoptar después si aparece deep-linking; hoy no hace falta.

### D2 - Estructura de archivos

```
frontend/src/
├── main.ts                       # sin cambios (monta App.vue)
├── App.vue                       # barra del agente + layout de dos paneles + estado de la sesión
├── format.ts                     # formateo de fechas (RFC 3339 -> hora local)
├── api/
│   ├── client.ts                 # fetch + Bearer + parseo de problem+json -> ApiError
│   └── types.ts                  # tipos de las respuestas/entradas de /v1
└── components/
    ├── ConversationList.vue      # estado, filtro, items en línea, "Cargar más"
    ├── ConversationDetail.vue    # encabezado, historial en línea, responder, resolver, "Cargar más"
    └── WebhookSimulator.vue      # modal del simulador
```

- `App.vue` conserva `agentId` (restaurado de `localStorage` y validado con `GET /v1/users/{id}`), `selectedId`, `status` y `contactFilter`, y conecta la lista con el detalle: la lista emite `select`, el detalle emite `changed` y `App.vue` fuerza recargar la lista (por ejemplo con una `reloadKey`).
- Alternativa descartada: que cada componente haga su propio fetch y coordine por eventos globales; más piezas y estado duplicado.

### D3 - Cliente tipado y errores problem+json

`api/client.ts` expone funciones delgadas (`listConversations`, `getConversation`, `sendMessage`, `markRead`, `resolveConversation`, `simulateWebhook`, `getUser`) que centralizan: la cabecera `Authorization: Bearer <agentId>`, el parseo de JSON y la conversión de cualquier respuesta no-2xx en un `ApiError` con `status`, `title` y `detail` (leídos de `application/problem+json`, con fallback genérico). Los componentes solo muestran `error.title`/`error.detail`.

- Razón: el contrato de error ya es uniforme (`http-errors`), así que un único punto de traducción evita repetir manejo en cada llamada.
- Los tipos de `api/types.ts` reflejan los Response de los specs de lectura y escritura (no se generan desde el backend; se escriben a mano).

### D4 - Paginación por cursor en ambos paneles

Ambos listados guardan el par `(next_before_sent_at, next_before_id)` de la última respuesta. La acción "Cargar más" solo aparece cuando ambos son no nulos y consulta pasando `before_sent_at`/`before_id`, anexando el resultado. El historial del detalle mantiene el orden cronológico ascendente que ya devuelve la API.

- Razón: el cursor explícito es el mecanismo que define el contrato; un botón es la forma más simple de exponerlo sin scroll infinito.
- Alternativa descartada: scroll infinito (más lógica de observadores y sin ventaja para un panel interno).

### D5 - Acciones visibles según el estado, espejo del dominio

| Acción | Se muestra / dispara cuando | Regla de dominio |
|---|---|---|
| Responder (campo de texto) | `status` es `pending` o `assigned` | una finalizada responde 409 |
| Resolver (botón) | `status === "assigned"` | resolver exige agente asignado; `pending` da 403 |
| Marcar leído (automático al abrir) | `agent_id === agentId` y `unread_count > 0` | solo el asignado; funciona también en finalizadas (`resolved`/`expired` asignadas) |
| Marcar leído en `pending` | no se intenta | sin agente asignado el dominio lo rechaza |

- Decisión confirmada en exploración: marcar leído es **automático al abrir** una conversación asignada al agente, no un botón. En una conversación sin agente asignado (`pending` o `expired` sin agente) el dominio lo rechaza y no se intenta.
- Razón: "abrir = leer" es una lectura natural del caso de uso "agente lee conversación", y evita un botón extra. Tras marcar, se refresca la lista para bajar el contador. El detalle se actualiza **en sitio** (se pone `unread_count` en 0 y se marcan los mensajes del contacto como leídos), sin volver a pedir la conversación, para no perder el foco/scroll con un re-render.

### D8 - Marcas de leído y editado en los mensajes

Cada mensaje del historial muestra **leído** cuando `read_at` no es nulo y **editado** cuando `edited_at` no es nulo, con colores distintos (verde/ámbar) y el mismo peso y tamaño. Un mensaje `deleted` no muestra ninguna de las dos marcas.

- Razón: los dos campos ya vienen en el contrato de lectura (`http-read-api`); es información útil para el agente y para probar el flujo (los mensajes del contacto se marcan leídos al abrir la conversación; los del agente al recibir `message.read` por el simulador).
- Alternativa descartada: derivarlas de `status`, que colapsa `deleted`/`read` en un solo valor y pierde el matiz de "leído pero editado".

### D9 - Desplazamiento del historial

El historial es el único contenedor con scroll del detalle. Al abrir una conversación se posiciona en el último mensaje (`scrollTop = scrollHeight`). Al cargar mensajes más antiguos se guarda `scrollHeight`/`scrollTop` antes de anteponerlos y se restaura `scrollTop = nuevoScrollHeight - alturaPrevia + topPrevio`, para no perder la posición de lectura.

- Razón: sin lo primero el agente aterriza al inicio del historial; sin lo segundo, precargar mensajes viejos saltaría al fondo o al tope.
- El encabezado suma el id de la conversación debajo del `external_id` del contacto, como ayuda para probar el simulador (que pide `external_contact_id`/`external_message_id`).

### D6 - Simulador de webhook desde el panel de la lista

El modal se abre desde el panel de la lista, para que funcione sin conversación seleccionada (es justo el caso de `message.received` con un contacto nuevo, que crea una conversación). Si hay una seleccionada, se prellena `external_contact_id` con su `contact.external_id`.

- El `event` es un select con `message.received` por defecto; `text` se muestra solo para `received`/`edited`.
- Para `edited`/`deleted`/`read` el `external_message_id` se escribe a mano: la API de lectura no expone el identificador externo de los mensajes. Alternativa descartada: agregar `external_id` al contrato de lectura (cambio de backend fuera de alcance de este change).
- Tras un 204 se refresca la lista y, si el contacto coincide, el detalle.

### D7 - Fechas, estilo y scroll

Los timestamps llegan en RFC 3339 UTC; la consola los formatea a hora local con `Intl.DateTimeFormat`. El estilo es CSS plano sin librería de UI: layout de dos columnas con `flex`, apilado en pantallas angostas, con colores en los badges de estado (`pending`/`assigned`/`resolved`/`expired`) y en el contador de no leídos, que se alinea a la derecha del item (verde salvo en `expired`, que va gris: una `expired` puede no tener agente y entonces no se marca como leída al abrirla). Los mensajes del agente se alinean a la derecha y los del contacto a la izquierda. La lista y el historial respetan su contenedor con scroll interno (`overflow-y: auto`), dejando fijos los controles de la lista, el encabezado del detalle y el campo de respuesta. Se cuidan estados vacíos, de carga y de error (texto legible), que es la "UX" pedida sin agregar una librería.

## Risks / Trade-offs

- [El token (id del agente) queda en `localStorage` en claro] -> Es una herramienta interna de mismo origen sobre un Bearer que no es un secreto de producción; aceptable para este alcance, documentado como no-goal de "sesión real".
- [Sin tiempo real, la lista no se actualiza sola al llegar mensajes del contacto] -> El simulador y las acciones propias refrescan explícitamente; un refresco por sondeo queda como mejora posterior.
- [Marcar leído automático puede sorprender] -> Se limita a conversaciones asignadas con no leídos y se refleja el cambio en la lista.
- [`external_message_id` manual para edit/delete/read del simulador] -> Se explica en el propio modal; es una limitación del contrato de lectura, no del simulador.
- [Sin tests del frontend] -> Verificación manual con `make run` + `make run-web` y el simulador; los gates de CI siguen siendo type-check y build.
- [Dos capabilities en un mismo change] -> `agent-console` y `webhook-simulator` comparten el mismo `api/client.ts` y la misma barra de agente en `App.vue`; mantenerlas en un change evita repetir el andamiaje.

## Migration Plan

- No hay migración de datos ni de esquema: el cambio es solo de `frontend/src/`. El backend y el contrato `/v1` no se tocan.
- Despliegue: se reconstruye y publica solo la imagen `web` (por sha); `api` no cambia. Orden: sin dependencias de esquema, puede desplegarse en cualquier momento.
- Rollback: volver a la imagen `web` anterior por sha; no hay estado persistido que revertir (lo único persistido es el `agentId` en el navegador).

## Open Questions

- Formato exacto de fecha mostrado (absoluto vs. relativo); se puede ajustar durante la implementación sin cambiar los specs.
- Si más adelante conviene un refresco periódico de la lista; hoy queda fuera por simplicidad.
