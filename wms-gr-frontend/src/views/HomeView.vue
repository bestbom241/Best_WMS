<template>
  <div class="container">
    <div class="form-card">
      <h2>รับสินค้าเข้าคลัง</h2>
      <div class="form-group">
        <label>PO Number</label>
        <input v-model="form.po_number" readonly class="input-readonly" />
      </div>
      <div class="form-group">
        <label>สินค้า (SKU) <span class="required">*</span></label>
        <select v-model="form.sku" :class="{ 'input-error': errors.sku }">
          <option value="">-- เลือกสินค้า --</option>
          <option v-for="p in products" :key="p.id" :value="p.sku">
            {{ p.sku }} — {{ p.name }}
          </option>
        </select>
        <span v-if="errors.sku" class="error-msg">{{ errors.sku }}</span>
      </div>
      <div class="form-group">
        <label>จำนวน (Qty) <span class="required">*</span></label>
        <input v-model.number="form.qty" type="number" :class="{ 'input-error': errors.qty }"/>
        <span v-if="errors.qty" class="error-msg">{{ errors.qty }}</span>
      </div>
      <div class="form-group">
        <label>Location <span class="required">*</span></label>
        <select v-model="form.location_id" :class="{ 'input-error': errors.location_id }">
          <option value="">-- เลือก Location --</option>
          <option v-for="loc in locations" :key="loc.id" :value="loc.id">
            {{ loc.location_code }} (Zone: {{ loc.zone }})
          </option>
        </select>
        <span v-if="errors.location_id" class="error-msg">{{ errors.location_id }}</span>
      </div>
      <button @click="submitGR" :disabled="loading">
        {{ loading ? 'กำลังบันทึก...' : 'Confirm GR' }}
      </button>
      <div v-if="message" :class="['message', success ? 'success' : 'error']">
        {{ message }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { authHeader } from '../auth'

const form = ref({ po_number: '', sku: '', qty: 0, location_id: '' })
const errors = ref({ sku: '', qty: '', location_id: '' })
const locations = ref([])
const products = ref([])
const loading = ref(false)
const message = ref('')
const success = ref(false)

const fetchLocations = async () => {
  try {
    const res = await axios.get('/api/locations')
    locations.value = res.data
  } catch (err) {
    console.error('โหลด location ไม่ได้:', err)
  }
}

const fetchProducts = async () => {
  try {
    const res = await axios.get('/api/products')
    products.value = res.data
  } catch (err) {
    console.error('โหลดสินค้าไม่ได้:', err)
  }
}

const fetchNextPO = async () => {
  try {
    const res = await axios.get('/api/receiving/next-po-number', { headers: authHeader() })
    form.value.po_number = res.data.po_number
  } catch (err) {
    console.error('โหลด PO number ไม่ได้:', err)
  }
}

const validate = () => {
  errors.value = { sku: '', qty: '', location_id: '' }
  let valid = true
  if (!form.value.sku) { errors.value.sku = 'Required field'; valid = false }
  if (form.value.qty <= 0) { errors.value.qty = 'Required field'; valid = false }
  if (!form.value.location_id) { errors.value.location_id = 'Required field'; valid = false }
  return valid
}

const submitGR = async () => {
  if (!validate()) return
  loading.value = true
  message.value = ''
  try {
    await axios.post('/api/receiving', form.value, { headers: authHeader() })
    message.value = 'รับสินค้าสำเร็จ!'
    success.value = true
    form.value = { po_number: '', sku: '', qty: 0, location_id: '' }
    errors.value = { sku: '', qty: '', location_id: '' }
    fetchNextPO()
  } catch (err) {
    message.value = err.response?.data?.error || 'เกิดข้อผิดพลาด'
    success.value = false
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchNextPO()
  fetchLocations()
  fetchProducts()
})
</script>

<style scoped>
.container { max-width: 800px; margin: 40px auto; padding: 0 20px; font-family: sans-serif; }
.form-card { background: #f8f9fa; border-radius: 12px; padding: 24px; margin-bottom: 24px; }
h2 { margin-bottom: 16px; color: #34495e; }
.form-group { margin-bottom: 16px; }
label { display: block; margin-bottom: 6px; font-weight: 500; color: #555; }
.required { color: #e74c3c; }
input, select { width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 8px; font-size: 14px; box-sizing: border-box; background: white; }
.input-error { border-color: #e74c3c !important; background: #fff5f5; }
.input-readonly { background: #f0f0f0; color: #666; cursor: not-allowed; }
.error-msg { color: #e74c3c; font-size: 12px; margin-top: 4px; display: block; }
button { background: #3498db; color: white; border: none; padding: 12px 24px; border-radius: 8px; cursor: pointer; font-size: 14px; margin-top: 8px; }
button:hover { background: #2980b9; }
button:disabled { background: #aaa; cursor: not-allowed; }
.message { margin-top: 16px; padding: 12px; border-radius: 8px; }
.success { background: #d4edda; color: #155724; }
.error { background: #f8d7da; color: #721c24; }
</style>
