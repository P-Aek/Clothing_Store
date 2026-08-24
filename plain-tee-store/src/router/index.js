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
  { 
    path: '/admin', 
    name: 'admin', 
    component: AdminView,
    meta: { requiresAdmin: true } // 🔒 กำหนดสิทธิ์เฉพาะ Admin เท่านั้น
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 🛡️ Navigation Guard (ตรวจสิทธิ์ก่อนเปลี่ยนหน้า)
router.beforeEach((to, from, next) => {
  const store = useAppStore()

  if (to.meta.requiresAdmin) {
    if (!store.isAdmin) {
      alert('คุณไม่มีสิทธิ์เข้าถึงหน้า Admin Control Panel')
      next('/account') // ดีดกลับไปหน้าเข้าสู่ระบบทันที
    } else {
      next()
    }
  } else {
    next()
  }
})

export default router