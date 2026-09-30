# Design

## Context

Ver `proposal.md` - Why para la motivación. Estado observado que condiciona el diseño:

- `frontend/` es un SPA de una sola pantalla: `src/main.ts` monta `App.vue`, que **es** la consola del agente. No hay router, no hay store y no hay capa pública. El `README.md` del frontend presume explícitamente "sin router ni store".
- `App.vue` mezcla el chrome de la consola (barra con el campo `Id del agente`) con el layout de tres paneles (`lista | detalle | persona`), con estilos scoped y colores hardcodeados. La consola no debe cambiar.
- `nginx.conf` ya resuelve rutas profundas con `try_files $uri $uri/ /index.html`, y `vite dev` maneja el mismo fallback. Es decir, el servidor **ya** soporta rutas cliente; solo falta que el SPA las interprete.
- El contrato `/v1` y el backend no se tocan. La landing y los legales son estáticos, sin llamadas a la API.
- CI del frontend: `pnpm type-check` y `pnpm build`. No hay tests de frontend.

## Goals / Non-Goals

**Goals:**

- Introducir navegación por rutas reales con la mínima dependencia posible, aprovechando el fallback de nginx ya existente.
- Servir una landing B2B y dos páginas legales como URLs propias y compartibles.
- Mover la consola a `/consola` **sin tocar una línea de su interfaz ni de su comportamiento**.
- Encapsular la identidad visual pública (accent `#7540BF`, Tech Minimalist) para que no se filtre a la consola.

**Non-Goals:**

- No rediseñar la consola ni migrarla a tokens de color.
- No autenticar la landing ni las páginas legales; el acceso a la consola sigue siendo el `Id del agente`.
- No internacionalización: todo el sitio público va en español.
- No SSR ni prerender: es un SPA servido por nginx.
- No formularios de contacto, newsletter, analítica ni backend nuevo.

## Decisions

### D1 - `vue-router` en history mode

Se agrega `vue-router` con `createWebHistory()` y `scrollBehavior` que vuelve al tope en cada navegación. Tabla de rutas:

```
/            -> LandingView      (layout publico)
/consola     -> AgentConsoleView (layout propio de la consola)
/privacidad  -> PrivacyView      (layout publico)
/terminos    -> TermsView        (layout publico)
/:pathMatch(.*)* -> redirige a /
```

- Razón: `vue-router` es el estándar y el costo es una dependencia chica; las URLs legales deben existir de verdad (Meta exige una Privacy Policy URL). El fallback de nginx y de Vite ya hace que una recarga en `/privacidad` funcione sin tocar el servidor.
- Alternativas descartadas: (a) mini-router propio con `location.pathname` — pura dependencia cero, pero termina siendo un router peor y más caro de mantener; (b) hash routing (`#/privacidad`) — no requiere fallback, pero produce URLs feas y peores para compartir/SEO; (c) renderizar todo en una página con anclas — los legales dejan de ser URLs propias.

### D2 - La consola se mueve sin reescribirse

`src/App.vue` actual se traslada **verbatim** a `src/views/AgentConsoleView.vue` (mismo `<template>`, mismo `<style scoped>`, mismos colores). El nuevo `App.vue` queda reducido a `<RouterView />`.

- Consecuencia deliberada: **BREAKING** para quien tenga `/` guardado; la consola pasa a `/consola`. El estado ya persistido (`localStorage` con `agent-console.agent-id`) sigue vigente porque la clave no cambia.
- No se agrega al header de la consola un enlace de vuelta a la landing: eso tocaría la interfaz de la consola. Se llega por el CTA de la landing o por URL.
- Alternativa descartada: mantener la consola en `/` y poner la landing en otra ruta. La landing es la cara pública del producto y debe vivir en la raíz.

### D3 - Layout público con rutas anidadas

Las páginas públicas se envuelven en un `PublicLayout.vue` (encabezado simple + `<RouterView />` + pie) usando rutas hijas; la consola es una ruta hermana sin ese layout.

```
routes
  { path: '/', component: PublicLayout,
    children: [
      { path: '',           component: LandingView },
      { path: 'privacidad', component: PrivacyView },
      { path: 'terminos',   component: TermsView },
    ] }
  { path: '/consola', component: AgentConsoleView }
```

- Razón: el encabezado/pie son comunes a las tres páginas públicas y ausentes en la consola; anidar expresa eso mejor que duplicar el chrome en cada vista.
- Alternativa descartada: condicionar el chrome dentro de `App.vue` según la ruta; mezcla la consola con lo público y arriesga tocar su render.

