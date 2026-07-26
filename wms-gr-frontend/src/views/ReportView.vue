<template>
  <div class="container">
    <div class="list-card">
      <h2>Export Report</h2>

      <div class="report-row">
        <div>
          <strong>Inventory</strong>
          <p>สต็อกสินค้าคงคลัง (SKU, ชื่อสินค้า, จำนวน, Warehouse, Location)</p>
        </div>
        <button @click="exportInventory" :disabled="loading.inventory">
          {{ loading.inventory ? 'กำลังส่งออก...' : 'Export Inventory.xlsx' }}
        </button>
      </div>

      <div class="report-row">
        <div>
          <strong>Locations</strong>
          <p>Location master (Code, Zone, Rack, Shelf, Capacity)</p>
        </div>
        <button @click="exportLocations" :disabled="loading.locations">
          {{ loading.locations ? 'กำลังส่งออก...' : 'Export Locations.xlsx' }}
        </button>
      </div>

      <div class="report-row">
        <div>
          <strong>Products</strong>
          <p>Product master (SKU, ชื่อสินค้า, Category, Unit)</p>
        </div>
        <button @click="exportProducts" :disabled="loading.products">
          {{ loading.products ? 'กำลังส่งออก...' : 'Export Products.xlsx' }}
        </button>
      </div>

      <div v-if="errorMsg" class="message error">{{ errorMsg }}</div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import axios from 'axios'
import * as XLSX from 'xlsx'

const loading = ref({ inventory: false, locations: false, products: false })
const errorMsg = ref('')

const downloadAsXlsx = (rows, sheetName, filename) => {
  const sheet = XLSX.utils.json_to_sheet(rows)
  const workbook = XLSX.utils.book_new()
  XLSX.utils.book_append_sheet(workbook, sheet, sheetName)
  XLSX.writeFile(workbook, filename)
}

const runExport = async (key, url, columns, sheetName, filename) => {
  loading.value[key] = true
  errorMsg.value = ''
  try {
    const res = await axios.get(url)
    const rows = res.data.map(row => {
      const picked = {}
      columns.forEach(([field, label]) => { picked[label] = row[field] })
      return picked
    })
    downloadAsXlsx(rows, sheetName, filename)
  } catch (err) {
    errorMsg.value = `Export ${sheetName} ไม่สำเร็จ: ${err.response?.data?.error || err.message}`
  } finally {
    loading.value[key] = false
  }
}

const exportInventory = () => runExport(
  'inventory', '/api/report/inventory',
  [['sku', 'SKU'], ['name', 'ชื่อสินค้า'], ['qty', 'จำนวนคงเหลือ'], ['warehouse_code', 'Warehouse'], ['location', 'Location']],
  'Inventory', 'inventory.xlsx'
)

const exportLocations = () => runExport(
  'locations', '/api/report/locations',
  [['location_code', 'Location Code'], ['zone', 'Zone'], ['rack', 'Rack'], ['shelf', 'Shelf'], ['capacity', 'Capacity'], ['is_active', 'Active']],
  'Locations', 'locations.xlsx'
)

const exportProducts = () => runExport(
  'products', '/api/report/products',
  [['sku', 'SKU'], ['name', 'ชื่อสินค้า'], ['category', 'Category'], ['unit', 'Unit'], ['is_active', 'Active']],
  'Products', 'products.xlsx'
)
</script>

<style scoped>
.container { max-width: 800px; margin: 40px auto; padding: 0 20px; font-family: sans-serif; }
.list-card { background: #f8f9fa; border-radius: 12px; padding: 24px; margin-bottom: 24px; }
h2 { margin-bottom: 16px; color: #34495e; }
.report-row { display: flex; align-items: center; justify-content: space-between; gap: 16px; padding: 16px 0; border-bottom: 1px solid #e0e0e0; }
.report-row:last-of-type { border-bottom: none; }
.report-row strong { display: block; color: #2c3e50; }
.report-row p { margin: 4px 0 0; color: #777; font-size: 13px; }
button { background: #3498db; color: white; border: none; padding: 12px 24px; border-radius: 8px; cursor: pointer; font-size: 14px; white-space: nowrap; }
button:hover { background: #2980b9; }
button:disabled { background: #aaa; cursor: not-allowed; }
.message { margin-top: 16px; padding: 12px; border-radius: 8px; }
.error { background: #f8d7da; color: #721c24; }
</style>
