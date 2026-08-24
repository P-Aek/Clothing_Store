<template>
  <div class="container" style="max-width: 420px; margin: 40px auto; padding: 0 20px;">
    <div class="card" style="padding: 30px; background: white; border-radius: 12px; border: 1px solid #e2e8f0; box-shadow: 0 4px 12px rgba(0,0,0,0.05);">
      <h2 style="text-align: center; margin-top: 0; margin-bottom: 24px; color: #0f172a;">สมัครสมาชิก</h2>
      
      <form @submit.prevent="handleRegister">
        <div class="form-group" style="margin-bottom: 16px;">
          <label style="display: block; font-weight: 600; margin-bottom: 6px; color: #475569; font-size: 0.9rem;">ชื่อ-นามสกุล</label>
          <input type="text" v-model="form.name" required placeholder="เช่น สมใจ ใจดี" style="width: 100%; padding: 10px 12px; border-radius: 8px; border: 1px solid #cbd5e1; box-sizing: border-box; font-size: 0.95rem;">
        </div>

        <div class="form-group" style="margin-bottom: 16px;">
          <label style="display: block; font-weight: 600; margin-bottom: 6px; color: #475569; font-size: 0.9rem;">อีเมล</label>
          <input type="email" v-model="form.email" required placeholder="name@example.com" style="width: 100%; padding: 10px 12px; border-radius: 8px; border: 1px solid #cbd5e1; box-sizing: border-box; font-size: 0.95rem;">
        </div>

        <div class="form-group" style="margin-bottom: 20px;">
          <label style="display: block; font-weight: 600; margin-bottom: 6px; color: #475569; font-size: 0.9rem;">รหัสผ่าน</label>
          <input type="password" v-model="form.pass" required placeholder="••••••••" style="width: 100%; padding: 10px 12px; border-radius: 8px; border: 1px solid #cbd5e1; box-sizing: border-box; font-size: 0.95rem;">
        </div>

        <button type="submit" class="btn btn-primary" style="width: 100%; padding: 12px; font-weight: 600; background: #2563eb; color: white; border: none; border-radius: 8px; cursor: pointer; font-size: 0.95rem;">
          ยืนยันการสมัคร
        </button>
      </form>

      <div style="text-align: center; margin-top: 20px;">
        <router-link to="/account" style="color: #2563eb; text-decoration: none; font-size: 0.88rem;">
          มีบัญชีอยู่แล้ว? เข้าสู่ระบบ
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '../stores/appStore'

const store = useAppStore()
const router = useRouter()

const form = reactive({
  name: '',
  email: '',
  pass: ''
})

const handleRegister = async () => {
  const res = await store.register(form)
  if (res.success) {
    alert('สมัครสมาชิกสำเร็จ!')
    router.push('/account')
  } else {
    alert('เกิดข้อผิดพลาดในการสมัครสมาชิก')
  }
}
</script>