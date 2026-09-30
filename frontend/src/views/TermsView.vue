<template>
  <div class="legal">
    <div class="container">
      <h1>Términos y condiciones</h1>
      <p class="meta">Última actualización: septiembre de 2026</p>

      <p class="legal-notice">
        Borrador pendiente de revisión legal. Este texto es orientativo y no constituye
        asesoría jurídica. Debe ser revisado por un profesional del derecho antes de
        considerarse definitivo.
      </p>

      <h2>1. Descripción del servicio</h2>
      <p>
        1Channel es un servicio de software que ofrece una bandeja unificada para
        gestionar conversaciones de clientes en distintos canales, asignarlas a agentes
        y mantener su historial. El servicio se presta a empresas ("cliente") que lo
        utilizan para atender a sus propios contactos.
      </p>

      <h2>2. Uso aceptable</h2>
      <p>
        El cliente se compromete a usar el servicio conforme a la ley y a no emplearlo
        para enviar contenido ilícito, abusivo, engañoso o no solicitado, ni para
        vulnerar derechos de terceros. También se compromete a cumplir las condiciones
        de los canales conectados, como Messenger.
      </p>

      <h2>3. Cuentas de agente y acceso</h2>
      <p>
        El acceso a la consola se realiza mediante un identificador de agente. El cliente
        es responsable de las credenciales que entrega a su equipo y de la actividad que
        se realice con ellas. Cada conversación admite un único agente asignado.
      </p>

      <h2>4. Propiedad y uso de los datos</h2>
      <p>
        Los datos de los contactos y el contenido de las conversaciones pertenecen al
        cliente. 1Channel los trata como encargado y solo para prestar el servicio,
        conforme a la política de privacidad. El cliente declara contar con las
        autorizaciones necesarias para el tratamiento de esos datos.
      </p>

      <h2>5. Disponibilidad</h2>
      <p>
        El servicio se ofrece tal como está y puede sufrir interrupciones por
        mantenimiento, fallas de terceros (como los canales o la infraestructura) o
        causas de fuerza mayor. No se garantiza una disponibilidad ininterrumpida.
      </p>

      <h2>6. Terminación</h2>
      <p>
        El acceso puede suspenderse o darse por terminado por incumplimiento de estos
        términos o por decisión de cualquiera de las partes. Al terminar, se aplicarán
        las reglas de retención y eliminación de la política de privacidad.
      </p>

      <h2>7. Ley aplicable</h2>
      <p>
        Estos términos se rigen por las leyes de <strong>Colombia</strong>. Cualquier
        controversia se someterá a la jurisdicción competente en Colombia.
      </p>

      <h2>8. Contacto</h2>
      <p>
        Para consultas sobre estos términos, escribí a
        <a href="mailto:jdgonzalez907@gmail.com">jdgonzalez907@gmail.com</a>.
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

p {
  color: var(--fg);
}

.back {
  margin-top: 3rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border);
}
</style>
