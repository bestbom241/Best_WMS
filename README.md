# WMS — Warehouse Management System

Warehouse management system แบบ microservices (Go + Postgres + Vue)

## Services

| Service | Port | หน้าที่ |
|---|---|---|
| `wms-auth-api` | 3002 | login / register / issue token |
| `wms-inventory-api` | 3001 | สต็อกสินค้า (SKU, qty, location), รับ/ตัดสต็อก |
| `wms-master-data-api` | 3003 | Product / Location / Supplier / Customer master data |
| `wms-gr-api` | 3000 | ขาเข้า (Inbound): GR Plan → รับสินค้าจริงแบบ partial ได้ |
| `wms-outbound-api` | 3004 | ขาออก (Outbound): Outbound Plan → หยิบสินค้าจริงแบบ partial ได้ |
| `wms-gr-frontend` | 80 | หน้าเว็บ Vue 3 |

## Flow หลัก

- **Inbound:** สร้าง GR Plan (supplier, sku, qty, วันที่) → รับสินค้าจริงอิงจาก plan (รับทีละส่วนได้) → status `New → Partial → Completed` → sync เข้า `wms-inventory-api` อัตโนมัติ
- **Outbound:** สร้าง Outbound Plan (customer, sku, qty, วันที่) → หยิบสินค้าจริงอิงจาก plan (หยิบทีละส่วนได้) → status `New → Partial → Completed` → ตัดสต็อกจาก `wms-inventory-api` (ปฏิเสธถ้าของไม่พอ กันสต็อกติดลบ)

## รันโปรเจกต์

```sh
cp .env.example .env   # แล้วแก้ค่าตามต้องการ
docker compose up -d --build
```

เปิด `http://localhost` (หรือ `FRONTEND_PORT` ที่ตั้งไว้ใน `.env`)

## Branch strategy

- `main` — โค้ด deploy จริง
- `develop` — งานที่ทดสอบแล้ว รอขึ้น main
- `feature/*` — แยกทำงานทีละเรื่อง แล้ว merge กลับเข้า `develop`
