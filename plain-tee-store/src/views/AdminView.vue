<template>
  <div class="admin-wrapper">
    <header class="admin-header">
      <div class="admin-header-content">
        <div class="admin-brand">
          <span class="shield-icon">🛡️</span>
          <div>
            <h1>Admin Control Panel</h1>
            <p class="admin-subtitle">ระบบจัดการสต็อก หมวดหมู่ และรายการสั่งซื้อ Plain Tee Store</p>
          </div>
        </div>
        <div class="header-actions">
          <router-link to="/" class="btn btn-outline-light">← กลับไปหน้าร้านค้า</router-link>
          <button @click="handleLogout" class="btn btn-danger-sm">ออกจากระบบ</button>
        </div>
      </div>
    </header>

    <main class="admin-container">
      <!-- 📌 ปุ่มสลับโหมด: จัดการสินค้า / จัดการออเดอร์ -->
      <div class="tab-menu mb-4">
        <button class="tab-btn" :class="{ active: activeTab === 'products' }" @click="activeTab = 'products'">📦 จัดการสินค้า & หมวดหมู่</button>
        <button class="tab-btn" :class="{ active: activeTab === 'orders' }" @click="activeTab = 'orders'">📑 รายการสั่งซื้อ (Orders) [{{ store.orders.length }}]</button>
      </div>

      <!-- ---------------- TAB 1: จัดการสินค้า & หมวดหมู่ ---------------- -->
      <div v-if="activeTab === 'products'" class="admin-grid">
        <div class="left-col">
          <!-- CRUD Category -->
          <section class="card admin-card mb-4">
            <h2>🏷️ จัดการหมวดหมู่ (Category)</h2>
            <div class="admin-form">
              <div class="form-group row-group">
                <input type="text" v-model="newCatName" placeholder="ชื่อหมวดหมู่ใหม่..." />
                <button type="button" class="btn btn-dark" @click="handleAddCategory">เพิ่ม</button>
              </div>
              <div class="cat-tags">
                <span v-for="cat in store.categories" :key="cat" class="cat-chip">
                  {{ cat }}
                  <button v-if="cat !== 'ทั้งหมด'" @click="store.deleteCategory(cat)" class="cat-del">×</button>
                </span>
              </div>
            </div>
          </section>

          <!-- ฟอร์ม เพิ่ม / แก้ไข สินค้า -->
          <section class="card admin-card">
            <h2>{{ isEditing ? '✏️ แก้ไขสินค้า' : '➕ เพิ่มสินค้าใหม่' }}</h2>
            <div class="admin-form">
              <div class="form-group">
                <label>ชื่อสินค้า</label>
                <input type="text" v-model="form.name" placeholder="เช่น Plain Tee - Oversized" />
              </div>

              <div class="form-group">
                <label>หมวดหมู่</label>
                <select v-model="form.category" class="form-select">
                  <option v-for="cat in store.categories.filter(c => c !== 'ทั้งหมด')" :key="cat" :value="cat">
                    {{ cat }}
                  </option>
                </select>
              </div>

              <div class="form-group">
                <label>URL รูปภาพสินค้า</label>
                <input type="text" v-model="form.image" placeholder="https://example.com/image.jpg" />
              </div>

              <div class="form-row">
                <div class="form-group">
                  <label>ราคา (บาท)</label>
                  <input type="number" v-model="form.price" placeholder="290" />
                </div>
                <div class="form-group">
                  <label>ไซส์</label>
                  <select v-model="form.size" class="form-select">
                    <option value="S">S</option>
                    <option value="M">M</option>
                    <option value="L">L</option>
                    <option value="XL">XL</option>
                    <option value="Free Size">Free Size</option>
                  </select>
                </div>
              </div>

              <button type="button" class="btn btn-dark btn-block" @click="handleSubmitProduct">
                {{ isEditing ? 'อัปเดตข้อมูลสินค้า' : 'บันทึกสินค้าลงระบบ' }}
              </button>
              <button v-if="isEditing" type="button" class="btn btn-outline btn-block" @click="resetForm" style="margin-top: 8px;">
                ยกเลิกการแก้ไข
              </button>
            </div>
          </section>
        </div>

        <!-- ตารางสินค้า -->
        <section class="card admin-card">
          <div class="section-title-row">
            <h2>📦 รายการสินค้าในระบบ</h2>
            <button class="btn btn-sm btn-outline" @click="store.fetchProducts">🔄 รีเฟรช</button>
          </div>

          <div v-if="store.products.length === 0" class="state-msg">ยังไม่มีรายการสินค้า</div>
          <div v-else class="table-responsive">
            <table class="admin-table">
              <thead>
                <tr>
                  <th>รูปภาพ</th>
                  <th>ชื่อสินค้า</th>
                  <th>หมวดหมู่</th>
                  <th>ไซส์</th>
                  <th>ราคา</th>
                  <th style="text-align: center;">จัดการ</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(item, index) in store.products" :key="item.id || item._id || index">
                  <td>
                    <img v-if="item.image" :src="item.image" alt="preview" class="admin-img-thumb" />
                    <div v-else class="admin-img-placeholder">ไม่มีรูป</div>
                  </td>
                  <td class="font-bold">{{ item.name }}</td>
                  <td><span class="tag cat-tag">{{ item.category || 'ทั่วไป' }}</span></td>
                  <td><span class="tag">{{ item.size || '-' }}</span></td>
                  <td class="price-text">฿{{ item.price }}</td>
                  <td style="text-align: center; gap: 6px; display: flex; justify-content: center;">
                    <button type="button" class="btn-edit" @click="startEdit(item)">แก้ไข</button>
                    <button type="button" class="btn-delete" @click="store.deleteProduct(item.id || item._id, index)">ลบ</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </div>

      <!-- ---------------- TAB 2: ดู Order & อัปเดตสถานะ ---------------- -->
      <div v-else-if="activeTab === 'orders'" class="card admin-card">
        <h2>📑 รายการสั่งซื้อของลูกค้า</h2>
        <div v-if="store.orders.length === 0" class="state-msg">ยังไม่มีรายการสั่งซื้อเข้ามา</div>
        <div v-else class="table-responsive">
          <table class="admin-table">
            <thead>
              <tr>
                <th>เลขที่ Order</th>
                <th>วันที่สั่งซื้อ</th>
                <th>ลูกค้า</th>
                <th>รายการสินค้า</th>
                <th>ยอดรวม</th>
                <th>สถานะ</th>
                <th>อัปเดตสถานะ</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="ord in store.orders" :key="ord.id">
                <td class="font-bold">{{ ord.id }}</td>
                <td style="font-size: 0.8rem; color: #64748b;">{{ ord.createdAt }}</td>
                <td>
                  <div><strong>{{ ord.customer?.name || 'ลูกค้าทั่วไป' }}</strong></div>
                  <div style="font-size: 0.8rem; color: #64748b;">{{ ord.customer?.phone || '-' }}</div>
                </td>
                <td>
                  <ul style="padding-left: 16px; margin: 0; font-size: 0.85rem;">
                    <li v-for="(it, i) in ord.items" :key="i">{{ it.name }} (฿{{ it.price }})</li>
                  </ul>
                </td>
                <td class="price-text">฿{{ ord.totalPrice }}</td>
                <td>
                  <span class="status-badge" :class="ord.status.toLowerCase()">{{ ord.status }}</span>
                </td>
                <td>
                  <select :value="ord.status" @change="e => store.updateOrderStatus(ord.id, e.target.value)" class="form-select-sm">
                    <option value="Pending">Pending (รอดำเนินการ)</option>
                    <option value="Processing">Processing (กำลังจัดส่ง)</option>
                    <option value="Completed">Completed (สำเร็จ)</option>
                    <option value="Cancelled">Cancelled (ยกเลิก)</option>
                  </select>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

    </main>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '../stores/appStore'

