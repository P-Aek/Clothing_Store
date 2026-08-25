import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { apiRequest, getErrorMessage } from '../services/api'

function readStorage(key, fallback) {
  try { return JSON.parse(localStorage.getItem(key) || JSON.stringify(fallback)) } catch { return fallback }
}

function slugify(value) {
  return value.toLowerCase().trim().replace(/\s+/g, '-').replace(/[^\w-ก-๙]/g, '')
}

export const useAppStore = defineStore('app', () => {
  const token = ref(localStorage.getItem('token') || null)
  const guestCart = ref(readStorage('cart', []))
  const cart = ref([])
  const currentUser = ref(readStorage('currentUser', null))
  const products = ref([])
  const categories = ref([])
  const selectedCategory = ref('all')
  const orders = ref([])
  const toastMessage = ref('')
  const toastType = ref('success')
  const isToastVisible = ref(false)
  const isAuthRestored = ref(false)
  const isLoading = ref(false)

  const isAuthenticated = computed(() => Boolean(token.value && currentUser.value))
  const isAdmin = computed(() => currentUser.value?.role === 'admin')
  const cartTotal = computed(() => cart.value.reduce((sum, item) => sum + Number(item.price || 0) * Number(item.quantity || 1), 0))
  const cartCount = computed(() => cart.value.reduce((sum, item) => sum + Number(item.quantity || 1), 0))
  const filteredProducts = computed(() => selectedCategory.value === 'all' ? products.value : products.value.filter(product => product.categoryId === selectedCategory.value))

  function showNotification(message, type = 'success') {
    toastMessage.value = message; toastType.value = type; isToastVisible.value = true
    window.setTimeout(() => { isToastVisible.value = false }, 3000)
  }

  async function fetchProducts() {
    try { products.value = (await apiRequest('/products/'))?.products || [] }
    catch (error) { showNotification(getErrorMessage(error, 'โหลดสินค้าไม่สำเร็จ'), 'error') }
  }

  async function fetchCategories() {
    try { categories.value = (await apiRequest('/categories/'))?.categories || [] }
    catch (error) { showNotification(getErrorMessage(error, 'โหลดหมวดหมู่ไม่สำเร็จ'), 'error') }
  }

  async function addProduct(product) {
    try { await apiRequest('/products/', { method: 'POST', body: JSON.stringify(product) }); await fetchProducts(); showNotification('เพิ่มสินค้าเรียบร้อยแล้ว') }
    catch (error) { showNotification(getErrorMessage(error, 'เพิ่มสินค้าไม่สำเร็จ'), 'error'); throw error }
  }

  async function updateProduct(id, product) {
    try { await apiRequest(`/products/${id}`, { method: 'PUT', body: JSON.stringify(product) }); await fetchProducts(); showNotification('อัปเดตสินค้าเรียบร้อยแล้ว') }
    catch (error) { showNotification(getErrorMessage(error, 'อัปเดตสินค้าไม่สำเร็จ'), 'error'); throw error }
  }

  async function deleteProduct(id) {
    try { await apiRequest(`/products/${id}`, { method: 'DELETE' }); await fetchProducts(); showNotification('ลบสินค้าเรียบร้อยแล้ว') }
    catch (error) { showNotification(getErrorMessage(error, 'ลบสินค้าไม่สำเร็จ'), 'error') }
  }

  async function addCategory(name) {
    const trimmed = name.trim()
    if (!trimmed) throw new Error('กรุณาระบุชื่อหมวดหมู่')
    try {
      await apiRequest('/categories/', { method: 'POST', body: JSON.stringify({ name: trimmed, slug: slugify(trimmed) }) })
      await fetchCategories(); showNotification('เพิ่มหมวดหมู่เรียบร้อยแล้ว')
    } catch (error) { showNotification(getErrorMessage(error, 'เพิ่มหมวดหมู่ไม่สำเร็จ'), 'error'); throw error }
  }

  async function deleteCategory(category) {
    try {
      await apiRequest(`/categories/${category.id}`, { method: 'DELETE' })
      if (selectedCategory.value === category.id) selectedCategory.value = 'all'
      await fetchCategories(); showNotification('ลบหมวดหมู่เรียบร้อยแล้ว')
    } catch (error) { showNotification(getErrorMessage(error, 'ลบหมวดหมู่ไม่สำเร็จ'), 'error') }
  }

  function normalizeCartItem(item) {
    const product = products.value.find(candidate => candidate.id === item.productId)
    const variant = product?.variants?.find(candidate => candidate.id === item.variantId)
    if (!product) return null
    return { ...item, product, name: product.name, price: product.price, image: product.images?.[0], color: variant?.color, size: variant?.size }
  }

  async function loadCart() {
    if (!isAuthenticated.value) { cart.value = [...guestCart.value]; return }
    try { const response = await apiRequest('/cart/'); cart.value = (response?.cart?.items || []).map(normalizeCartItem).filter(Boolean) }
    catch (error) { if (error.status === 401) clearAuth(); showNotification(getErrorMessage(error, 'โหลดตะกร้าไม่สำเร็จ'), 'error'); throw error }
  }

  async function addToCart(product, variant = product.variants?.[0]) {
    if (!variant) { showNotification('สินค้านี้ยังไม่มีตัวเลือกสินค้า', 'error'); return }
    const item = { productId: product.id, variantId: variant.id, quantity: 1, product, price: product.price, name: product.name, image: product.images?.[0], color: variant.color, size: variant.size }
    if (!isAuthenticated.value) {
      const existing = guestCart.value.find(candidate => candidate.productId === item.productId && candidate.variantId === item.variantId)
      if (existing) existing.quantity += 1; else guestCart.value.push(item)
      persistGuestCart(); cart.value = [...guestCart.value]; showNotification('เพิ่มสินค้าลงตะกร้าเรียบร้อยแล้ว'); return
    }
    try { await apiRequest('/cart/items', { method: 'POST', body: JSON.stringify({ productId: item.productId, variantId: item.variantId, quantity: 1 }) }); await loadCart(); showNotification('เพิ่มสินค้าลงตะกร้าเรียบร้อยแล้ว') }
    catch (error) { const message = error.status === 409 ? 'สินค้าตัวเลือกนี้หมดหรือมีจำนวนไม่เพียงพอ' : getErrorMessage(error, 'เพิ่มสินค้าลงตะกร้าไม่สำเร็จ'); showNotification(message, 'error'); if (error.status === 401) clearAuth(); throw error }
  }

  async function updateCartItem(item, quantity) {
    if (quantity < 1) return removeCartItem(item)
    if (!isAuthenticated.value) { item.quantity = quantity; persistGuestCart(); return }
    try { await apiRequest(`/cart/items/${item.productId}/${item.variantId}`, { method: 'PUT', body: JSON.stringify({ quantity }) }); await loadCart() }
    catch (error) { showNotification(getErrorMessage(error, 'อัปเดตตะกร้าไม่สำเร็จ'), 'error'); if (error.status === 401) clearAuth(); throw error }
  }

  async function removeCartItem(item) {
    if (!isAuthenticated.value) {
      guestCart.value = guestCart.value.filter(candidate => !(candidate.productId === item.productId && candidate.variantId === item.variantId))
      persistGuestCart(); cart.value = [...guestCart.value]; return
    }
    try { await apiRequest(`/cart/items/${item.productId}/${item.variantId}`, { method: 'DELETE' }); await loadCart() }
    catch (error) { showNotification(getErrorMessage(error, 'ลบสินค้าออกจากตะกร้าไม่สำเร็จ'), 'error'); if (error.status === 401) clearAuth(); throw error }
  }

  async function clearCart() {
    if (isAuthenticated.value) {
      for (const item of [...cart.value]) {
        try { await apiRequest(`/cart/items/${item.productId}/${item.variantId}`, { method: 'DELETE' }) }
        catch (error) { showNotification(getErrorMessage(error, 'ล้างตะกร้าบางรายการไม่สำเร็จ'), 'error') }
      }
      await loadCart()
      return
    }
    guestCart.value = []; persistGuestCart(); cart.value = []
  }

  async function checkout() {
    try { const response = await apiRequest('/orders/', { method: 'POST' }); await loadCart(); await fetchOrders(); showNotification('สั่งซื้อสินค้าเรียบร้อยแล้ว'); return response?.order }
    catch (error) { showNotification(getErrorMessage(error, 'สั่งซื้อไม่สำเร็จ กรุณาลองใหม่อีกครั้ง'), 'error'); throw error }
  }

  async function fetchOrders(admin = false) {
    try { orders.value = (await apiRequest(admin ? '/admin/orders/' : '/orders/'))?.orders || []; return orders.value }
    catch (error) { showNotification(getErrorMessage(error, 'โหลดรายการสั่งซื้อไม่สำเร็จ'), 'error'); return [] }
  }

  async function updateOrderStatus(orderId, status) {
    try {
      const response = await apiRequest(`/admin/orders/${orderId}/status`, { method: 'PUT', body: JSON.stringify({ status }) })
      const index = orders.value.findIndex(order => order.id === orderId)
      if (index !== -1 && response?.order) orders.value[index] = response.order
      showNotification('อัปเดตสถานะคำสั่งซื้อแล้ว')
    } catch (error) { showNotification(getErrorMessage(error, 'อัปเดตสถานะไม่สำเร็จ'), 'error') }
  }

  async function register(userData) {
    try { await apiRequest('/auth/register', { method: 'POST', body: JSON.stringify({ name: userData.name, email: userData.email, password: userData.password || userData.pass }) }); showNotification('สมัครสมาชิกสำเร็จ'); return { success: true } }
    catch (error) { showNotification(getErrorMessage(error, 'สมัครสมาชิกไม่สำเร็จ'), 'error'); return { success: false, message: getErrorMessage(error) } }
  }

  async function login(email, password) {
    try {
      const guestItems = [...guestCart.value]
      const data = await apiRequest('/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) })
      token.value = data.token; localStorage.setItem('token', token.value); currentUser.value = { ...(data.user || {}), email }
      if (!await restoreSession()) throw new Error('ไม่สามารถยืนยันเซสชันได้')
      await syncGuestCart(guestItems); showNotification('เข้าสู่ระบบสำเร็จ')
      return { success: true, role: currentUser.value.role }
    } catch (error) { clearAuth(); showNotification(getErrorMessage(error, 'อีเมลหรือรหัสผ่านไม่ถูกต้อง'), 'error'); return { success: false, message: getErrorMessage(error) } }
  }

  async function logout() {
    try { if (token.value) await apiRequest('/auth/logout', { method: 'POST' }) } catch { /* local cleanup still runs */ }
    clearAuth(); showNotification('ออกจากระบบเรียบร้อยแล้ว')
  }

  async function restoreSession() {
    if (!token.value) { isAuthRestored.value = true; await loadCart(); return false }
    try {
      const me = await apiRequest('/auth/me')
      currentUser.value = { ...(currentUser.value || {}), id: me.userId || currentUser.value?.id, role: me.role }
      localStorage.setItem('currentUser', JSON.stringify(currentUser.value)); isAuthRestored.value = true; await loadCart(); return true
    } catch { clearAuth(); return false }
  }

  async function syncGuestCart(items = guestCart.value) {
    if (!isAuthenticated.value || !items.length) return
    const failed = []
    for (const item of items) {
      try { await apiRequest('/cart/items', { method: 'POST', body: JSON.stringify({ productId: item.productId, variantId: item.variantId, quantity: item.quantity }) }) }
      catch (error) { failed.push(item); showNotification(error.status === 409 ? 'สินค้าตัวเลือกนี้หมดหรือมีจำนวนไม่เพียงพอ' : getErrorMessage(error, 'ซิงค์ตะกร้าบางรายการไม่สำเร็จ'), 'error') }
    }
    guestCart.value = failed; persistGuestCart(); await loadCart()
  }

  function clearAuth() { token.value = null; currentUser.value = null; isAuthRestored.value = true; localStorage.removeItem('token'); localStorage.removeItem('currentUser'); cart.value = [...guestCart.value] }
  function persistGuestCart() { localStorage.setItem('cart', JSON.stringify(guestCart.value)) }

  async function initialize() {
    isLoading.value = true; await Promise.all([fetchProducts(), fetchCategories()]); await restoreSession(); isLoading.value = false
  }

  return {
    token, cart, guestCart, currentUser, isAdmin, isAuthenticated, isAuthRestored, isLoading,
    products, categories, selectedCategory, filteredProducts, orders, cartTotal, cartCount,
    toastMessage, toastType, isToastVisible, showNotification, fetchProducts, fetchCategories,
    addProduct, updateProduct, deleteProduct, addCategory, deleteCategory, addToCart, loadCart,
    updateCartItem, removeCartItem, clearCart, checkout, fetchOrders, updateOrderStatus,
    register, login, logout, restoreSession, initialize
  }
})
