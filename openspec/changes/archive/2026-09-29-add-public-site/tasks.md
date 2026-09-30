# Tasks

## 1. Router y estructura base

- [x] 1.1 Agregar `vue-router` a `frontend/package.json` (instalar con `pnpm --dir frontend install`) y verificar que se actualicen `package.json` y `pnpm-lock.yaml` sin errores.
- [x] 1.2 Crear `frontend/src/router/index.ts` con `createWebHistory`, `scrollBehavior` que vuelve al tope y la tabla de rutas (landing, `/consola`, `/privacidad`, `/terminos`, catch-all a `/`); registrar el router en `frontend/src/main.ts` con `app.use(router)`. Verificar con `pnpm --dir frontend type-check`.
- [x] 1.3 Reducir `frontend/src/App.vue` a `<RouterView />` (sin chrome ni estilos de la consola) y verificar con `pnpm --dir frontend build` que el SPA compila.

## 2. Mudanza de la consola sin cambios

- [x] 2.1 Mover el contenido actual de `frontend/src/App.vue` a `frontend/src/views/AgentConsoleView.vue` con su `<template>` y `<style scoped>` idénticos; conectar la ruta `/consola`. Verificar con `pnpm --dir frontend build` y comparando el componente movido contra el original para confirmar que no hay cambios de marcado ni de estilos.
- [x] 2.2 Verificar en el navegador (`make run-web` con la API arriba) que la consola en `/consola` mantiene su interfaz y comportamiento: barra del agente, tres paneles, listado, detalle, panel de persona, simulador y restauración de sesión desde `localStorage`.

## 3. Layout público e identidad visual

- [x] 3.1 Crear `frontend/src/assets/theme.css` con las variables de la identidad pública (accent `#7540bf`, neutros, bordes, radios, tipografía) definidas sobre la raíz del layout público y no sobre `:root`; importarlo en `main.ts`. Verificar con `pnpm --dir frontend build` y comprobando en el navegador que la consola no hereda cambios de color.
- [x] 3.2 Crear `frontend/src/layouts/PublicLayout.vue` con encabezado simple (wordmark 1Channel, sin formulario de agente), `<RouterView />` y pie con copyright, correo de contacto y enlaces a `/privacidad` y `/terminos`; envolver las rutas públicas como rutas hijas y dejar `/consola` fuera del layout. Verificar en el navegador que los enlaces del pie navegan a cada página legal y que la consola no muestra el encabezado ni el pie públicos.

## 4. Landing

- [x] 4.1 Crear `frontend/src/views/LandingView.vue` con las secciones de introducción, "cómo funciona" (contacto -> canal -> bandeja -> agente) y capacidades (bandeja unificada, varios agentes, canales modulares), usando la estética Tech Minimalist y el accent `#7540bf`. Verificar visualmente en `/` sin identificador de agente y con `pnpm --dir frontend build`.
- [x] 4.2 Agregar el llamado a la acción principal hacia `/consola` (y un secundario dentro de la misma página) y verificar en el navegador que navega a la consola sin recargar el documento.

## 5. Páginas legales

- [x] 5.1 Crear `frontend/src/views/PrivacyView.vue` con el borrador en español: responsable (persona natural en Colombia y su correo de contacto), canal de datos personales (`jdgonzalez907@gmail.com`), categorías de datos, finalidades, retención, terceros (Meta/Messenger), derechos del titular, sección de encargado/responsable para contactos finales, y aviso visible de revisión legal. Verificar en `/privacidad`, incluida una recarga directa de la URL, y con `pnpm --dir frontend build`.
- [x] 5.2 Crear `frontend/src/views/TermsView.vue` con el borrador en español: descripción del servicio, uso aceptable, cuentas de agente y acceso, propiedad y uso de datos, disponibilidad, terminación, ley aplicable (Colombia), canal de contacto y aviso visible de revisión legal. Verificar en `/terminos`, incluida una recarga directa de la URL, y con `pnpm --dir frontend build`.

## 6. Metadatos y documentación

- [x] 6.1 Declarar `meta.title` en cada ruta y fijar `document.title` en un guard `afterEach`; ajustar el título base de `frontend/index.html`. Verificar que el título del documento cambie al navegar entre landing, consola, privacidad y términos.
- [x] 6.2 Actualizar `frontend/README.md`: reemplazar la afirmación de "sin router" por la tabla de rutas, el comando para llegar a la consola en `/consola` y una nota de que los textos legales son un borrador pendiente de revisión legal. Verificar que los comandos documentados corran tal cual.

## 7. Verificación de integración

- [x] 7.1 Ejecutar los gates del frontend `pnpm --dir frontend type-check` y `pnpm --dir frontend build` y confirmar que terminan sin errores.
- [x] 7.2 Con `make up`, `make migrate-up`, `make seed`, `make run-api` y `make run-web`: cargar `/`, navegar a `/consola` por el CTA, entrar con un id de agente del seed y operar una conversación; recargar directamente `/privacidad` y `/terminos`; y acceder a una ruta desconocida confirmando que cae en la landing.