const store = useAppStore()
const router = useRouter()

const activeTab = ref('products')
const newCatName = ref('')
const isEditing = ref(false)
const currentEditId = ref(null)

const form = reactive({
  name: '',
  category: 'เสื้อยืด',
  image: '',
  price: '',
  size: 'Free Size'
})

const handleAddCategory = () => {
  if (!newCatName.value.trim()) return
  store.addCategory(newCatName.value.trim())
  newCatName.value = ''
}

const startEdit = (item) => {
  isEditing.value = true
  currentEditId.value = item.id || item._id
  form.name = item.name
  form.category = item.category || 'เสื้อยืด'
  form.image = item.image
  form.price = item.price
  form.size = item.size || 'Free Size'
}

const resetForm = () => {
  isEditing.value = false
  currentEditId.value = null
  form.name = ''
  form.image = ''
  form.price = ''
  form.size = 'Free Size'
}

const handleSubmitProduct = () => {
  if (!form.name || !form.price) {
    alert('กรุณากรอกชื่อสินค้าและราคาให้ครบถ้วน')
    return
  }

  if (isEditing.value) {
    store.updateProduct({
      id: currentEditId.value,
      name: form.name,
      category: form.category,
      price: Number(form.price),
      size: form.size,
      image: form.image
    })
    resetForm()
  } else {
    store.addProduct({
      id: Date.now(),
      name: form.name,
      category: form.category,
      price: Number(form.price),
      size: form.size,
      image: form.image || 'https://images.unsplash.com/photo-1521572267360-ee0c2909d518?w=500'
    })
    resetForm()
  }
}

const handleLogout = async () => {
  await store.logout()
  alert('ออกจากระบบเรียบร้อยแล้ว')
  router.push('/')
}

onMounted(() => {
  store.fetchProducts()
})
</script>

