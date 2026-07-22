<template>
  <div class="container">
    <div class="form-card">
      <h2>เพิ่มสินค้า (Product)</h2>
      <div class="form-group">
        <label>SKU <span class="required">*</span></label>
        <input v-model="form.sku" :class="{ 'input-error': errors.sku }" />
        <span v-if="errors.sku" class="error-msg">{{ errors.sku }}</span>
      </div>
      <div class="form-group">
        <label>ชื่อสินค้า</label>
        <input v-model="form.name" />
      </div>
      <div class="form-group">
        <label>Category</label>
        <input v-model="form.category" />
      </div>
      <div class="form-group">
        <label>Unit</label>
        <input v-model="form.unit" />
      </div>
      <button @click="createProduct" :disabled="creating">
        {{ creating ? 'กำลังบันทึก...' : 'เพิ่มสินค้า' }}
      </button>
      <div v-if="message" :class="['message', success ? 'success' : 'error']">{{ message }}</div>
    </div>

    <div class="list-card">
      <h2>สินค้าทั้งหมด</h2>
      <button @click="fetchProducts">Refresh</button>
      <table v-if="products.length > 0">
        <thead>
          <tr>
            <th>SKU</th><th>ชื่อสินค้า</th><th>Category</th><th>Unit</th><th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in products" :key="p.id">
            <td>{{ p.sku }}</td>
            <td>{{ p.name }}</td>
            <td>{{ p.category }}</td>
            <td>{{ p.unit }}</td>
            <td><button class="delete-btn" @click="deactivate(p)">ปิดใช้งาน</button></td>
          </tr>
        </tbody>
      </table>
      <p v-else>ยังไม่มีสินค้า</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const form = ref({ sku: '', name: '', category: '', unit: '' })
const errors = ref({ sku: '' })
const products = ref([])
const creating = ref(false)
const message = ref('')
const success = ref(false)

const fetchProducts = async () => {
  try {
    const res = await axios.get('/api/products')
    products.value = res.data
  } catch (err) {
    console.error('โหลดสินค้าไม่ได้:', err)
  }
}

const createProduct = async () => {
  errors.value = { sku: '' }
  if (!form.value.sku) {
    errors.value.sku = 'Required field'
    return
  }
  creating.value = true
  message.value = ''
  try {
    await axios.post('/api/products', form.value)
    message.value = 'เพิ่มสินค้าสำเร็จ!'
    success.value = true
    form.value = { sku: '', name: '', category: '', unit: '' }
    fetchProducts()
  } catch (err) {
    message.value = err.response?.data?.error || 'เกิดข้อผิดพลาด'
    success.value = false
  } finally {
    creating.value = false
  }
}

const deactivate = async (p) => {
  try {
    await axios.delete(`/api/products/${p.id}`)
    fetchProducts()
  } catch (err) {
    console.error('ปิดสินค้าไม่ได้:', err)
  }
}

onMounted(fetchProducts)
</script>

<style scoped>
.container { max-width: 800px; margin: 40px auto; padding: 0 20px; font-family: sans-serif; }
.form-card, .list-card { background: #f8f9fa; border-radius: 12px; padding: 24px; margin-bottom: 24px; }
h2 { margin-bottom: 16px; color: #34495e; }
.form-group { margin-bottom: 16px; }
label { display: block; margin-bottom: 6px; font-weight: 500; color: #555; }
.required { color: #e74c3c; }
input { width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 8px; font-size: 14px; box-sizing: border-box; background: white; }
.input-error { border-color: #e74c3c !important; background: #fff5f5; }
.error-msg { color: #e74c3c; font-size: 12px; margin-top: 4px; display: block; }
button { background: #3498db; color: white; border: none; padding: 12px 24px; border-radius: 8px; cursor: pointer; font-size: 14px; margin-top: 8px; }
button:hover { background: #2980b9; }
button:disabled { background: #aaa; cursor: not-allowed; }
.delete-btn { background: #e74c3c; padding: 6px 12px; font-size: 12px; margin-top: 0; }
.delete-btn:hover { background: #c0392b; }
.message { margin-top: 16px; padding: 12px; border-radius: 8px; }
.success { background: #d4edda; color: #155724; }
.error { background: #f8d7da; color: #721c24; }
table { width: 100%; border-collapse: collapse; margin-top: 16px; }
th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
th { background: #ecf0f1; font-weight: 600; }
</style>
