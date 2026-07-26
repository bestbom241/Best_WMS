<template>
  <div class="container">
    <div class="form-card">
      <h2>เพิ่ม Customer</h2>
      <div class="form-group">
        <label>Customer Code <span class="required">*</span></label>
        <input v-model="form.customer_code" :class="{ 'input-error': errors.customer_code }" />
        <span v-if="errors.customer_code" class="error-msg">{{ errors.customer_code }}</span>
      </div>
      <div class="form-group">
        <label>ชื่อ Customer</label>
        <input v-model="form.name" />
      </div>
      <button @click="createCustomer" :disabled="creating">
        {{ creating ? 'กำลังบันทึก...' : 'เพิ่ม Customer' }}
      </button>
      <div v-if="message" :class="['message', success ? 'success' : 'error']">{{ message }}</div>
    </div>

    <div class="list-card">
      <h2>Customer ทั้งหมด</h2>
      <button @click="fetchCustomers">Refresh</button>
      <table v-if="customers.length > 0">
        <thead>
          <tr>
            <th>Customer Code</th><th>ชื่อ</th><th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="cust in customers" :key="cust.id">
            <td>{{ cust.customer_code }}</td>
            <td>{{ cust.name }}</td>
            <td><button class="delete-btn" @click="deactivate(cust)">ปิดใช้งาน</button></td>
          </tr>
        </tbody>
      </table>
      <p v-else>ยังไม่มี customer</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const form = ref({ customer_code: '', name: '' })
const errors = ref({ customer_code: '' })
const customers = ref([])
const creating = ref(false)
const message = ref('')
const success = ref(false)

const fetchCustomers = async () => {
  try {
    const res = await axios.get('/api/customers')
    customers.value = res.data
  } catch (err) {
    console.error('โหลด customer ไม่ได้:', err)
  }
}

const createCustomer = async () => {
  errors.value = { customer_code: '' }
  if (!form.value.customer_code) {
    errors.value.customer_code = 'Required field'
    return
  }
  creating.value = true
  message.value = ''
  try {
    await axios.post('/api/customers', form.value)
    message.value = 'เพิ่ม customer สำเร็จ!'
    success.value = true
    form.value = { customer_code: '', name: '' }
    fetchCustomers()
  } catch (err) {
    message.value = err.response?.data?.error || 'เกิดข้อผิดพลาด'
    success.value = false
  } finally {
    creating.value = false
  }
}

const deactivate = async (cust) => {
  try {
    await axios.delete(`/api/customers/${cust.id}`)
    fetchCustomers()
  } catch (err) {
    console.error('ปิด customer ไม่ได้:', err)
  }
}

onMounted(fetchCustomers)
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
