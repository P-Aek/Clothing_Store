<template>
  <div class="container cart-container">
    <h2>ตะกร้าสินค้า</h2>
    <div v-if="store.cart.length === 0" class="card empty"><p>ยังไม่มีสินค้าในตะกร้า</p><router-link to="/" class="btn btn-primary">ไปเลือกซื้อสินค้า</router-link></div>
    <div v-else>
      <div v-for="item in store.cart" :key="`${item.productId}-${item.variantId}`" class="card cart-item-card">
        <div class="cart-item-left"><img v-if="item.image" :src="item.image" :alt="item.name" class="cart-img" /><div v-else class="cart-img-placeholder">👕</div><div><h3>{{ item.name }}</h3><p>ไซส์: {{ item.size || '-' }} | สี: {{ item.color || '-' }}</p></div></div>
        <div class="quantity"><button @click="store.updateCartItem(item, item.quantity - 1)">−</button><span>{{ item.quantity }}</span><button @click="store.updateCartItem(item, item.quantity + 1)">+</button></div>
        <strong class="cart-price">฿{{ (Number(item.price) * Number(item.quantity)).toFixed(2) }}</strong>
        <button class="remove" @click="store.removeCartItem(item)">ลบ</button>
      </div>
      <div class="card summary"><h3>ราคารวมทั้งหมด: <span>฿{{ store.cartTotal.toFixed(2) }}</span></h3><div class="actions"><button class="btn btn-primary" :disabled="loading" @click="checkout">สั่งซื้อสินค้า</button><button class="btn btn-danger" @click="store.clearCart">ล้างตะกร้า</button></div></div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '../stores/appStore'
const store = useAppStore(); const router = useRouter(); const loading = ref(false)
async function checkout() {
  if (!store.isAuthenticated) { store.showNotification('กรุณาเข้าสู่ระบบก่อนทำการสั่งซื้อ', 'error'); router.push('/account'); return }
  loading.value = true
  try { await store.checkout(); router.push('/') } catch { /* store displays the API error */ } finally { loading.value = false }
}
</script>

<style scoped>
.cart-container { max-width: 800px; margin: 40px auto; padding: 0 20px; }.empty { text-align: center; padding: 40px; }.cart-item-card { margin-bottom: 12px; display: flex; gap: 16px; justify-content: space-between; align-items: center; padding: 12px 20px; }.cart-item-left { display: flex; align-items: center; gap: 16px; flex: 1; }.cart-item-left h3 { margin: 0; font-size: 1.05rem; }.cart-item-left p { margin: 4px 0 0; color: #64748b; font-size: .85rem; }.cart-img, .cart-img-placeholder { width: 56px; height: 56px; object-fit: cover; border-radius: 6px; }.cart-img-placeholder { background: #f1f5f9; display: flex; align-items: center; justify-content: center; font-size: 1.5rem; }.quantity { display: flex; gap: 8px; align-items: center; }.quantity button { width: 28px; height: 28px; border: 1px solid #cbd5e1; background: #fff; border-radius: 5px; cursor: pointer; }.cart-price { min-width: 90px; text-align: right; }.remove { border: 0; background: none; color: #dc2626; cursor: pointer; }.summary { margin-top: 20px; background: #f8fafc; padding: 20px; }.summary span { color: #2563eb; }.actions { margin-top: 15px; display: flex; gap: 10px; }
</style>
