<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <div class="landing">
    <section class="hero">
      <div class="container hero__inner">
        <p class="eyebrow">CRM conversacional multicanal</p>
        <h1>Una bandeja. Varios agentes. Todos los canales.</h1>
        <p class="lead">
         1Channel centraliza las conversaciones de tus clientes en una sola bandeja,
          las reparte entre tu equipo de agentes y mantiene cada hilo en orden, sin
          importar el canal por el que escriba el contacto.
        </p>
        <div class="actions">
          <RouterLink to="/consola" class="btn btn--primary">Entrar a la consola</RouterLink>
          <a href="#capacidades" class="btn btn--ghost">Ver capacidades</a>
        </div>
      </div>
    </section>

    <section class="section" id="como-funciona">
      <div class="container">
        <h2 class="section__title">Cómo funciona</h2>
        <ol class="steps">
          <li class="step">
            <span class="step__index">01</span>
            <h3>El contacto escribe</h3>
            <p>El cliente inicia la conversación por un canal como Messenger.</p>
          </li>
          <li class="step">
            <span class="step__index">02</span>
            <h3>Cae en la bandeja</h3>
            <p>El mensaje entra a una bandeja unificada, con su historial y su estado.</p>
          </li>
          <li class="step">
            <span class="step__index">03</span>
            <h3>Un agente responde</h3>
            <p>Un agente toma la conversación, responde y la resuelve cuando termina.</p>
          </li>
        </ol>
        <div class="flow" aria-hidden="true">
          <span>contacto</span>
          <span class="flow__arrow">-&gt;</span>
          <span>canal</span>
          <span class="flow__arrow">-&gt;</span>
          <span>bandeja unificada</span>
          <span class="flow__arrow">-&gt;</span>
          <span>agente</span>
        </div>
      </div>
    </section>

    <section class="section section--soft" id="capacidades">
      <div class="container">
        <h2 class="section__title">Capacidades</h2>
        <div class="grid">
          <article class="card">
            <h3>Bandeja unificada</h3>
            <p>Conversaciones pendientes, asignadas y resueltas en una sola vista, con no leídos y último mensaje.</p>
          </article>
          <article class="card">
            <h3>Varios agentes</h3>
            <p>Cada conversación tiene un solo agente asignado; el equipo se reparte el trabajo sin pisarse.</p>
          </article>
          <article class="card">
            <h3>Canales modulares</h3>
            <p>Messenger hoy; la arquitectura admite sumar más canales sin cambiar la bandeja.</p>
          </article>
        </div>
      </div>
    </section>

    <section class="section closing">
      <div class="container closing__inner">
        <h2>Tu equipo, listo para conversar.</h2>
        <RouterLink to="/consola" class="btn btn--primary">Entrar a la consola</RouterLink>
      </div>
    </section>
  </div>
</template>

<style scoped>
.container {
  width: 100%;
  max-width: var(--maxw);
  margin: 0 auto;
  padding: 0 1.5rem;
}

.hero__inner {
  padding-top: 5rem;
  padding-bottom: 5rem;
  max-width: 52rem;
}

.eyebrow {
  margin: 0 0 1rem;
  color: var(--accent);
  font-family: var(--font-mono);
  font-size: 0.8rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

h1 {
  margin: 0 0 1.25rem;
  font-size: clamp(2rem, 5vw, 3.25rem);
  line-height: 1.1;
  letter-spacing: -0.03em;
}

.lead {
  margin: 0 0 2rem;
  max-width: 44rem;
  color: var(--fg-muted);
  font-size: 1.125rem;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.section {
  padding: 4rem 0;
  border-top: 1px solid var(--border);
}

.section--soft {
  background: var(--bg-soft);
}

.section__title {
  margin: 0 0 2rem;
  font-size: 1.5rem;
  letter-spacing: -0.02em;
}

.section__title::before {
  content: '';
  display: inline-block;
  width: 1.5rem;
  height: 2px;
  margin-right: 0.75rem;
  background: var(--accent);
  vertical-align: middle;
}

.steps {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(14rem, 1fr));
  gap: 1.5rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.step {
  border-top: 1px solid var(--border);
  padding-top: 1rem;
}

.step__index {
  color: var(--accent);
  font-family: var(--font-mono);
  font-size: 0.8rem;
}

.step h3 {
  margin: 0.5rem 0;
  font-size: 1.05rem;
}

.step p,
.card p {
  margin: 0;
  color: var(--fg-muted);
  font-size: 0.95rem;
}

.flow {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
  margin-top: 2.5rem;
  color: var(--fg-muted);
  font-family: var(--font-mono);
  font-size: 0.8rem;
}

.flow__arrow {
  color: var(--accent);
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(16rem, 1fr));
  gap: 1.5rem;
}

.card {
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 1.5rem;
  background: var(--bg);
}

.card h3 {
  margin: 0 0 0.5rem;
  font-size: 1.05rem;
}

.closing__inner {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 1.5rem;
}

.closing h2 {
  margin: 0;
  font-size: 1.5rem;
  letter-spacing: -0.02em;
}
</style>
