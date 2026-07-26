package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Inventory struct {
	ID            string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	SKU           string         `json:"sku" gorm:"uniqueIndex:idx_sku_location"`
	Name          string         `json:"name"`
	Qty           int            `json:"qty"`
	Location      string         `json:"location" gorm:"uniqueIndex:idx_sku_location"`
	WarehouseCode string         `json:"warehouse_code"`
}

type UpdateStockRequest struct {
	SKU           string `json:"sku"`
	Qty           int    `json:"qty"`
	Name          string `json:"name"`
	Location      string `json:"location"`
	WarehouseCode string `json:"warehouse_code"`
}

type DeductStockRequest struct {
	SKU      string `json:"sku"`
	Location string `json:"location"`
	Qty      int    `json:"qty"`
}

type InventoryReportRow struct {
	SKU           string `json:"sku"`
	Name          string `json:"name"`
	Qty           int    `json:"qty"`
	Location      string `json:"location"`
	WarehouseCode string `json:"warehouse_code"`
}

// InventoryReportV2Row เหมือนตัวเดิม แต่เสริมข้อมูล product (name/category/unit) และ zone/rack/shelf จาก locations เข้ามาด้วย
type InventoryReportV2Row struct {
	SKU           string `json:"sku"`
	ProductName   string `json:"product_name"`
	Category      string `json:"category"`
	Unit          string `json:"unit"`
	Name          string `json:"name"`
	Qty           int    `json:"qty"`
	Location      string `json:"location"`
	WarehouseCode string `json:"warehouse_code"`
	Zone          string `json:"zone"`
	Rack          string `json:"rack"`
	Shelf         string `json:"shelf"`
}

var db *gorm.DB

func initDB() {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	dsn := fmt.Sprintf("host=%s user=admin password=1234 dbname=wms_gr port=5432 sslmode=disable", host)
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("เชื่อม db ไม่ได้: " + err.Error())
	}
	fmt.Println("เชื่อม db สำเร็จ")
	db.AutoMigrate(&Inventory{})
	fmt.Println("AutoMigrate สำเร็จ")
}

func main() {
	initDB()
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173", "http://localhost:5174", "http://localhost"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Authorization", "Content-Type"},
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// ── Stock endpoints ──────────────────────────────

	r.GET("/api/stock", func(c *gin.Context) {
		var list []Inventory
		db.Find(&list)
		c.JSON(http.StatusOK, list)
	})

	r.GET("/api/stock/:sku", func(c *gin.Context) {
		sku := c.Param("sku")
		var inv Inventory
		result := db.Where("sku = ?", sku).First(&inv)
		if result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบ SKU นี้"})
			return
		}
		c.JSON(http.StatusOK, inv)
	})

	r.POST("/api/stock", func(c *gin.Context) {
		var req UpdateStockRequest
		c.BindJSON(&req)
		loc := req.Location
		if loc == "" {
			loc = "A-01" // fallback เผื่อไม่ได้ส่งมา
		}
		var inv Inventory
		result := db.Where("sku = ? AND location = ?", req.SKU, loc).First(&inv)
		if result.Error != nil {
			inv = Inventory{
				ID:            uuid.New().String(),
				SKU:           req.SKU,
				Name:          req.Name,
				Qty:           req.Qty,
				Location:      loc,
				WarehouseCode: req.WarehouseCode,
			}
			db.Create(&inv)
		} else {
			inv.Qty += req.Qty
			if req.WarehouseCode != "" {
				inv.WarehouseCode = req.WarehouseCode
			}
			db.Save(&inv)
		}
		c.JSON(http.StatusOK, inv)
	})

	// ตัดสต็อกออก (ใช้ตอน picking/ส่งของออก) — ป้องกันไม่ให้ตัดจนติดลบ
	r.POST("/api/stock/deduct", func(c *gin.Context) {
		var req DeductStockRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var inv Inventory
		result := db.Where("sku = ? AND location = ?", req.SKU, req.Location).First(&inv)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่พบสต็อกของ SKU นี้ที่ location นี้"})
			return
		}
		if inv.Qty < req.Qty {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("สต็อกไม่พอ (มีอยู่ %d, ขอตัด %d)", inv.Qty, req.Qty)})
			return
		}
		inv.Qty -= req.Qty
		db.Save(&inv)
		c.JSON(http.StatusOK, inv)
	})

	// ── Report endpoints (raw SQL) ────────────────────

	r.GET("/api/report/inventory", func(c *gin.Context) {
		rows := []InventoryReportRow{}
		db.Raw("SELECT sku, name, qty, location, warehouse_code FROM inventories WHERE deleted_at IS NULL ORDER BY sku").Scan(&rows)
		c.JSON(http.StatusOK, rows)
	})

	// v2: join กับ locations (คนละ service เป็นเจ้าของ table นี้ แต่ใช้ database เดียวกัน) เพื่อโชว์ zone/rack/shelf เพิ่ม
	r.GET("/api/report/inventory_v2", func(c *gin.Context) {
		rows := []InventoryReportV2Row{}
		db.Raw(`SELECT i.sku, p.name AS product_name, p.category, p.unit, i.name, i.qty, i.location, i.warehouse_code, l.zone, l.rack, l.shelf
			FROM inventories i
			LEFT JOIN locations l ON l.location_code = i.location
			LEFT JOIN products p ON p.sku = i.sku
			WHERE i.deleted_at IS NULL
			ORDER BY i.sku`).Scan(&rows)
		c.JSON(http.StatusOK, rows)
	})

	r.Run(":3001")
}
