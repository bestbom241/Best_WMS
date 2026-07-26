<template>
  <div v-if="!token" class="container">
    <h1>WMS</h1>
    <div class="form-card">
      <h2>เข้าสู่ระบบ</h2>
      <div class="form-group">
        <label>Email</label>
        <input v-model="loginForm.email" type="email" placeholder="admin@example.com" />
      </div>
      <div class="form-group">
        <label>Password</label>
        <input v-model="loginForm.password" type="password" />
      </div>
      <button @click="doLogin" :disabled="loginLoading">
        {{ loginLoading ? 'กำลังเข้าสู่ระบบ...' : 'Login' }}
      </button>
      <div v-if="loginError" class="message error">{{ loginError }}</div>
    </div>
  </div>

  <div v-else class="app-shell">
    <nav class="navbar">
      <div class="nav-brand">WMS</div>
      <RouterLink to="/gr-plans">GR Plan</RouterLink>
      <RouterLink to="/">รับสินค้า (GR)</RouterLink>
      <RouterLink to="/outbound-plans">Outbound Plan</RouterLink>
      <RouterLink to="/picking">Picking</RouterLink>
      <RouterLink to="/inventory">Inventory</RouterLink>

      <div class="dropdown" ref="dropdownRef">
        <span class="dropdown-trigger" :class="{ active: isMasterRoute }" @click="masterOpen = !masterOpen">Master ▾</span>
        <div class="dropdown-menu" v-if="masterOpen">
          <RouterLink to="/products" @click="masterOpen = false">Products</RouterLink>
          <RouterLink to="/locations" @click="masterOpen = false">Locations</RouterLink>
          <RouterLink to="/suppliers" @click="masterOpen = false">Suppliers</RouterLink>
          <RouterLink to="/customers" @click="masterOpen = false">Customers</RouterLink>
        </div>
      </div>

      <RouterLink to="/report">Report</RouterLink>
      <button class="logout-btn" @click="logout">Logout</button>
    </nav>
    <RouterView />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { token, login, logout } from './auth'

const loginForm = ref({ email: '', password: '' })
const loginLoading = ref(false)
const loginError = ref('')

const route = useRoute()
const masterOpen = ref(false)
const dropdownRef = ref(null)
const masterPaths = ['/products', '/locations', '/suppliers', '/customers']
const isMasterRoute = computed(() => masterPaths.includes(route.path))

const handleClickOutside = (e) => {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target)) {
    masterOpen.value = false
  }
}

const doLogin = async () => {
  loginLoading.value = true
  loginError.value = ''
  try {
    await login(loginForm.value.email, loginForm.value.password)
  } catch (err) {
    loginError.value = err.response?.data?.error || 'เข้าสู่ระบบไม่ได้'
  } finally {
    loginLoading.value = false
  }
}

onMounted(() => {
  localStorage.removeItem('wms_token')
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.container { max-width: 800px; margin: 40px auto; padding: 0 20px; font-family: sans-serif; }
h1 { color: #2c3e50; margin-bottom: 24px; }
.form-card { background: #f8f9fa; border-radius: 12px; padding: 24px; margin-bottom: 24px; }
h2 { margin-bottom: 16px; color: #34495e; }
.form-group { margin-bottom: 16px; }
label { display: block; margin-bottom: 6px; font-weight: 500; color: #555; }
input { width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 8px; font-size: 14px; box-sizing: border-box; background: white; }
button { background: #3498db; color: white; border: none; padding: 12px 24px; border-radius: 8px; cursor: pointer; font-size: 14px; margin-top: 8px; }
button:hover { background: #2980b9; }
button:disabled { background: #aaa; cursor: not-allowed; }
.message { margin-top: 16px; padding: 12px; border-radius: 8px; }
.error { background: #f8d7da; color: #721c24; }

.app-shell { font-family: sans-serif; }
.navbar { display: flex; align-items: center; gap: 20px; background: #2c3e50; padding: 14px 24px; }
.nav-brand { color: white; font-weight: 700; font-size: 18px; margin-right: 12px; }
.navbar :deep(a) { color: #ecf0f1; text-decoration: none; font-size: 14px; padding: 6px 10px; border-radius: 6px; }
.navbar :deep(a.router-link-exact-active) { background: #3498db; }
.logout-btn { margin-left: auto; background: #e74c3c; margin-top: 0; padding: 8px 16px; font-size: 13px; }
.logout-btn:hover { background: #c0392b; }

.dropdown { position: relative; }
.dropdown-trigger { color: #ecf0f1; font-size: 14px; padding: 6px 10px; border-radius: 6px; cursor: pointer; user-select: none; }
.dropdown-trigger.active { background: #3498db; }
.dropdown-trigger:hover { background: #34495e; }
.dropdown-menu {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
  padding: 6px;
  min-width: 140px;
  display: flex;
  flex-direction: column;
  z-index: 10;
}
.dropdown-menu :deep(a) {
  color: #34495e !important;
  padding: 8px 12px !important;
  border-radius: 6px;
  font-size: 14px;
}
.dropdown-menu :deep(a:hover) { background: #ecf0f1; }
.dropdown-menu :deep(a.router-link-exact-active) { background: #3498db; color: white !important; }
</style>
