import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import CartView from '../views/CartView.vue'
import AccountView from '../views/AccountView.vue'
import RegisterView from '../views/RegisterView.vue'
import AdminView from '../views/AdminView.vue'
import { useAppStore } from '../stores/appStore'

const routes = [
  { path: '/', name: 'home', component: HomeView },
  { path: '/cart', name: 'cart', component: CartView },
  { path: '/account', name: 'account', component: AccountView },
  { path: '/register', name: 'register', component: RegisterView },
  { path: '/admin', name: 'admin', component: AdminView, meta: { requiresAdmin: true } }
]

const router = createRouter({ history: createWebHistory(), routes })

router.beforeEach(async (to) => {
  const store = useAppStore()
  if (!store.isAuthRestored) await store.restoreSession()
  if (to.meta.requiresAdmin && !store.isAdmin) return store.isAuthenticated ? '/' : '/account'
  return true
})

export default router
