<template>
  <div class="container">
    <div class="list-card">
      <h2>สต็อกสินค้าคงคลัง (Inventory Balance)</h2>
      <button @click="fetchStock" :disabled="loading">{{ loading ? 'กำลังโหลด...' : 'Refresh' }}</button>
      <table v-if="stockList.length > 0">
        <thead>
          <tr>
            <th>SKU</th><th>ชื่อสินค้า</th><th>จำนวนคงเหลือ</th><th>Warehouse</th><th>Location</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in stockList" :key="item.id">
            <td>{{ item.sku }}</td>
            <td>{{ item.name }}</td>
            <td>{{ item.qty }}</td>
            <td>{{ item.warehouse_code }}</td>
            <td>{{ item.location }}</td>
          </tr>
        </tbody>
      </table>
      <p v-else>ยังไม่มีข้อมูลสต็อก</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const stockList = ref([])
const loading = ref(false)

const fetchStock = async () => {
  loading.value = true
  try {
    const res = await axios.get('/api/stock')
    stockList.value = res.data
  } catch (err) {
    console.error('โหลด stock ไม่ได้:', err)
  } finally {
    loading.value = false
  }
}

onMounted(fetchStock)
</script>

<style scoped>
.container { max-width: 800px; margin: 40px auto; padding: 0 20px; font-family: sans-serif; }
.list-card { background: #f8f9fa; border-radius: 12px; padding: 24px; margin-bottom: 24px; }
h2 { margin-bottom: 16px; color: #34495e; }
button { background: #3498db; color: white; border: none; padding: 12px 24px; border-radius: 8px; cursor: pointer; font-size: 14px; margin-top: 8px; }
button:hover { background: #2980b9; }
button:disabled { background: #aaa; cursor: not-allowed; }
table { width: 100%; border-collapse: collapse; margin-top: 16px; }
th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
th { background: #ecf0f1; font-weight: 600; }
</style>
