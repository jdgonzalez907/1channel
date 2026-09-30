# Spec Delta

## Purpose

Capa pública del SPA de 1Channel: presenta el producto a empresas y publica las páginas legales, además de definir la navegación por rutas entre la landing, la consola del agente y los documentos legales.

## ADDED Requirements

### Requirement: Navegación por rutas del SPA

El SPA SHALL servir la landing en `/`, la consola del agente en `/consola`, la política de privacidad en `/privacidad` y los términos y condiciones en `/terminos`. La navegación entre estas páginas SHALL ser del lado del cliente, sin recarga completa. Cada ruta SHALL ser direccionable: cargar o recargar la URL directamente SHALL resolver la misma página. Una ruta desconocida SHALL resolver a la landing.

#### Scenario: Landing en la raíz

- **WHEN** se accede a `/`
- **THEN** el SPA muestra la landing

#### Scenario: Consola en su ruta

- **WHEN** se accede a `/consola`
- **THEN** el SPA muestra la consola del agente

#### Scenario: Acceso directo a una página legal

- **WHEN** se carga o recarga la URL `/privacidad` o `/terminos` de forma directa
- **THEN** el SPA muestra esa página legal y no la landing

#### Scenario: Navegación sin recarga

- **WHEN** el usuario navega entre páginas del SPA
- **THEN** el cambio de página no recarga el documento completo

#### Scenario: Ruta desconocida

- **WHEN** se accede a una ruta que no corresponde a ninguna página
- **THEN** el SPA muestra la landing

### Requirement: Landing pública del producto

La landing SHALL ser accesible sin autenticación y SHALL presentar a 1Channel como un CRM conversacional multicanal operado por varios agentes. SHALL incluir una introducción al producto, una explicación del flujo (el contacto escribe por un canal, el mensaje cae en una bandeja unificada y un agente responde) y una sección de capacidades. La landing SHALL NOT solicitar ni depender de un identificador de agente para mostrarse.

#### Scenario: Acceso sin credenciales

- **WHEN** un visitante sin identificador de agente accede a `/`
- **THEN** la landing se muestra completa y no pide credenciales

#### Scenario: Contenido mínimo de la landing

- **WHEN** la landing se muestra
- **THEN** incluye introducción, explicación del flujo y capacidades

### Requirement: Llamado a la acción hacia la consola

La landing SHALL incluir un llamado a la acción visible que navegue a `/consola`.

#### Scenario: Entrar a la consola

- **WHEN** el visitante activa el llamado a la acción de la landing
- **THEN** el SPA navega a `/consola`

### Requirement: Política de privacidad

La página `/privacidad` SHALL describir, en español, el tratamiento de datos personales: el responsable (una persona natural en Colombia, identificada por su nombre y su correo de contacto), el canal para asuntos de datos personales (`jdgonzalez907@gmail.com`), las categorías de datos tratados (datos de contacto e identificación, contenido de las conversaciones e identificadores de canal), las finalidades, la retención, la transferencia a terceros proveedores de canal (por ejemplo Meta/Messenger) y los derechos del titular con la forma de ejercerlos. SHALL incluir una sección que aclare que, frente a los contactos finales, 1Channel actúa como encargado y la empresa cliente como responsable. SHALL mostrar un aviso visible de que el texto es un borrador pendiente de revisión legal.

#### Scenario: Identificación del responsable

- **WHEN** se muestra `/privacidad`
- **THEN** indica quién es el responsable del tratamiento y su correo de contacto

#### Scenario: Derechos del titular

- **WHEN** se muestra `/privacidad`
- **THEN** enumera los derechos del titular y cómo ejercerlos

#### Scenario: Rol de encargado

- **WHEN** se muestra `/privacidad`
- **THEN** aclara que, respecto de los contactos finales, 1Channel es encargado y la empresa cliente es responsable

#### Scenario: Aviso de revisión legal

- **WHEN** se muestra `/privacidad`
- **THEN** muestra un aviso visible de que el texto está pendiente de revisión legal

### Requirement: Términos y condiciones

La página `/terminos` SHALL describir, en español, las condiciones de uso del servicio para las empresas clientes: descripción del servicio, uso aceptable, cuentas de agente y acceso, propiedad y uso de los datos, disponibilidad, terminación, ley aplicable (Colombia) y un canal de contacto. SHALL mostrar un aviso visible de que el texto es un borrador pendiente de revisión legal.

#### Scenario: Condiciones del servicio

- **WHEN** se muestra `/terminos`
- **THEN** describe el uso aceptable, las cuentas de agente, la propiedad de los datos, la disponibilidad, la terminación y la ley aplicable

#### Scenario: Aviso de revisión legal de los términos

- **WHEN** se muestra `/terminos`
- **THEN** muestra un aviso visible de que el texto está pendiente de revisión legal

### Requirement: Layout público compartido

La landing y las páginas legales SHALL compartir un layout público con un encabezado simple y un pie de página. El encabezado del layout público SHALL NOT incluir el formulario de identidad del agente. La consola SHALL quedar fuera de este layout.

#### Scenario: Encabezado público sin identidad de agente

- **WHEN** se muestra cualquier página pública
- **THEN** su encabezado no incluye el formulario de identidad del agente

#### Scenario: Consola fuera del layout público

- **WHEN** se muestra la consola
- **THEN** no se presenta el encabezado ni el pie del layout público

### Requirement: Pie de página con enlaces legales

El pie del layout público SHALL mostrar el nombre 1Channel, una nota de copyright con el año y enlaces navegables a la política de privacidad y a los términos y condiciones. El correo de contacto SHALL NOT figurar en el pie; su publicación se limita al contenido de las páginas legales.

#### Scenario: Enlaces legales disponibles

- **WHEN** se muestra una página pública
- **THEN** su pie incluye enlaces a `/privacidad` y `/terminos`

#### Scenario: Navegar desde el pie

- **WHEN** el usuario activa el enlace de privacidad o de términos del pie
- **THEN** el SPA navega a la página legal correspondiente

#### Scenario: Pie en todas las páginas públicas

- **WHEN** se muestra la landing o cualquiera de las páginas legales
- **THEN** el pie de página está presente

### Requirement: Identidad visual del sitio público

Las páginas públicas SHALL usar el color de acento `#7540BF` en los elementos de marca y en los llamados a la acción, sobre una paleta neutra de estética minimalista (tipografía de sistema, bordes finos y ausencia de decoración superflua). La identidad visual y los colores de la consola SHALL NOT modificarse.

#### Scenario: Acento en el llamado a la acción

- **WHEN** se muestra la landing
- **THEN** el llamado a la acción principal usa el color de acento `#7540BF`

#### Scenario: Acento en los elementos de marca

- **WHEN** se muestra una página pública
- **THEN** los elementos de marca y enlaces usan el color de acento sobre una paleta neutra

#### Scenario: Consola sin cambios visuales

- **WHEN** se muestra la consola en `/consola`
- **THEN** conserva sus colores y estilos anteriores, sin el acento `#7540BF`

### Requirement: Título del documento por ruta

Cada ruta SHALL fijar el título del documento del navegador con el nombre de la página que se muestra.

#### Scenario: Título según la página

- **WHEN** se navega a la landing, a la consola o a una página legal
- **THEN** el título del documento corresponde a esa página
