<template>
  <div class="container">
    <div class="form-card">
      <h2>เพิ่ม Location</h2>
      <div class="form-group">
        <label>Location Code <span class="required">*</span></label>
        <input v-model="form.location_code" :class="{ 'input-error': errors.location_code }" />
        <span v-if="errors.location_code" class="error-msg">{{ errors.location_code }}</span>
      </div>
      <div class="form-group">
        <label>Warehouse</label>
        <select v-model="form.warehouse_code">
          <option value="">-- เลือก Warehouse --</option>
          <option v-for="wh in warehouses" :key="wh.id" :value="wh.warehouse_code">
            {{ wh.warehouse_code }} — {{ wh.name }}
          </option>
        </select>
      </div>
      <div class="form-group">
        <label>Zone</label>
        <input v-model="form.zone" />
      </div>
      <div class="form-group">
        <label>Rack</label>
        <input v-model="form.rack" />
      </div>
      <div class="form-group">
        <label>Shelf</label>
        <input v-model="form.shelf" />
      </div>
      <div class="form-group">
        <label>Capacity</label>
        <input v-model.number="form.capacity" type="number" />
      </div>
      <button @click="createLocation" :disabled="creating">
        {{ creating ? 'กำลังบันทึก...' : 'เพิ่ม Location' }}
      </button>
      <div v-if="message" :class="['message', success ? 'success' : 'error']">{{ message }}</div>
    </div>

    <div class="list-card">
      <h2>Locations ทั้งหมด</h2>
      <button @click="fetchLocations">Refresh</button>
      <table v-if="locations.length > 0">
        <thead>
          <tr>
            <th>Code</th><th>Warehouse</th><th>Zone</th><th>Rack</th><th>Shelf</th><th>Capacity</th><th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="loc in locations" :key="loc.id">
            <td>{{ loc.location_code }}</td>
            <td>{{ loc.warehouse_code }}</td>
            <td>{{ loc.zone }}</td>
            <td>{{ loc.rack }}</td>
            <td>{{ loc.shelf }}</td>
            <td>{{ loc.capacity }}</td>
            <td><button class="delete-btn" @click="deactivate(loc)">ปิดใช้งาน</button></td>
          </tr>
        </tbody>
      </table>
      <p v-else>ยังไม่มี location</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const form = ref({ location_code: '', warehouse_code: '', zone: '', rack: '', shelf: '', capacity: 0 })
const errors = ref({ location_code: '' })
const locations = ref([])
const warehouses = ref([])
const creating = ref(false)
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

const fetchWarehouses = async () => {
  try {
    const res = await axios.get('/api/warehouses')
    warehouses.value = res.data
  } catch (err) {
    console.error('โหลด warehouse ไม่ได้:', err)
  }
}

const createLocation = async () => {
  errors.value = { location_code: '' }
  if (!form.value.location_code) {
    errors.value.location_code = 'Required field'
    return
  }
  creating.value = true
  message.value = ''
  try {
    await axios.post('/api/locations', form.value)
    message.value = 'เพิ่ม location สำเร็จ!'
    success.value = true
    form.value = { location_code: '', warehouse_code: '', zone: '', rack: '', shelf: '', capacity: 0 }
    fetchLocations()
  } catch (err) {
    message.value = err.response?.data?.error || 'เกิดข้อผิดพลาด'
    success.value = false
  } finally {
    creating.value = false
  }
}

const deactivate = async (loc) => {
  try {
    await axios.delete(`/api/locations/${loc.id}`)
    fetchLocations()
  } catch (err) {
    console.error('ปิด location ไม่ได้:', err)
  }
}

onMounted(() => {
  fetchLocations()
  fetchWarehouses()
})
</script>

<style scoped>
.container { max-width: 800px; margin: 40px auto; padding: 0 20px; font-family: sans-serif; }
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
.delete-btn { background: #e74c3c; padding: 6px 12px; font-size: 12px; margin-top: 0; }
.delete-btn:hover { background: #c0392b; }
.message { margin-top: 16px; padding: 12px; border-radius: 8px; }
.success { background: #d4edda; color: #155724; }
.error { background: #f8d7da; color: #721c24; }
table { width: 100%; border-collapse: collapse; margin-top: 16px; }
th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
th { background: #ecf0f1; font-weight: 600; }
</style>
