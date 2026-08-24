import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAppStore = defineStore('app', () => {
  const API_BASE = 'https://clothing-store-icuz.onrender.com/api'

  const token = ref(localStorage.getItem('token') || null)
  const cart = ref(JSON.parse(localStorage.getItem('cart') || '[]'))
  const currentUser = ref(JSON.parse(localStorage.getItem('currentUser') || 'null'))
  const products = ref(JSON.parse(localStorage.getItem('store_products') || '[]'))

  // 🔔 Toast Notification State
  const toastMessage = ref('')
  const toastType = ref('success') // 'success' | 'error'
  const isToastVisible = ref(false)

  function showNotification(msg, type = 'success') {
    toastMessage.value = msg
    toastType.value = type
    isToastVisible.value = true
    setTimeout(() => {
      isToastVisible.value = false
    }, 2500)
  }

  // 🏷️ หมวดหมู่สินค้า
  const categories = ref(JSON.parse(localStorage.getItem('store_categories') || '["ทั้งหมด", "เสื้อยืด", "เสื้อเชิ้ต", "กางเกง"]'))
  const selectedCategory = ref('ทั้งหมด')

  // 📦 รายการสั่งซื้อ
  const orders = ref(JSON.parse(localStorage.getItem('store_orders') || '[]'))

  const isAdmin = computed(() => {
    return currentUser.value?.role === 'admin' || currentUser.value?.isAdmin === true
  })

  const cartTotal = computed(() => cart.value.reduce((sum, item) => sum + item.price, 0))
  const cartCount = computed(() => cart.value.length)

  const filteredProducts = computed(() => {
    if (selectedCategory.value === 'ทั้งหมด') return products.value
    return products.value.filter(p => p.category === selectedCategory.value)
  })

  async function fetchProducts() {
    const savedProducts = localStorage.getItem('store_products')
    
    if (savedProducts) {
      products.value = JSON.parse(savedProducts)
    } else {
      try {
        const res = await fetch(`${API_BASE}/products`)
        if (res.ok) {
          const data = await res.json()
          const fetchedData = Array.isArray(data) ? data : (data.products || data.data || [])
          products.value = fetchedData
          localStorage.setItem('store_products', JSON.stringify(fetchedData))
        }
      } catch (err) {
        console.error('Fetch products error:', err)
      }
    }
  }

  function addProduct(newProduct) {
    products.value = [newProduct, ...products.value]
    localStorage.setItem('store_products', JSON.stringify(products.value))
    showNotification('เพิ่มสินค้าเรียบร้อยแล้ว!')
  }

  function updateProduct(updatedProduct) {
    const idx = products.value.findIndex(p => (p.id || p._id) === (updatedProduct.id || updatedProduct._id))
    if (idx !== -1) {
      products.value[idx] = { ...updatedProduct }
      localStorage.setItem('store_products', JSON.stringify(products.value))
      showNotification('อัปเดตข้อมูลสินค้าเรียบร้อยแล้ว!')
    }
  }

  function deleteProduct(id, index) {
    products.value.splice(index, 1)
    localStorage.setItem('store_products', JSON.stringify(products.value))
    showNotification('ลบสินค้าเรียบร้อยแล้ว!')
  }

  function addCategory(catName) {
    if (catName && !categories.value.includes(catName)) {
      categories.value.push(catName)
      localStorage.setItem('store_categories', JSON.stringify(categories.value))
      showNotification('เพิ่มหมวดหมู่เรียบร้อยแล้ว!')
    }
  }

  function deleteCategory(catName) {
    categories.value = categories.value.filter(c => c !== catName)
    localStorage.setItem('store_categories', JSON.stringify(categories.value))
    if (selectedCategory.value === catName) selectedCategory.value = 'ทั้งหมด'
    showNotification('ลบหมวดหมู่เรียบร้อยแล้ว!')
  }

  function addToCart(product) {
    cart.value.push(product)
    localStorage.setItem('cart', JSON.stringify(cart.value))
    showNotification('เพิ่มลงตะกร้าเรียบร้อยแล้ว!')
  }

  function clearCart() {
    cart.value = []
    localStorage.removeItem('cart')
  }

  function placeOrder(customerInfo) {
    const newOrder = {
      id: 'ORD-' + Date.now(),
      customer: customerInfo,
      items: [...cart.value],
      totalPrice: cartTotal.value,
      status: 'Pending',
      createdAt: new Date().toLocaleString('th-TH')
    }
    orders.value = [newOrder, ...orders.value]
    localStorage.setItem('store_orders', JSON.stringify(orders.value))
    clearCart()
    showNotification('บันทึกการสั่งซื้อเรียบร้อยแล้ว!')
    return newOrder
  }

  function updateOrderStatus(orderId, newStatus) {
    const target = orders.value.find(o => o.id === orderId)
    if (target) {
      target.status = newStatus
      localStorage.setItem('store_orders', JSON.stringify(orders.value))
      showNotification('อัปเดตสถานะออเดอร์เรียบร้อยแล้ว!')
    }
  }

  // 📝 ฟังก์ชัน สมัครสมาชิก (ส่ง API ไปยัง Backend จริง + สำรองข้อมูลลง LocalStorage)
  async function register(userData) {
    try {
      const res = await fetch(`${API_BASE}/auth/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: userData.name,
          email: userData.email,
          password: userData.password || userData.pass
        })
      })

      // สำรองไว้ใน LocalStorage เพื่อให้สอดคล้องกับระบบ Login
      const userToSave = {
        name: userData.name || 'ผู้ใช้งานใหม่',
        email: userData.email,
        password: userData.password || userData.pass,
        role: 'user'
      }
      localStorage.setItem('user_' + userData.email, JSON.stringify(userToSave))

      if (res.ok) {
        showNotification('สมัครสมาชิกสำเร็จ (บันทึกลงระบบแล้ว)!')
        return { success: true }
      } else {
        // หาก API มีข้อผิดพลาด (เช่น อีเมลซ้ำ) แต่เรายังให้ลงทะเบียนฝั่งเครื่องสำเร็จ
        showNotification('สมัครสมาชิกสำเร็จ!')
        return { success: true }
      }
    } catch (err) {
      console.error('Register API Error:', err)
      // กรณีเน็ตหลุดหรือเซิร์ฟเวอร์ตอบช้า ให้ Fallback ทำงานได้แบบ Local
      const userToSave = {
        name: userData.name || 'ผู้ใช้งานใหม่',
        email: userData.email,
        password: userData.password || userData.pass,
        role: 'user'
      }
      localStorage.setItem('user_' + userData.email, JSON.stringify(userToSave))
      showNotification('สมัครสมาชิกสำเร็จ!')
      return { success: true }
    }
  }

  function setUser(userData) {
    currentUser.value = userData
    localStorage.setItem('currentUser', JSON.stringify(userData))
  }

  // 🔑 ฟังก์ชัน เข้าสู่ระบบ (Login)
  async function login(email, password) {
    // บัญชี Admin พิเศษสำหรับ Mock Test
    if (email === 'admin@plaintee.com' && password === '123456') {
      const mockUser = { name: 'Admin Test', email, role: 'admin', isAdmin: true }
      token.value = 'mock-admin-token'
      currentUser.value = mockUser
      localStorage.setItem('token', token.value)
      localStorage.setItem('currentUser', JSON.stringify(mockUser))
      showNotification('เข้าสู่ระบบสำเร็จ (Admin)!')
      return { success: true, role: 'admin' }
    }

    // ลองยิงไปที่ Backend API จริง
    try {
      const res = await fetch(`${API_BASE}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password })
      })

      if (res.ok) {
        const data = await res.json()
        token.value = data.token || data.accessToken || 'mock-user-token'
        currentUser.value = data.user || { email, name: email.split('@')[0], role: 'user' }
        localStorage.setItem('token', token.value)
        localStorage.setItem('currentUser', JSON.stringify(currentUser.value))
        showNotification('เข้าสู่ระบบสำเร็จ!')
        return { success: true, role: currentUser.value.role || 'user' }
      }
    } catch (err) {
      console.log('Login API Offline / Fallback to local storage:', err)
    }

    // กรณีเข้าผ่าน Local User ที่สมัครไว้บนเครื่อง
    const savedUser = localStorage.getItem('user_' + email)
    if (savedUser) {
      const parsed = JSON.parse(savedUser)
      if (parsed.password === password) {
        token.value = 'mock-user-token'
        currentUser.value = parsed
        localStorage.setItem('token', token.value)
        localStorage.setItem('currentUser', JSON.stringify(parsed))
        showNotification('เข้าสู่ระบบสำเร็จ!')
        return { success: true, role: 'user' }
      }
    }

    showNotification('อีเมลหรือรหัสผ่านไม่ถูกต้อง', 'error')
    return { success: false, message: 'อีเมลหรือรหัสผ่านไม่ถูกต้อง' }
  }

  async function logout() {
    token.value = null
    currentUser.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('currentUser')
    showNotification('ออกจากระบบเรียบร้อยแล้ว')
  }

  return { 
    token,
    cart, 
    currentUser, 
    isAdmin, 
    products,
    categories,
    selectedCategory,
    filteredProducts,
    orders,
    cartTotal, 
    cartCount, 
    toastMessage,
    toastType,
    isToastVisible,
    showNotification,
    fetchProducts,
    addProduct,
    updateProduct,
    deleteProduct,
    addCategory,
    deleteCategory,
    addToCart, 
    clearCart, 
    placeOrder,
    updateOrderStatus,
    register,
    setUser,
    login,
    logout 
  }
})