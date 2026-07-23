package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Product struct {
	ID        string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	SKU       string         `json:"sku" gorm:"uniqueIndex"`
	Name      string         `json:"name"`
	Category  string         `json:"category"`
	Unit      string         `json:"unit"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
}

type Location struct {
	ID           string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	LocationCode string         `json:"location_code" gorm:"uniqueIndex"`
	Zone         string         `json:"zone"`
	Rack         string         `json:"rack"`
	Shelf        string         `json:"shelf"`
	Capacity     int            `json:"capacity"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
}

type Supplier struct {
	ID           string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	SupplierCode string         `json:"supplier_code" gorm:"uniqueIndex"`
	Name         string         `json:"name"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
}

type ProductReportRow struct {
	SKU      string `json:"sku"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Unit     string `json:"unit"`
	IsActive bool   `json:"is_active"`
}

type LocationReportRow struct {
	LocationCode string `json:"location_code"`
	Zone         string `json:"zone"`
	Rack         string `json:"rack"`
	Shelf        string `json:"shelf"`
	Capacity     int    `json:"capacity"`
	IsActive     bool   `json:"is_active"`
}

type ExchangeRateResponse struct {
	Result            string             `json:"result"`
	BaseCode          string             `json:"base_code"`
	TimeLastUpdateUTC string             `json:"time_last_update_utc"`
	Rates             map[string]float64 `json:"rates"`
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
	db.AutoMigrate(&Product{})
	db.AutoMigrate(&Location{})
	db.AutoMigrate(&Supplier{})
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

	// ── Product endpoints ────────────────────────────

	r.GET("/api/products", func(c *gin.Context) {
		var list []Product
		db.Where("is_active = ?", true).Find(&list)
		c.JSON(http.StatusOK, list)
	})

	r.GET("/api/products/:id", func(c *gin.Context) {
		id := c.Param("id")
		var p Product
		result := db.Where("id = ? OR sku = ?", id, id).First(&p)
		if result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบสินค้านี้"})
			return
		}
		c.JSON(http.StatusOK, p)
	})

	r.POST("/api/products", func(c *gin.Context) {
		var body struct {
			SKU      string `json:"sku"`
			Name     string `json:"name"`
			Category string `json:"category"`
			Unit     string `json:"unit"`
		}
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if body.SKU == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "sku ห้ามว่าง"})
			return
		}
		p := Product{
			ID:       uuid.New().String(),
			SKU:      body.SKU,
			Name:     body.Name,
			Category: body.Category,
			Unit:     body.Unit,
			IsActive: true,
		}
		result := db.Create(&p)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "sku นี้มีอยู่แล้ว"})
			return
		}
		c.JSON(http.StatusCreated, p)
	})

	r.PUT("/api/products/:id", func(c *gin.Context) {
		id := c.Param("id")
		var p Product
		if err := db.First(&p, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบสินค้านี้"})
			return
		}
		var body struct {
			Name     string `json:"name"`
			Category string `json:"category"`
			Unit     string `json:"unit"`
			IsActive bool   `json:"is_active"`
		}
		c.BindJSON(&body)
		p.Name = body.Name
		p.Category = body.Category
		p.Unit = body.Unit
		p.IsActive = body.IsActive
		db.Save(&p)
		c.JSON(http.StatusOK, p)
	})

	r.DELETE("/api/products/:id", func(c *gin.Context) {
		id := c.Param("id")
		var p Product
		if err := db.First(&p, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบสินค้านี้"})
			return
		}
		p.IsActive = false
		db.Save(&p)
		c.JSON(http.StatusOK, gin.H{"message": "ปิดสินค้าสำเร็จ"})
	})

	// ── Supplier endpoints ───────────────────────────

	r.GET("/api/suppliers", func(c *gin.Context) {
		var list []Supplier
		db.Where("is_active = ?", true).Find(&list)
		c.JSON(http.StatusOK, list)
	})

	r.GET("/api/suppliers/:id", func(c *gin.Context) {
		id := c.Param("id")
		var s Supplier
		result := db.Where("id = ? OR supplier_code = ?", id, id).First(&s)
		if result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบ supplier นี้"})
			return
		}
		c.JSON(http.StatusOK, s)
	})

	r.POST("/api/suppliers", func(c *gin.Context) {
		var body struct {
			SupplierCode string `json:"supplier_code"`
			Name         string `json:"name"`
		}
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if body.SupplierCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "supplier_code ห้ามว่าง"})
			return
		}
		s := Supplier{
			ID:           uuid.New().String(),
			SupplierCode: body.SupplierCode,
			Name:         body.Name,
			IsActive:     true,
		}
		result := db.Create(&s)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "supplier_code นี้มีอยู่แล้ว"})
			return
		}
		c.JSON(http.StatusCreated, s)
	})

	r.PUT("/api/suppliers/:id", func(c *gin.Context) {
		id := c.Param("id")
		var s Supplier
		if err := db.First(&s, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบ supplier นี้"})
			return
		}
		var body struct {
			Name     string `json:"name"`
			IsActive bool   `json:"is_active"`
		}
		c.BindJSON(&body)
		s.Name = body.Name
		s.IsActive = body.IsActive
		db.Save(&s)
		c.JSON(http.StatusOK, s)
	})

	r.DELETE("/api/suppliers/:id", func(c *gin.Context) {
		id := c.Param("id")
		var s Supplier
		if err := db.First(&s, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบ supplier นี้"})
			return
		}
		s.IsActive = false
		db.Save(&s)
		c.JSON(http.StatusOK, gin.H{"message": "ปิด supplier สำเร็จ"})
	})

	// ── Location endpoints ───────────────────────────

	r.GET("/api/locations", func(c *gin.Context) {
		var list []Location
		db.Where("is_active = ?", true).Find(&list)
		c.JSON(http.StatusOK, list)
	})

	r.GET("/api/locations/:id", func(c *gin.Context) {
		id := c.Param("id")
		var loc Location
		result := db.Where("id = ? OR location_code = ?", id, id).First(&loc)
		if result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบ location นี้"})
			return
		}
		c.JSON(http.StatusOK, loc)
	})

	r.POST("/api/locations", func(c *gin.Context) {
		var body struct {
			LocationCode string `json:"location_code"`
			Zone         string `json:"zone"`
			Rack         string `json:"rack"`
			Shelf        string `json:"shelf"`
			Capacity     int    `json:"capacity"`
		}
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if body.LocationCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "location_code ห้ามว่าง"})
			return
		}
		loc := Location{
			ID:           uuid.New().String(),
			LocationCode: body.LocationCode,
			Zone:         body.Zone,
			Rack:         body.Rack,
			Shelf:        body.Shelf,
			Capacity:     body.Capacity,
			IsActive:     true,
		}
		result := db.Create(&loc)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "location_code นี้มีอยู่แล้ว"})
			return
		}
		c.JSON(http.StatusCreated, loc)
	})

	r.PUT("/api/locations/:id", func(c *gin.Context) {
		id := c.Param("id")
		var loc Location
		if err := db.First(&loc, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบ location นี้"})
			return
		}
		var body struct {
			Zone     string `json:"zone"`
			Rack     string `json:"rack"`
			Shelf    string `json:"shelf"`
			Capacity int    `json:"capacity"`
			IsActive bool   `json:"is_active"`
		}
		c.BindJSON(&body)
		loc.Zone = body.Zone
		loc.Rack = body.Rack
		loc.Shelf = body.Shelf
		loc.Capacity = body.Capacity
		loc.IsActive = body.IsActive
		db.Save(&loc)
		c.JSON(http.StatusOK, loc)
	})

	r.DELETE("/api/locations/:id", func(c *gin.Context) {
		id := c.Param("id")
		var loc Location
		if err := db.First(&loc, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบ location นี้"})
			return
		}
		loc.IsActive = false
		db.Save(&loc)
		c.JSON(http.StatusOK, gin.H{"message": "ปิด location สำเร็จ"})
	})

	// ── Report endpoints (raw SQL) ────────────────────

	r.GET("/api/report/products", func(c *gin.Context) {
		rows := []ProductReportRow{}
		db.Raw("SELECT sku, name, category, unit, is_active FROM products WHERE deleted_at IS NULL ORDER BY sku").Scan(&rows)
		c.JSON(http.StatusOK, rows)
	})

	r.GET("/api/report/locations", func(c *gin.Context) {
		rows := []LocationReportRow{}
		db.Raw("SELECT location_code, zone, rack, shelf, capacity, is_active FROM locations WHERE deleted_at IS NULL ORDER BY location_code").Scan(&rows)
		c.JSON(http.StatusOK, rows)
	})

	// ── External API: exchange rate ─────────────────

	r.GET("/api/exchange-rate/:base", func(c *gin.Context) {
		base := strings.ToUpper(c.Param("base"))

		client := http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get("https://open.er-api.com/v6/latest/" + base)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "เรียก exchange rate api ไม่ได้"})
			return
		}
		defer resp.Body.Close()

		var data ExchangeRateResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "แปลงข้อมูลไม่ได้"})
			return
		}
		if data.Result != "success" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "base currency ไม่ถูกต้อง"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"base":        data.BaseCode,
			"updated_utc": data.TimeLastUpdateUTC,
			"thb":         data.Rates["THB"],
			"usd":         data.Rates["USD"],
			"eur":         data.Rates["EUR"],
			"jpy":         data.Rates["JPY"],
		})
	})

	r.Run(":3003")
}
