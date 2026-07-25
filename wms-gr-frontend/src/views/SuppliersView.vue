<template>
  <div class="container">
    <div class="form-card">
      <h2>เพิ่ม Supplier</h2>
      <div class="form-group">
        <label>Supplier Code <span class="required">*</span></label>
        <input v-model="form.supplier_code" :class="{ 'input-error': errors.supplier_code }" />
        <span v-if="errors.supplier_code" class="error-msg">{{ errors.supplier_code }}</span>
      </div>
      <div class="form-group">
        <label>ชื่อ Supplier</label>
        <input v-model="form.name" />
      </div>
      <button @click="createSupplier" :disabled="creating">
        {{ creating ? 'กำลังบันทึก...' : 'เพิ่ม Supplier' }}
      </button>
      <div v-if="message" :class="['message', success ? 'success' : 'error']">{{ message }}</div>
    </div>

    <div class="list-card">
      <h2>Supplier ทั้งหมด</h2>
      <button @click="fetchSuppliers">Refresh</button>
      <table v-if="suppliers.length > 0">
        <thead>
          <tr>
            <th>Supplier Code</th><th>ชื่อ</th><th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in suppliers" :key="s.id">
            <td>{{ s.supplier_code }}</td>
            <td>{{ s.name }}</td>
            <td><button class="delete-btn" @click="deactivate(s)">ปิดใช้งาน</button></td>
          </tr>
        </tbody>
      </table>
      <p v-else>ยังไม่มี supplier</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const form = ref({ supplier_code: '', name: '' })
const errors = ref({ supplier_code: '' })
const suppliers = ref([])
const creating = ref(false)
const message = ref('')
const success = ref(false)

const fetchSuppliers = async () => {
  try {
    const res = await axios.get('/api/suppliers')
    suppliers.value = res.data
  } catch (err) {
    console.error('โหลด supplier ไม่ได้:', err)
  }
}

const createSupplier = async () => {
  errors.value = { supplier_code: '' }
  if (!form.value.supplier_code) {
    errors.value.supplier_code = 'Required field'
    return
  }
  creating.value = true
  message.value = ''
  try {
    await axios.post('/api/suppliers', form.value)
    message.value = 'เพิ่ม supplier สำเร็จ!'
    success.value = true
    form.value = { supplier_code: '', name: '' }
    fetchSuppliers()
  } catch (err) {
    message.value = err.response?.data?.error || 'เกิดข้อผิดพลาด'
    success.value = false
  } finally {
    creating.value = false
  }
}

const deactivate = async (s) => {
  try {
    await axios.delete(`/api/suppliers/${s.id}`)
    fetchSuppliers()
  } catch (err) {
    console.error('ปิด supplier ไม่ได้:', err)
  }
}

onMounted(fetchSuppliers)
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
