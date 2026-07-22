package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
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

type GoodsReceive struct {
	ID        string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	PONumber  string         `json:"po_number" gorm:"uniqueIndex"`
	Status    string         `json:"status"`
	SKU       string         `json:"sku"`
	Qty       int            `json:"qty"`
}

type CreateGoodsReceiveRequest struct {
	SKU        string `json:"sku"`
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
	err = db.AutoMigrate(&GoodsReceive{})
	if err != nil {
		fmt.Println("AutoMigrate error:", err.Error())
	} else {
		fmt.Println("AutoMigrate สำเร็จ")
	}
}

func generatePONumber() (string, error) {
	today := time.Now().Format("20060102")
	prefix := "PO-" + today + "-"

	var count int64
	db.Model(&GoodsReceive{}).
		Where("po_number LIKE ?", prefix+"%").
		Count(&count)

	// ลองสร้างเลขถัดไป ถ้าชน (เผื่อ concurrent) ให้ลองเพิ่มอีก
	for i := int64(1); i <= 20; i++ {
		candidate := fmt.Sprintf("%s%04d", prefix, count+i)
		var existing GoodsReceive
		result := db.Where("po_number = ?", candidate).First(&existing)
		if result.Error != nil { // ไม่เจอ = ใช้ได้
			return candidate, nil
		}
	}
	return "", fmt.Errorf("ไม่สามารถ generate PO number ได้")
}

type ShopeeOrder struct {
	OrderSN   string `json:"order_sn"`
	SKU       string `json:"sku"`
	ItemName  string `json:"item_name"`
	Qty       int    `json:"qty"`
	OrderTime string `json:"order_time"`
}

var mockSKUs = []string{"SKU-001", "SKU-002", "SKU-003", "SKU-004", "SKU-005"}
var mockItemNames = []string{"เสื้อยืดสีขาว", "กางเกงยีนส์", "รองเท้าผ้าใบ", "กระเป๋าสะพาย", "หมวกแก๊ป"}

