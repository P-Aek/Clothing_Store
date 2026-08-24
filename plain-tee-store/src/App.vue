<template>
  <div id="app">
    <!-- 🔔 Toast Notification แจ้งเตือนมุมขวาล่าง -->
    <Transition name="toast">
      <div v-if="store.isToastVisible" class="toast-notification" :class="store.toastType">
        <span class="toast-icon">{{ store.toastType === 'success' ? '✅' : '⚠️' }}</span>
        <span>{{ store.toastMessage }}</span>
      </div>
    </Transition>

    <!-- แสดง Navbar หน้าร้าน เฉพาะตอนที่ไม่ใช่หน้า Admin -->
    <Navbar v-if="!isAdminRoute" />

    <main>
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import Navbar from './components/Navbar.vue'
import { useAppStore } from './stores/appStore'

const route = useRoute()
const store = useAppStore()
const isAdminRoute = computed(() => route.path.startsWith('/admin'))
</script>

<style scoped>
/* 🔔 สไตล์สำหรับ Toast Notification (มุมขวาล่าง) */
.toast-notification {
  position: fixed;
  bottom: 24px; /* 📍 ย้ายมาอยู่ด้านล่าง */
  right: 24px;
  z-index: 9999;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 20px;
  border-radius: 10px;
  background: #0f172a;
  color: white;
  font-size: 0.95rem;
  font-weight: 500;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.15);
}

.toast-notification.success {
  border-left: 4px solid #22c55e;
}

.toast-notification.error {
  border-left: 4px solid #ef4444;
}

.toast-icon {
  font-size: 1.1rem;
}

/* 🎬 Animation ค่อยๆ ลอยขึ้นมาจากล่างแล้วจางหาย */
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(20px); /* ลอยขึ้นจากด้านล่าง */
}
</style>