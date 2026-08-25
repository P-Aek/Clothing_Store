<template>
  <header class="navbar-header">
    <div class="nav-container">
      <router-link to="/" class="brand-logo"><span class="logo-icon">👕</span> PLAIN TEE</router-link>
      <nav class="nav-links">
        <router-link to="/" class="nav-item">หน้าแรก</router-link>
        <router-link to="/cart" class="nav-item cart-badge-wrapper">ตะกร้าสินค้า <span v-if="store.cartCount > 0" class="cart-badge">{{ store.cartCount }}</span></router-link>
        <router-link v-if="!store.isAuthenticated" to="/account" class="nav-item">เข้าสู่ระบบ</router-link>
        <div v-else class="account-menu">
          <button class="nav-item account-button" @click="isOpen = !isOpen">◯ {{ store.currentUser?.name || store.currentUser?.email }}</button>
          <div v-if="isOpen" class="account-dropdown">
            <router-link to="/account" class="dropdown-item" @click="isOpen = false">บัญชีของฉัน</router-link>
            <router-link v-if="store.isAdmin" to="/admin" class="dropdown-item" @click="isOpen = false">Admin Management</router-link>
            <button class="dropdown-item logout-item" @click="handleLogout">ออกจากระบบ</button>
          </div>
        </div>
        <router-link v-if="store.isAdmin" to="/admin" class="nav-item admin-link">Admin Management</router-link>
      </nav>
    </div>
  </header>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '../stores/appStore'

const store = useAppStore()
const router = useRouter()
const isOpen = ref(false)
async function handleLogout() { isOpen.value = false; await store.logout(); router.push('/') }
</script>

<style scoped>
.navbar-header { background: #fff; border-bottom: 1px solid #e2e8f0; position: sticky; top: 0; z-index: 100; }
.nav-container { max-width: 1080px; margin: 0 auto; padding: 16px 20px; display: flex; justify-content: space-between; align-items: center; }
.brand-logo { font-size: 1.25rem; font-weight: 800; color: #0f172a; text-decoration: none; letter-spacing: -.03em; display: flex; align-items: center; gap: 8px; }
.nav-links { display: flex; gap: 24px; align-items: center; }
.nav-item { text-decoration: none; color: #475569; font-weight: 500; font-size: .95rem; transition: color .2s ease; }
.nav-item:hover, .nav-item.router-link-active { color: #2563eb; font-weight: 600; }
.cart-badge-wrapper { position: relative; display: inline-flex; align-items: center; }
.cart-badge { background: #2563eb; color: #fff; font-size: .75rem; font-weight: 700; padding: 2px 7px; border-radius: 999px; margin-left: 6px; }
.account-menu { position: relative; }
.account-button { border: 0; background: transparent; cursor: pointer; }
.account-dropdown { position: absolute; right: 0; top: calc(100% + 10px); min-width: 170px; padding: 6px; background: #fff; border: 1px solid #e2e8f0; border-radius: 8px; box-shadow: 0 8px 20px rgba(15,23,42,.12); }
.dropdown-item { display: block; width: 100%; padding: 8px 10px; color: #334155; background: #fff; border: 0; text-align: left; text-decoration: none; cursor: pointer; box-sizing: border-box; }
.dropdown-item:hover { background: #f1f5f9; }
.logout-item { color: #dc2626; }
</style>
