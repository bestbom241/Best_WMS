# WMS — Warehouse Management System

Warehouse management system แบบ microservices (Go + Postgres + Vue)

## Services

| Service | Port | หน้าที่ |
|---|---|---|
| `wms-auth-api` | 3002 | login / register / issue token |
| `wms-inventory-api` | 3001 | สต็อกสินค้า (SKU, qty, location) |
| `wms-master-data-api` | 3003 | Product / Location master data |
| `wms-gr-api` | 3000 | รับสินค้าเข้าคลัง (Goods Receiving), ยิง sync ไปที่ inventory-api |
| `wms-gr-frontend` | 80 | หน้าเว็บ Vue 3 |

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
