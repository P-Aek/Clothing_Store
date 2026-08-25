<template>
  <div class="container account-container">
    <div v-if="store.isAuthenticated" class="card account-card">
      <h2>บัญชีของฉัน</h2>
      <p><strong>ชื่อ:</strong> {{ store.currentUser?.name || '-' }}</p>
      <p><strong>อีเมล:</strong> {{ store.currentUser?.email || '-' }}</p>
      <p><strong>สิทธิ์:</strong> {{ store.isAdmin ? 'ผู้ดูแลระบบ' : 'ลูกค้า' }}</p>
      <button class="btn btn-danger logout" @click="handleLogout">ออกจากระบบ</button>
      <section class="orders-section"><div class="orders-heading"><h3>คำสั่งซื้อของฉัน</h3><button class="btn-refresh" @click="loadOrders">รีเฟรช</button></div><p v-if="!store.orders.length" class="muted">ยังไม่มีคำสั่งซื้อ</p><ul v-else class="orders-list"><li v-for="order in store.orders" :key="order.id"><span>{{ order.id }} · {{ order.status }}</span><strong>฿{{ order.totalPrice }}</strong></li></ul></section>
    </div>
    <div v-else class="card account-card">
      <h2>{{ isLogin ? 'เข้าสู่ระบบ' : 'สมัครสมาชิก' }}</h2>
      <form @submit.prevent="handleSubmit">
        <div v-if="!isLogin" class="form-group"><label>ชื่อ-นามสกุล</label><input v-model.trim="form.name" required autocomplete="name" /></div>
        <div class="form-group"><label>อีเมล</label><input v-model.trim="form.email" type="email" required autocomplete="email" /></div>
        <div class="form-group"><label>รหัสผ่าน</label><input v-model="form.password" type="password" required minlength="8" autocomplete="current-password" /></div>
        <button type="submit" class="btn btn-primary submit" :disabled="loading">{{ loading ? 'กำลังดำเนินการ...' : (isLogin ? 'เข้าสู่ระบบ' : 'ยืนยันการลงทะเบียน') }}</button>
      </form>
      <div class="switch-link"><a href="#" @click.prevent="isLogin = !isLogin">{{ isLogin ? 'ยังไม่มีบัญชี? สมัครสมาชิก' : 'มีบัญชีอยู่แล้ว? เข้าสู่ระบบ' }}</a></div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '../stores/appStore'

const store = useAppStore(); const router = useRouter(); const isLogin = ref(true); const loading = ref(false)
const form = reactive({ name: '', email: '', password: '' })
async function handleSubmit() {
  loading.value = true
  const result = isLogin.value ? await store.login(form.email, form.password) : await store.register(form)
  loading.value = false
  if (result.success) { if (isLogin.value) router.push(store.isAdmin ? '/admin' : '/'); else { isLogin.value = true; form.password = '' } }
}
async function handleLogout() { await store.logout(); router.push('/') }
async function loadOrders() { if (store.isAuthenticated) await store.fetchOrders() }
onMounted(loadOrders)
</script>

<style scoped>
.account-container { max-width: 420px; margin: 40px auto; padding: 0 20px; }.account-card { padding: 30px; }.account-card h2 { text-align: center; margin-top: 0; }.submit, .logout { width: 100%; }.submit:disabled { opacity: .6; cursor: wait; }.switch-link { text-align: center; margin-top: 20px; }.switch-link a { color: #2563eb; text-decoration: none; font-size: .88rem; }.orders-section { margin-top: 28px; border-top: 1px solid #e2e8f0; padding-top: 18px; }.orders-heading { display: flex; justify-content: space-between; align-items: center; }.orders-heading h3 { margin: 0; }.btn-refresh { border: 1px solid #cbd5e1; background: #fff; padding: 6px 10px; border-radius: 6px; cursor: pointer; }.orders-list { padding: 0; list-style: none; }.orders-list li { display: flex; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid #e2e8f0; }.muted { color: #64748b; }
</style>
