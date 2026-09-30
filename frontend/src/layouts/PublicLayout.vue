<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'

const year = new Date().getFullYear()
</script>

<template>
  <div class="public-site">
    <header class="site-header">
      <div class="site-header__inner">
        <RouterLink to="/" class="wordmark">1Channel</RouterLink>
        <RouterLink to="/consola" class="header-cta">Entrar a la consola</RouterLink>
      </div>
    </header>

    <main class="site-main">
      <RouterView />
    </main>

    <footer class="site-footer">
      <div class="site-footer__inner">
        <span>1Channel &copy; {{ year }}</span>
        <nav class="footer-links">
          <RouterLink to="/privacidad">Privacidad</RouterLink>
          <RouterLink to="/terminos">Términos</RouterLink>
        </nav>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.header-cta {
  font-size: 0.9rem;
  font-weight: 600;
  text-decoration: none;
}

.header-cta:hover {
  text-decoration: underline;
}
</style>