### D4 - Tokens de diseño acotados a lo público

`src/assets/theme.css` define las variables (`--accent: #7540bf`, neutros de texto/borde/fondo, radios y tipografía) **sobre la raíz del layout público** (`.public-site`), no sobre `:root`. Las páginas públicas consumen esas variables; la consola, que vive fuera de `.public-site`, no las hereda.

- Razón: es la forma de introducir una identidad visual nueva sin que la consola cambie de aspecto por herencia de variables globales.
- Alternativa descartada: tokens en `:root` + refactor de la consola a tokens; es lo correcto a futuro, pero el alcance acordado es no tocar la consola.

### D5 - Estética Tech Minimalist

Sin framework de CSS ni de UI. La landing se compone con CSS plano y unas pocas secciones: hero (título, subtítulo, CTA primario y secundario), "cómo funciona" en tres pasos (`contacto -> canal -> bandeja -> agente`), grilla de capacidades y un bloque de cierre con el mismo CTA. El acento `#7540BF` se usa con moderación: CTA principal, enlaces, marcadores de sección y el wordmark. El resto es neutro: tipografía de sistema, bordes de 1px, radios chicos, sin sombras ni gradientes. La tipografía monoespaciada se reserva para datos/identificadores.

- Razón: "minimalista" se logra con restricción, no con una librería; el repo ya usa CSS plano y no quiere dependencias de UI.

### D6 - Contenido legal como plantillas estáticas

`PrivacyView.vue` y `TermsView.vue` contienen el texto en marcado estático (sin markdown ni CMS), en español, con un componente/aviso de "borrador pendiente de revisión legal" visible y consistente. La política cubre responsable, datos, finalidades, retención, terceros (Meta/Messenger), derechos y la sección de encargado/responsable para contactos finales. Los términos cubren servicio, uso aceptable, cuentas de agente, datos, disponibilidad, terminación, ley aplicable y contacto.

- Razón: no hay pipeline de contenido y el texto cambia poco; markup estático es lo más simple y evita sumar un parser de markdown.
- El texto es orientativo y no asesoría jurídica; el aviso lo deja explícito.

### D7 - Título del documento por ruta

El router declara `meta.title` en cada ruta y un `afterEach` fija `document.title`. `index.html` mantiene un título base para la carga inicial.

- Razón: requisito de SEO/UX barato; sin él todas las páginas comparten el título.

### D8 - Sin cambios de build ni de servido

`nginx.conf`, `vite.config.ts` y el `Dockerfile` del frontend no cambian: el fallback SPA y el proxy `/v1` ya existen. El despliegue se limita a reconstruir la imagen `web`.

- Razón: el único requisito de servidor para history mode ya estaba resuelto.

## Risks / Trade-offs

- [La consola deja de estar en `/`; bookmarks y enlaces viejos caen en la landing] -> Cambio BREAKING declarado; el CTA de la landing lleva a `/consola` y el `localStorage` del agente sobrevive. No se agrega redirect porque la raíz ahora es la landing.
- [Convivencia de dos paletas: consola azul y sitio público con `#7540BF`] -> Es la consecuencia explícita de no tocar la consola; los tokens están acotados a `.public-site` para que no haya sangrado.
- [El texto legal no es asesoría jurídica] -> Aviso visible de "pendiente de revisión legal" en ambas páginas; se documenta en el README.
- [`vue-router` contradice el "sin router" que afirma el README] -> Se actualiza el README del frontend como parte del cambio.
- [History mode depende del fallback del servidor] -> Ya está en `nginx.conf` y en Vite; un servidor distinto al `web` actual requeriría configuración equivalente.
- [Sin tests de frontend] -> Verificación manual de rutas, recargas directas y del aspecto de la consola; los gates siguen siendo `type-check` y `build`.

## Migration Plan

- No hay migración de datos ni de esquema. El cambio toca solo `frontend/`.
- Despliegue: reconstruir y publicar la imagen `web` (por sha). `api` no cambia y no hay dependencias de esquema.
- Rollback: volver a la imagen `web` anterior por sha; lo único persistido es el `agentId` en el navegador, que no se toca.

## Open Questions

- Texto legal definitivo y revisión jurídica: el borrador se puede ajustar sin cambiar los specs.
- Copy final de la landing (titulares y descripción de capacidades): ajustable dentro de la estructura de secciones definida.
