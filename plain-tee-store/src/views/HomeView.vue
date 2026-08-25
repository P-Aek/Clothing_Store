<template>
  <div class="container" style="max-width: 1100px; margin: 30px auto; padding: 0 20px;">
    <h2>คอลเลกชันสินค้า</h2>
    <p style="color: #64748b; margin-bottom: 20px;">เสื้อยืดเนื้อผ้าคุณภาพ เรียบง่าย เหมาะกับทุกวัน</p>
    <div class="category-bar">
      <button class="cat-btn" :class="{ active: store.selectedCategory === 'all' }" @click="store.selectedCategory = 'all'">ทั้งหมด</button>
      <button v-for="category in store.categories" :key="category.id" class="cat-btn" :class="{ active: store.selectedCategory === category.id }" @click="store.selectedCategory = category.id">{{ category.name }}</button>
    </div>
    <div v-if="store.filteredProducts.length === 0" style="text-align: center; padding: 40px; color: #94a3b8;">ไม่พบรายการสินค้าในหมวดหมู่นี้</div>
    <div v-else class="product-grid">
      <div v-for="item in store.filteredProducts" :key="item.id" class="product-card">
        <div class="img-container"><img :src="item.images?.[0] || fallbackImage" :alt="item.name" /></div>
        <div class="product-info">
          <h3>{{ item.name }}</h3>
          <p class="specs">{{ item.description }}<br>มี {{ item.variants?.length || 0 }} ตัวเลือก</p>
          <div class="price-row"><span class="price">฿{{ item.price }}</span><button class="btn btn-dark" :disabled="!item.active" @click="handleAddToCart(item)">{{ item.active ? '+ เพิ่มลงตะกร้า' : 'สินค้าหมด' }}</button></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useAppStore } from '../stores/appStore'
const store = useAppStore()
const fallbackImage = 'https://images.unsplash.com/photo-1521572267360-ee0c2909d518?w=500'
async function handleAddToCart(product) {
  try { await store.addToCart(product) } catch { /* the store displays the business/API error */ }
}
onMounted(() => { if (!store.products.length) store.fetchProducts(); if (!store.categories.length) store.fetchCategories() })
</script>

<style scoped>
.category-bar { display: flex; gap: 10px; margin-bottom: 24px; flex-wrap: wrap; }
.cat-btn { padding: 8px 16px; border-radius: 20px; border: 1px solid #cbd5e1; background: white; color: #475569; cursor: pointer; font-size: .9rem; transition: all .2s; }
.cat-btn:hover, .cat-btn.active { background: #0f172a; color: white; border-color: #0f172a; font-weight: 600; }
.product-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 24px; }
.product-card { background: #fff; border: 1px solid #e2e8f0; border-radius: 12px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,.05); display: flex; flex-direction: column; }
.img-container { width: 100%; height: 220px; background: #f8fafc; }
.img-container img { width: 100%; height: 100%; object-fit: cover; }
.product-info { padding: 16px; display: flex; flex-direction: column; flex: 1; }
.product-info h3 { font-size: 1.05rem; margin: 0 0 6px; color: #0f172a; }
.specs { font-size: .85rem; color: #64748b; margin-bottom: 16px; }
.price-row { margin-top: auto; display: flex; justify-content: space-between; align-items: center; gap: 8px; }
.price { font-size: 1.15rem; font-weight: 700; color: #0f172a; }
.btn-dark { background: #0f172a; color: white; border: none; padding: 8px 12px; border-radius: 6px; cursor: pointer; font-size: .85rem; font-weight: 600; }
.btn-dark:disabled { opacity: .5; cursor: not-allowed; }
</style>
