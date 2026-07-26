<template>
  <div class="container">
    <div class="form-card">
      <h2>หยิบสินค้าส่งออก (Picking)</h2>
      <div class="form-group">
        <label>Outbound Plan <span class="required">*</span></label>
        <select v-model="form.plan_id" :class="{ 'input-error': errors.plan_id }">
          <option value="">-- เลือก Plan --</option>
          <option v-for="p in plans" :key="p.id" :value="p.id">
            {{ p.plan_code }} — {{ p.sku }} (เหลือหยิบ {{ p.plan_qty - p.picked_qty }}/{{ p.plan_qty }})
          </option>
        </select>
        <span v-if="errors.plan_id" class="error-msg">{{ errors.plan_id }}</span>
      </div>
      <div class="form-group">
        <label>จำนวนที่หยิบรอบนี้ (Qty) <span class="required">*</span></label>
        <input v-model.number="form.qty" type="number" :class="{ 'input-error': errors.qty }"/>
        <span v-if="errors.qty" class="error-msg">{{ errors.qty }}</span>
      </div>
      <div class="form-group">
        <label>Location (หยิบจากที่ไหน) <span class="required">*</span></label>
        <select v-model="form.location_id" :class="{ 'input-error': errors.location_id }">
          <option value="">-- เลือก Location --</option>
          <option v-for="loc in locations" :key="loc.id" :value="loc.id">
            {{ loc.location_code }} (Zone: {{ loc.zone }})
          </option>
        </select>
        <span v-if="errors.location_id" class="error-msg">{{ errors.location_id }}</span>
      </div>
      <button @click="submitPick" :disabled="loading">
        {{ loading ? 'กำลังบันทึก...' : 'Confirm Pick' }}
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

const form = ref({ plan_id: '', qty: 0, location_id: '' })
const errors = ref({ plan_id: '', qty: '', location_id: '' })
const plans = ref([])
const locations = ref([])
const loading = ref(false)
const message = ref('')
const success = ref(false)

const fetchPlans = async () => {
  try {
    const res = await axios.get('/api/outbound-plans', { headers: authHeader() })
    plans.value = res.data.filter(p => p.status !== 'Completed')
  } catch (err) {
    console.error('โหลด plan ไม่ได้:', err)
  }
}

const fetchLocations = async () => {
  try {
    const res = await axios.get('/api/locations')
    locations.value = res.data
  } catch (err) {
    console.error('โหลด location ไม่ได้:', err)
  }
}

const validate = () => {
  errors.value = { plan_id: '', qty: '', location_id: '' }
  let valid = true
  if (!form.value.plan_id) { errors.value.plan_id = 'Required field'; valid = false }
  if (form.value.qty <= 0) { errors.value.qty = 'Required field'; valid = false }
  if (!form.value.location_id) { errors.value.location_id = 'Required field'; valid = false }
  return valid
}

const submitPick = async () => {
  if (!validate()) return
  loading.value = true
  message.value = ''
  try {
    await axios.post('/api/picking', {
      plan_id: form.value.plan_id,
      qty: form.value.qty,
      location_id: form.value.location_id,
    }, { headers: authHeader() })
    message.value = 'หยิบสินค้าสำเร็จ!'
    success.value = true
    form.value = { plan_id: '', qty: 0, location_id: '' }
    errors.value = { plan_id: '', qty: '', location_id: '' }
    fetchPlans()
  } catch (err) {
    message.value = err.response?.data?.error || 'เกิดข้อผิดพลาด'
    success.value = false
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchPlans()
  fetchLocations()
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
.error-msg { color: #e74c3c; font-size: 12px; margin-top: 4px; display: block; }
button { background: #3498db; color: white; border: none; padding: 12px 24px; border-radius: 8px; cursor: pointer; font-size: 14px; margin-top: 8px; }
button:hover { background: #2980b9; }
button:disabled { background: #aaa; cursor: not-allowed; }
.message { margin-top: 16px; padding: 12px; border-radius: 8px; }
.success { background: #d4edda; color: #155724; }
.error { background: #f8d7da; color: #721c24; }
</style>
