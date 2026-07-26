package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var jwtSecret = []byte("wms-secret-key-2024")

// OutboundPlan คือแผนส่งสินค้าออก (คู่กับ GRPlan ฝั่งรับเข้า) — pick แบบ partial ได้
type OutboundPlan struct {
	ID           string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	PlanCode     string         `json:"plan_code" gorm:"uniqueIndex"`
	CustomerCode string         `json:"customer_code"`
	SKU          string         `json:"sku"`
	PlanQty      int            `json:"plan_qty"`
	PickedQty    int            `json:"picked_qty"`
	PlanDate     string         `json:"plan_date"`
	Status       string         `json:"status"` // New / Partial / Completed
}

type CreateOutboundPlanRequest struct {
	CustomerCode string `json:"customer_code"`
	SKU          string `json:"sku"`
	PlanQty      int    `json:"plan_qty"`
	PlanDate     string `json:"plan_date"`
}

// Picking คือใบหยิบสินค้าจริงแต่ละครั้ง อ้างอิง OutboundPlan
type Picking struct {
	ID         string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	PlanID     string         `json:"plan_id"`
	PickNumber string         `json:"pick_number" gorm:"uniqueIndex"`
	SKU        string         `json:"sku"`
	Qty        int            `json:"qty"`
}

type CreatePickingRequest struct {
	PlanID     string `json:"plan_id"`
	Qty        int    `json:"qty"`
	LocationID string `json:"location_id"`
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
	db.AutoMigrate(&OutboundPlan{})
	db.AutoMigrate(&Picking{})
	fmt.Println("AutoMigrate สำเร็จ")
}

func generateOutboundPlanCode() (string, error) {
	today := time.Now().Format("0601")
	prefix := "POUT" + today + "-"

	var count int64
	db.Model(&OutboundPlan{}).
		Where("plan_code LIKE ?", prefix+"%").
		Count(&count)

	for i := int64(1); i <= 20; i++ {
		candidate := fmt.Sprintf("%s%04d", prefix, count+i)
		var existing OutboundPlan
		result := db.Where("plan_code = ?", candidate).First(&existing)
		if result.Error != nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("ไม่สามารถ generate plan code ได้")
}

func generatePickNumber() (string, error) {
	today := time.Now().Format("20060102")
	prefix := "PICK-" + today + "-"

	var count int64
	db.Model(&Picking{}).
		Where("pick_number LIKE ?", prefix+"%").
		Count(&count)

	for i := int64(1); i <= 20; i++ {
		candidate := fmt.Sprintf("%s%04d", prefix, count+i)
		var existing Picking
		result := db.Where("pick_number = ?", candidate).First(&existing)
		if result.Error != nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("ไม่สามารถ generate pick number ได้")
}

func getLocationCode(locationID string) string {
	masterDataURL := os.Getenv("MASTER_DATA_URL")
	if masterDataURL == "" {
		masterDataURL = "http://localhost:3003"
	}
	resp, err := http.Get(masterDataURL + "/api/locations/" + locationID)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var loc struct {
		LocationCode string `json:"location_code"`
	}
	json.NewDecoder(resp.Body).Decode(&loc)
	return loc.LocationCode
}

// deductStock สั่ง inventory-api ตัดสต็อกออก คืน error ถ้าของไม่พอ
func deductStock(sku, location string, qty int) error {
	inventoryURL := os.Getenv("INVENTORY_URL")
	if inventoryURL == "" {
		inventoryURL = "http://localhost:3001"
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"sku":      sku,
		"location": location,
		"qty":      qty,
	})
	resp, err := http.Post(inventoryURL+"/api/stock/deduct", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("ติดต่อ inventory ไม่ได้: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var body struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&body)
		if body.Error == "" {
			body.Error = "ตัดสต็อกไม่สำเร็จ"
		}
		return fmt.Errorf("%s", body.Error)
	}
	return nil
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่มี token"})
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token ไม่ถูกต้อง"})
			c.Abort()
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		c.Set("user_id", claims["user_id"])
		c.Set("role", claims["role"])
		c.Next()
	}
}

func main() {
	initDB()
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5174",
			"http://localhost:5173",
			"http://localhost",
			"http://localhost:80",
		},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Authorization", "Content-Type"},
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	auth := r.Group("/api", authMiddleware())

	// ── Outbound Plan: วางแผนส่งสินค้าออก ──
	auth.GET("/outbound-plans", func(c *gin.Context) {
		var list []OutboundPlan
		query := db.Model(&OutboundPlan{})
		if status := c.Query("status"); status != "" {
			query = query.Where("status = ?", status)
		}
		if customer := c.Query("customer_code"); customer != "" {
			query = query.Where("customer_code ILIKE ?", "%"+customer+"%")
		}
		if sku := c.Query("sku"); sku != "" {
			query = query.Where("sku ILIKE ?", "%"+sku+"%")
		}
		query.Order("created_at desc").Find(&list)
		c.JSON(http.StatusOK, list)
	})

	auth.GET("/outbound-plans/:id", func(c *gin.Context) {
		id := c.Param("id")
		var plan OutboundPlan
		if err := db.First(&plan, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบ plan นี้"})
			return
		}
		c.JSON(http.StatusOK, plan)
	})

	auth.POST("/outbound-plans", func(c *gin.Context) {
		var req CreateOutboundPlanRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.SKU == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Please input SKU"})
			return
		}
		if req.PlanQty <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Please input a valid plan quantity"})
			return
		}

		planCode, err := generateOutboundPlanCode()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		plan := OutboundPlan{
			ID:           uuid.New().String(),
			PlanCode:     planCode,
			CustomerCode: req.CustomerCode,
			SKU:          req.SKU,
			PlanQty:      req.PlanQty,
			PickedQty:    0,
			PlanDate:     req.PlanDate,
			Status:       "New",
		}
		result := db.Create(&plan)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		c.JSON(http.StatusCreated, plan)
	})

	// ── Picking: หยิบสินค้าจริงอิงจาก plan (ตัดสต็อกจริงผ่าน inventory-api) ──
	auth.POST("/picking", func(c *gin.Context) {
		var req CreatePickingRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.PlanID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Please select an outbound plan"})
			return
		}
		if req.Qty <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Please input a valid quantity"})
			return
		}
		if req.LocationID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Please select a location"})
			return
		}

		var plan OutboundPlan
		if err := db.First(&plan, "id = ?", req.PlanID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบ plan นี้"})
			return
		}
		if plan.Status == "Completed" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "plan นี้ส่งของครบแล้ว"})
			return
		}

		locationCode := getLocationCode(req.LocationID)
		if err := deductStock(plan.SKU, locationCode, req.Qty); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		pickNumber, err := generatePickNumber()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		pick := Picking{
			ID:         uuid.New().String(),
			PlanID:     req.PlanID,
			PickNumber: pickNumber,
			SKU:        plan.SKU,
			Qty:        req.Qty,
		}
		if result := db.Create(&pick); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}

		plan.PickedQty += req.Qty
		if plan.PickedQty >= plan.PlanQty {
			plan.Status = "Completed"
		} else {
			plan.Status = "Partial"
		}
		db.Save(&plan)

		c.JSON(http.StatusCreated, pick)
	})

	r.Run(":3004")
}
