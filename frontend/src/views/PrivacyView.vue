<template>
  <div class="legal">
    <div class="container">
      <h1>Política de privacidad</h1>
      <p class="meta">Última actualización: septiembre de 2026</p>

      <p class="legal-notice">
        Borrador pendiente de revisión legal. Este texto es orientativo y no constituye
        asesoría jurídica. Debe ser revisado por un profesional del derecho antes de
        considerarse definitivo.
      </p>

      <p>
        1Channel es un sistema de gestión de conversaciones que permite a una empresa
        centralizar y responder los mensajes de sus clientes a través de distintos
        canales. Esta política describe cómo se tratan los datos personales asociados
        al servicio.
      </p>

      <h2>1. Responsable del tratamiento</h2>
      <p>
        El responsable del tratamiento es <strong>Juan David Gonzalez Bedoya</strong>,
        persona natural con domicilio en Colombia. Para cualquier asunto relacionado
        con datos personales, el canal de contacto es
        <a href="mailto:jdgonzalez907@gmail.com">jdgonzalez907@gmail.com</a>.
      </p>

      <h2>2. Datos que tratamos</h2>
      <ul>
        <li>Datos de contacto e identificación: documento, nombres, teléfono, correo y dirección.</li>
        <li>Contenido de las conversaciones y sus metadatos (fechas, estado y lecturas).</li>
        <li>Identificadores de canal, como el identificador externo del contacto en Messenger.</li>
      </ul>

      <h2>3. Finalidades</h2>
      <ul>
        <li>Prestar el servicio de bandeja unificada y respuesta a conversaciones.</li>
        <li>Permitir que los agentes de la empresa cliente atiendan a los contactos.</li>
        <li>Mantener el historial y el estado de cada conversación.</li>
      </ul>

      <h2>4. Retención</h2>
      <p>
        Los datos se conservan mientras exista una relación con la empresa cliente o
        mientras sean necesarios para la finalidad descrita, y luego se eliminan o
        anonimizan conforme a la ley aplicable.
      </p>

      <h2>5. Terceros y transferencia</h2>
      <p>
        Para enviar y recibir mensajes, los datos pueden compartirse con proveedores de
        canal, en particular <strong>Meta Platforms</strong> en el caso de Messenger, que
        tratan la información bajo sus propias condiciones. También pueden intervenir
        proveedores de infraestructura que alojan el servicio.
      </p>

      <h2>6. Derechos del titular</h2>
      <p>
        Los titulares pueden conocer, actualizar, rectificar y suprimir sus datos
        personales, así como revocar la autorización otorgada para su tratamiento,
        conforme a la Ley 1581 de 2012 y sus normas reglamentarias. Para ejercer estos
        derechos, escribí a
        <a href="mailto:jdgonzalez907@gmail.com">jdgonzalez907@gmail.com</a> indicando
        tu solicitud y los datos que permitan identificarte.
      </p>

      <h2>7. Si sos contacto final</h2>
      <p>
        Si escribiste a una empresa a través de un canal y recibiste atención mediante
        1Channel, la empresa con la que conversaste es la <strong>responsable</strong> del
        tratamiento de tus datos: es quien decide por qué y para qué se tratan. 1Channel
        actúa como <strong>encargado</strong> y los trata siguiendo sus instrucciones. Para
        ejercer tus derechos frente a esa conversación, dirigite en primer lugar a esa
        empresa.
      </p>

      <h2>8. Cambios a esta política</h2>
      <p>
        Esta política puede actualizarse para reflejar cambios del servicio o de la
        normativa. La versión vigente es la publicada en esta página.
      </p>

      <p class="back">
        <RouterLink to="/">Volver al inicio</RouterLink>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<style scoped>
.container {
  width: 100%;
  max-width: 48rem;
  margin: 0 auto;
  padding: 3.5rem 1.5rem 5rem;
}

h1 {
  margin: 0 0 0.25rem;
  font-size: clamp(1.75rem, 4vw, 2.5rem);
  letter-spacing: -0.03em;
}

.meta {
  margin: 0 0 2rem;
  color: var(--fg-muted);
  font-family: var(--font-mono);
  font-size: 0.8rem;
}

h2 {
  margin: 2.25rem 0 0.75rem;
  font-size: 1.15rem;
  letter-spacing: -0.01em;
}

p,
li {
  color: var(--fg);
}

ul {
  margin: 0;
  padding-left: 1.25rem;
}

li {
  margin-bottom: 0.35rem;
}

.back {
  margin-top: 3rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border);
}
</style>