func generateMockShopeeOrders(count int) []ShopeeOrder {
	orders := make([]ShopeeOrder, 0, count)
	for i := 0; i < count; i++ {
		idx := rand.Intn(len(mockSKUs))
		orders = append(orders, ShopeeOrder{
			OrderSN:   fmt.Sprintf("SP%d%04d", time.Now().Unix()%100000, i),
			SKU:       mockSKUs[idx],
			ItemName:  mockItemNames[idx],
			Qty:       rand.Intn(10) + 1,
			OrderTime: time.Now().Format(time.RFC3339),
		})
	}
	return orders
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

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่มี token"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
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

	// routes ที่ต้องการ token
	auth := r.Group("/api", authMiddleware())

	// ── ขอเลข PO number ถัดไปมาแสดงล่วงหน้า (ก่อน submit) ──
	auth.GET("/receiving/next-po-number", func(c *gin.Context) {
		po, err := generatePONumber()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"po_number": po})
	})

	auth.GET("/receiving", func(c *gin.Context) {
		var list []GoodsReceive
		query := db.Model(&GoodsReceive{})
		if status := c.Query("status"); status != "" {
			query = query.Where("status = ?", status)
		}
		if sku := c.Query("sku"); sku != "" {
			query = query.Where("sku ILIKE ?", "%"+sku+"%")
		}
		if po := c.Query("po_number"); po != "" {
			query = query.Where("po_number ILIKE ?", "%"+po+"%")
		}
		if from := c.Query("from"); from != "" {
			query = query.Where("created_at >= ?", from)
		}
		if to := c.Query("to"); to != "" {
			query = query.Where("created_at <= ?", to+" 23:59:59")
		}
		query.Find(&list)
		c.JSON(http.StatusOK, list)
	})

	auth.GET("/receiving/:id", func(c *gin.Context) {
		id := c.Param("id")
		var gr GoodsReceive
		db.First(&gr, "id = ?", id)
		c.JSON(http.StatusOK, gr)
	})

	auth.POST("/receiving", func(c *gin.Context) {
		var req CreateGoodsReceiveRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// validation — เช็ค blank input (po_number ไม่ต้องเช็คแล้ว เพราะ backend เป็นคน generate เอง)
		if req.SKU == "" {
			fmt.Println("Error: sku ว่าง")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Please input SKU"})
			return
		}
		if req.Qty <= 0 {
			fmt.Println("Error: qty ต้องมากกว่า 0")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Please input a valid quantity"})
			return
		}

		poNumber, err := generatePONumber()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		fmt.Println("รับ request:", poNumber, req.SKU, req.Qty)

		gr := GoodsReceive{
			ID:       uuid.New().String(),
			PONumber: poNumber,
			SKU:      req.SKU,
			Qty:      req.Qty,
			Status:   "pending",
		}

		result := db.Create(&gr)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}

		inventoryURL := os.Getenv("INVENTORY_URL") + "/api/stock"
		if os.Getenv("INVENTORY_URL") == "" {
			inventoryURL = "http://localhost:3001/api/stock"
		}

		locationCode := getLocationCode(req.LocationID)
		payload, _ := json.Marshal(map[string]interface{}{
			"sku":      req.SKU,
			"name":     "สินค้า " + req.SKU,
			"qty":      req.Qty,
			"location": locationCode,
		})

		resp, err := http.Post(inventoryURL, "application/json", bytes.NewBuffer(payload))

		if err != nil {
			fmt.Println("แจ้ง Inventory ไม่ได้:", err.Error())
		} else {
			fmt.Println("แจ้ง Inventory สำเร็จ status:", resp.StatusCode)
		}

		c.JSON(http.StatusCreated, gr)
	})

	auth.PUT("/receiving/:id", func(c *gin.Context) {
		id := c.Param("id")
		var gr GoodsReceive
		db.First(&gr, "id = ?", id)
		gr.Status = "received"
		db.Save(&gr)
		c.JSON(http.StatusOK, gr)
	})

	auth.DELETE("/receiving/:id", func(c *gin.Context) {
		id := c.Param("id")
		db.Delete(&GoodsReceive{}, "id = ?", id)
		c.JSON(http.StatusOK, gin.H{"message": "ลบสำเร็จ"})
	})

	// ── Mock Shopee: ดูออเดอร์ที่ "เข้ามาใหม่" (จำลอง) ──
	auth.GET("/shopee/orders", func(c *gin.Context) {
		orders := generateMockShopeeOrders(5)
		c.JSON(http.StatusOK, orders)
	})

	// ── Mock Shopee: sync ออเดอร์ → สร้าง GR อัตโนมัติทุกออเดอร์ ──
	auth.POST("/shopee/sync", func(c *gin.Context) {
		orders := generateMockShopeeOrders(5)
		created := []GoodsReceive{}

		for _, o := range orders {
			poNumber, err := generatePONumber()
			if err != nil {
				continue
			}

			gr := GoodsReceive{
				ID:       uuid.New().String(),
				PONumber: poNumber,
				SKU:      o.SKU,
				Qty:      o.Qty,
				Status:   "pending",
			}

			if result := db.Create(&gr); result.Error != nil {
				continue
			}
			created = append(created, gr)

			inventoryURL := os.Getenv("INVENTORY_URL") + "/api/stock"
			if os.Getenv("INVENTORY_URL") == "" {
				inventoryURL = "http://localhost:3001/api/stock"
			}
			payload, _ := json.Marshal(map[string]interface{}{
				"sku":  o.SKU,
				"name": o.ItemName,
				"qty":  o.Qty,
			})
			http.Post(inventoryURL, "application/json", bytes.NewBuffer(payload))
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       fmt.Sprintf("Sync สำเร็จ สร้าง GR ทั้งหมด %d รายการ", len(created)),
			"orders_synced": len(orders),
			"gr_created":    created,
		})
	})

	r.Run(":3000")
}
