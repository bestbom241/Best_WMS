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
	ID        string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	SKU       string         `json:"sku" gorm:"uniqueIndex:idx_sku_location"`
	Name      string         `json:"name"`
	Qty       int            `json:"qty"`
	Location  string         `json:"location" gorm:"uniqueIndex:idx_sku_location"`
}

type UpdateStockRequest struct {
	SKU      string `json:"sku"`
	Qty      int    `json:"qty"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

type InventoryReportRow struct {
	SKU      string `json:"sku"`
	Name     string `json:"name"`
	Qty      int    `json:"qty"`
	Location string `json:"location"`
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
				ID:       uuid.New().String(),
				SKU:      req.SKU,
				Name:     req.Name,
				Qty:      req.Qty,
				Location: loc,
			}
			db.Create(&inv)
		} else {
			inv.Qty += req.Qty
			db.Save(&inv)
		}
		c.JSON(http.StatusOK, inv)
	})

	// ── Report endpoints (raw SQL) ────────────────────

	r.GET("/api/report/inventory", func(c *gin.Context) {
		rows := []InventoryReportRow{}
		db.Raw("SELECT sku, name, qty, location FROM inventories WHERE deleted_at IS NULL ORDER BY sku").Scan(&rows)
		c.JSON(http.StatusOK, rows)
	})

	r.Run(":3001")
}
