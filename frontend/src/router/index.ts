import { createRouter, createWebHistory } from 'vue-router'
import PublicLayout from '@/layouts/PublicLayout.vue'
import LandingView from '@/views/LandingView.vue'
import AgentConsoleView from '@/views/AgentConsoleView.vue'
import PrivacyView from '@/views/PrivacyView.vue'
import TermsView from '@/views/TermsView.vue'

declare module 'vue-router' {
  interface RouteMeta {
    title: string
  }
}

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  scrollBehavior: () => ({ top: 0 }),
  routes: [
    {
      path: '/',
      component: PublicLayout,
      children: [
        {
          path: '',
          name: 'landing',
          component: LandingView,
          meta: { title: '1Channel · CRM conversacional multicanal' },
        },
        {
          path: 'privacidad',
          name: 'privacy',
          component: PrivacyView,
          meta: { title: 'Política de privacidad · 1Channel' },
        },
        {
          path: 'terminos',
          name: 'terms',
          component: TermsView,
          meta: { title: 'Términos y condiciones · 1Channel' },
        },
      ],
    },
    {
      path: '/consola',
      name: 'console',
      component: AgentConsoleView,
      meta: { title: 'Consola · 1Channel' },
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

router.afterEach((to) => {
  if (to.meta.title) document.title = to.meta.title
})

export default router
