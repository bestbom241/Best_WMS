<template>
  <div class="container">
    <div class="form-card">
      <h2>+ New Plan</h2>
      <div class="form-group">
       <label>Supplier Code</label>
       <select v-model="form.supplier_code">
        <option value="">-- เลือก Supplier --</option>
        <option v-for="s in suppliers" :key="s.id" :value="s.supplier_code">
      {{ s.supplier_code }} — {{ s.name }}
    </option>
  </select>
</div>
      <div class="form-group">
        <label>สินค้า (SKU) <span class="required">*</span></label>
        <select v-model="form.sku" :class="{ 'input-error': errors.sku }">
          <option value="">-- เลือกสินค้า --</option>
          <option v-for="p in products" :key="p.id" :value="p.sku">{{ p.sku }} — {{ p.name }}</option>
        </select>
        <span v-if="errors.sku" class="error-msg">{{ errors.sku }}</span>
      </div>
      <div class="form-group">
        <label>Plan Quantity <span class="required">*</span></label>
        <input v-model.number="form.plan_qty" type="number" :class="{ 'input-error': errors.plan_qty }" />
        <span v-if="errors.plan_qty" class="error-msg">{{ errors.plan_qty }}</span>
      </div>
      <div class="form-group">
        <label>Plan Receive Date</label>
        <input v-model="form.plan_date" type="date" />
      </div>
      <button @click="createPlan" :disabled="creating">
        {{ creating ? 'กำลังบันทึก...' : 'สร้าง Plan' }}
      </button>
      <div v-if="message" :class="['message', success ? 'success' : 'error']">{{ message }}</div>
    </div>

    <div class="list-card">
      <h2>Plan Goods Receipt</h2>
      <button @click="fetchPlans">Refresh</button>
      <table v-if="plans.length > 0">
        <thead>
          <tr>
            <th>Plan Code</th><th>Supplier</th><th>SKU</th><th>Plan Qty</th><th>Received</th><th>Plan Date</th><th>Status</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in plans" :key="p.id">
            <td>{{ p.plan_code }}</td>
            <td>{{ p.supplier_code }}</td>
            <td>{{ p.sku }}</td>
            <td>{{ p.plan_qty }}</td>
            <td>{{ p.received_qty }}</td>
            <td>{{ p.plan_date }}</td>
            <td><span :class="['status-badge', statusClass(p.status)]">{{ p.status }}</span></td>
          </tr>
        </tbody>
      </table>
      <p v-else>ยังไม่มี plan</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { authHeader } from '../auth'

const form = ref({ supplier_code: '', sku: '', plan_qty: 0, plan_date: '' })
const errors = ref({ sku: '', plan_qty: '' })
const plans = ref([])
const products = ref([])
const suppliers = ref([])
const creating = ref(false)
const message = ref('')
const success = ref(false)

const fetchPlans = async () => {
  try {
    const res = await axios.get('/api/gr-plans', { headers: authHeader() })
    plans.value = res.data
  } catch (err) {
    console.error('โหลด plan ไม่ได้:', err)
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

const fetchSuppliers = async () => {
  try {
    const res = await axios.get('/api/suppliers')
    suppliers.value = res.data
  } catch (err) {
    console.error('โหลด supplier ไม่ได้:', err)
  }
}

const statusClass = (status) => {
  if (status === 'Completed') return 'status-completed'
  if (status === 'Partial') return 'status-partial'
  return 'status-new'
}

const createPlan = async () => {
  errors.value = { sku: '', plan_qty: '' }
  let valid = true
  if (!form.value.sku) { errors.value.sku = 'Required field'; valid = false }
  if (form.value.plan_qty <= 0) { errors.value.plan_qty = 'Required field'; valid = false }
  if (!valid) return

  creating.value = true
  message.value = ''
  try {
    await axios.post('/api/gr-plans', form.value, { headers: authHeader() })
    message.value = 'สร้าง plan สำเร็จ!'
    success.value = true
    form.value = { supplier_code: '', sku: '', plan_qty: 0, plan_date: '' }
    fetchPlans()
  } catch (err) {
    message.value = err.response?.data?.error || 'เกิดข้อผิดพลาด'
    success.value = false
  } finally {
    creating.value = false
  }
}

onMounted(() => {
  fetchPlans()
  fetchProducts()
  fetchSuppliers()
})
</script>

<style scoped>
.container { max-width: 900px; margin: 40px auto; padding: 0 20px; font-family: sans-serif; }
.form-card, .list-card { background: #f8f9fa; border-radius: 12px; padding: 24px; margin-bottom: 24px; }
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
table { width: 100%; border-collapse: collapse; margin-top: 16px; }
th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
th { background: #ecf0f1; font-weight: 600; }
.status-badge { padding: 4px 10px; border-radius: 12px; font-size: 12px; font-weight: 600; }
.status-new { background: #fff3cd; color: #856404; }
.status-partial { background: #d1ecf1; color: #0c5460; }
.status-completed { background: #d4edda; color: #155724; }
</style>
