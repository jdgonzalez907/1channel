# Proposal

## Why

Hoy el frontend de 1Channel es únicamente la consola interna del agente: la raíz `/` exige un identificador de agente y no existe ninguna cara pública. No hay página de inicio que explique el producto, ni política de privacidad ni términos y condiciones. Esto es un hueco operativo: operar el servicio y la app de Messenger de Meta exige una Privacy Policy URL propia, y un producto B2B necesita una landing que presente la propuesta de valor antes de pedir credenciales.

## What Changes

- Se introduce `vue-router` (history mode) y una tabla de rutas explícita: `/` landing, `/consola` consola del agente, `/privacidad`, `/terminos`, y un catch-all que resuelve a la landing.
- `App.vue` pasa a ser un `<RouterView/>` sin chrome propio. La consola se mueve **tal cual** a `views/AgentConsoleView.vue`: mismo marcado, mismos estilos y mismos colores. **BREAKING** para quien tenga `/` guardado: la consola deja de vivir en la raíz y pasa a `/consola`.
- Se agrega una capa pública con layout propio (header simple + footer legal) que envuelve la landing y las páginas legales. La consola conserva su layout actual y **no** usa el layout público.
- Landing B2B en español: hero, "cómo funciona", capacidades (bandeja unificada, varios agentes, canales modulares) y CTA "Entrar a la consola" hacia `/consola`.
- Política de privacidad y Términos y condiciones en español, redactados como borrador con aviso visible de "pendiente de revisión legal", para una persona natural en Colombia (responsable: Juan David Gonzalez Bedoya, contacto `jdgonzalez907@gmail.com`). La política incluye una sección que aclara que, frente a los contactos finales, 1Channel actúa como **encargado** y la empresa cliente como **responsable**.
- Identidad visual pública: accent `#7540BF`, estética Tech Minimalist (neutros, tipografía de sistema, mono para datos, bordes finos), definida en tokens. Aplica **solo** a las páginas públicas; la consola no migra a tokens ni cambia su diseño.
- Cada ruta fija el título del documento.

## Capabilities

### New Capabilities

- `public-site`: capa pública del SPA: navegación cliente por rutas (landing, consola, legales), landing B2B, páginas de política de privacidad y términos y condiciones, footer con enlaces legales y la identidad visual del sitio público (accent `#7540BF`), sin alterar el comportamiento ni el diseño de la consola.

### Modified Capabilities

- Ninguna. El comportamiento de `agent-console` no cambia: la consola sigue haciendo lo mismo y con el mismo aspecto; solo cambia la URL en la que se sirve, responsabilidad que asume la navegación de `public-site`.

## Impact

- **Frontend (Vue)**: nueva dependencia `vue-router` (y actualización de `package.json`/`pnpm-lock.yaml`); `src/main.ts` registra el router; `App.vue` se reduce a `<RouterView/>`; nuevos `src/router/`, `src/views/` (`LandingView`, `AgentConsoleView`, `PrivacyView`, `TermsView`), `src/layouts/PublicLayout.vue` y `src/assets/theme.css`; `index.html` ajusta el título base.
- **Consola**: `src/App.vue` actual se renombra/mueve a `views/AgentConsoleView.vue` sin editar su marcado, estilos ni comportamiento.
- **Build/servido**: `nginx.conf` no cambia (ya tiene `try_files ... /index.html`); `vite.config.ts` tampoco. El SPA fallback ya soporta rutas profundas.
- **Backend / API `/v1` / SQLC / migraciones**: sin cambios.
- **Docs**: `frontend/README.md` deja de afirmar que no hay router y documenta las rutas públicas.
- **Contenido legal**: texto orientativo, no asesoría jurídica; requiere revisión legal antes de considerarse definitivo.