<style scoped>
.admin-wrapper { min-height: 100vh; background-color: #f1f5f9; }
.admin-header { background-color: #0f172a; color: #ffffff; padding: 24px 0; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); }
.admin-header-content { max-width: 1100px; margin: 0 auto; padding: 0 20px; display: flex; justify-content: space-between; align-items: center; }
.header-actions { display: flex; align-items: center; gap: 12px; }
.admin-brand { display: flex; align-items: center; gap: 16px; }
.shield-icon { font-size: 2.2rem; }
.admin-header h1 { font-size: 1.4rem; font-weight: 700; margin: 0; }
.admin-subtitle { font-size: 0.85rem; color: #94a3b8; margin: 4px 0 0 0; }
.admin-container { max-width: 1100px; margin: 32px auto; padding: 0 20px; }
.tab-menu { display: flex; gap: 12px; }
.tab-btn { padding: 10px 20px; border-radius: 8px; border: 1px solid #cbd5e1; background: white; cursor: pointer; font-weight: 600; color: #475569; }
.tab-btn.active { background: #0f172a; color: white; border-color: #0f172a; }
.admin-grid { display: grid; grid-template-columns: 380px 1fr; gap: 24px; align-items: start; }
.left-col { display: flex; flex-direction: column; gap: 20px; }
.mb-4 { margin-bottom: 16px; }
.admin-card { background: white; padding: 24px; border-radius: 12px; box-shadow: 0 1px 3px rgba(0,0,0,0.05); }
.admin-card h2 { font-size: 1.1rem; margin-top: 0; margin-bottom: 16px; color: #0f172a; }
.form-group { margin-bottom: 14px; }
.row-group { display: flex; gap: 8px; }
.row-group input { flex: 1; }
.form-group label { display: block; font-size: 0.85rem; font-weight: 600; color: #475569; margin-bottom: 6px; }
.form-group input { width: 100%; padding: 10px 12px; border: 1px solid #cbd5e1; border-radius: 8px; font-size: 0.9rem; box-sizing: border-box; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.form-select, .form-select-sm { width: 100%; padding: 8px 10px; border: 1px solid #cbd5e1; border-radius: 8px; font-size: 0.85rem; background: white; box-sizing: border-box; }
.btn-block { width: 100%; margin-top: 10px; cursor: pointer; padding: 12px; font-weight: 600; background: #0f172a; color: white; border: none; border-radius: 8px; }
.btn-outline-light { border: 1px solid #334155; color: #f8fafc; background: transparent; padding: 8px 16px; border-radius: 6px; text-decoration: none; font-size: 0.875rem; }
.btn-danger-sm { background: #ef4444; color: white; border: none; padding: 8px 14px; border-radius: 6px; cursor: pointer; font-size: 0.875rem; font-weight: 600; }
.section-title-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.cat-tags { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 12px; }
.cat-chip { background: #f1f5f9; padding: 4px 10px; border-radius: 16px; font-size: 0.8rem; color: #334155; display: inline-flex; align-items: center; gap: 6px; }
.cat-del { border: none; background: none; color: #ef4444; font-weight: bold; cursor: pointer; padding: 0; }
.admin-table { width: 100%; border-collapse: collapse; text-align: left; font-size: 0.88rem; }
.admin-table th { padding: 12px; background: #f8fafc; color: #475569; border-bottom: 2px solid #e2e8f0; font-weight: 600; }
.admin-table td { padding: 12px; border-bottom: 1px solid #e2e8f0; color: #334155; vertical-align: middle; }
.font-bold { font-weight: 600; color: #0f172a; }
.price-text { font-weight: 700; color: #2563eb; }
.tag { background: #f1f5f9; padding: 2px 8px; border-radius: 4px; font-size: 0.8rem; color: #475569; }
.cat-tag { background: #e0f2fe; color: #0369a1; }
.btn-edit { background: #e0e7ff; color: #3730a3; border: none; padding: 6px 12px; border-radius: 4px; cursor: pointer; font-size: 0.8rem; font-weight: 600; }
.btn-delete { background: #fee2e2; color: #dc2626; border: none; padding: 6px 12px; border-radius: 4px; cursor: pointer; font-size: 0.8rem; font-weight: 600; }
.status-badge { padding: 4px 8px; border-radius: 4px; font-size: 0.75rem; font-weight: 600; }
.status-badge.pending { background: #fef3c7; color: #92400e; }
.status-badge.processing { background: #dbeafe; color: #1e40af; }
.status-badge.completed { background: #dcfce7; color: #166534; }
.status-badge.cancelled { background: #fee2e2; color: #991b1b; }
.state-msg { text-align: center; padding: 40px; color: #94a3b8; }
.admin-img-thumb { width: 44px; height: 44px; object-fit: cover; border-radius: 6px; border: 1px solid #e2e8f0; }
.admin-img-placeholder { width: 44px; height: 44px; background: #f1f5f9; border-radius: 6px; display: flex; align-items: center; justify-content: center; font-size: 0.65rem; color: #94a3b8; }

@media (max-width: 868px) {
  .admin-grid { grid-template-columns: 1fr; }
}
</style>