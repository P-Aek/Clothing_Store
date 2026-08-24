<template>
  <header class="navbar-header">
    <div class="nav-container">
      <router-link to="/" class="brand-logo">
        <span class="logo-icon">👕</span> PLAIN TEE
      </router-link>
      <nav class="nav-links">
        <router-link to="/" class="nav-item">หน้าแรก</router-link>
        <router-link to="/cart" class="nav-item cart-badge-wrapper">
          ตะกร้าสินค้า
          <span v-if="store.cartCount > 0" class="cart-badge">{{ store.cartCount }}</span>
        </router-link>
        <router-link to="/account" class="nav-item">บัญชีของฉัน</router-link>
      </nav>
    </div>
  </header>
</template>

<script setup>
import { useAppStore } from '../stores/appStore'
const store = useAppStore()
</script>

<style scoped>
.navbar-header {
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
  position: sticky;
  top: 0;
  z-index: 100;
}
.nav-container {
  max-width: 1080px;
  margin: 0 auto;
  padding: 16px 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.brand-logo {
  font-size: 1.25rem;
  font-weight: 800;
  color: #0f172a;
  text-decoration: none;
  letter-spacing: -0.03em;
  display: flex;
  align-items: center;
  gap: 8px;
}
.nav-links {
  display: flex;
  gap: 24px;
  align-items: center;
}
.nav-item {
  text-decoration: none;
  color: #475569;
  font-weight: 500;
  font-size: 0.95rem;
  transition: color 0.2s ease;
}
.nav-item:hover, .nav-item.router-link-active {
  color: #2563eb;
  font-weight: 600;
}
.cart-badge-wrapper {
  position: relative;
  display: inline-flex;
  align-items: center;
}
.cart-badge {
  background: #2563eb;
  color: white;
  font-size: 0.75rem;
  font-weight: 700;
  padding: 2px 7px;
  border-radius: 999px;
  margin-left: 6px;
}
</style>