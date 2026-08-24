<template>
  <div class="container" style="max-width: 800px; margin: 40px auto; padding: 0 20px;">
    <h2>ตะกร้าสินค้า</h2>

    <div v-if="store.cart.length === 0" class="card" style="text-align: center; padding: 40px;">
      <p>ยังไม่มีสินค้าในตะกร้า</p>
      <router-link to="/" class="btn btn-primary" style="display: inline-block; margin-top: 10px; text-decoration: none;">ไปเลือกซื้อสินค้า</router-link>
    </div>

    <div v-else>
      <div v-for="(item, index) in store.cart" :key="index" class="card cart-item-card">
        <div class="cart-item-left">
          <img v-if="item.image" :src="item.image" alt="thumb" class="cart-img" />
          <div v-else class="cart-img-placeholder">👕</div>
          <div>
            <h3 style="margin: 0; font-size: 1.05rem;">{{ item.name }}</h3>
            <p style="margin: 4px 0 0 0; color: #64748b; font-size: 0.85rem;">ไซส์: {{ item.size || '-' }} | สี: {{ item.color || '-' }}</p>
          </div>
        </div>
        <p class="cart-price">฿{{ item.price }}</p>
      </div>

      <div class="card" style="margin-top: 20px; background: #f8fafc; padding: 20px;">
        <h3>ราคารวมทั้งหมด: <span style="color: #2563eb;">฿{{ store.cartTotal }}</span></h3>
        
        <div style="margin-top: 15px; display: flex; gap: 10px;">
          <button class="btn btn-primary" @click="checkout">สั่งซื้อสินค้า</button>
          <button class="btn btn-danger" @click="store.clearCart()">ล้างตะกร้า</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useAppStore } from '../stores/appStore'
import { useRouter } from 'vue-router'

const store = useAppStore()
const router = useRouter()

const checkout = () => {
  if (!store.currentUser) {
    store.showNotification('กรุณาเข้าสู่ระบบหรือสมัครสมาชิกก่อนทำการสั่งซื้อ', 'error')
    router.push('/account')
    return
  }

  // 📦 บันทึกคำสั่งซื้อลง Store (ใน placeOrder จะเรียก Toast แจ้งเตือนให้อัตโนมัติ)
  store.placeOrder({
    name: store.currentUser.name || 'ลูกค้าสมาชิก',
    email: store.currentUser.email,
    phone: store.currentUser.phone || '08X-XXX-XXXX'
  })

  router.push('/') // 🛍️ พาลูกค้ากลับหน้าร้านค้า
}
</script>

<style scoped>
.cart-item-card { margin-bottom: 12px; display: flex; justify-content: space-between; align-items: center; padding: 12px 20px; background: white; border-radius: 8px; border: 1px solid #e2e8f0; }
.cart-item-left { display: flex; align-items: center; gap: 16px; }
.cart-img { width: 56px; height: 56px; object-fit: cover; border-radius: 6px; }
.cart-img-placeholder { width: 56px; height: 56px; background: #f1f5f9; border-radius: 6px; display: flex; align-items: center; justify-content: center; font-size: 1.5rem; }
.cart-price { font-weight: 700; font-size: 1.1rem; color: #0f172a; margin: 0; }
.btn { padding: 10px 18px; border-radius: 6px; font-weight: 600; cursor: pointer; border: none; }
.btn-primary { background: #2563eb; color: white; }
.btn-danger { background: #ef4444; color: white; }
</style>