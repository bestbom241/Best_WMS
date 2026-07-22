package main

import (
	"fmt"
	"net/http"
	"os" // ← เพิ่มบรรทัดนี้
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Secret key สำหรับเซ็น JWT
var jwtSecret = []byte("wms-secret-key-2024")

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	Email     string         `json:"email" gorm:"uniqueIndex"`
	Password  string         `json:"-"` // ไม่ส่ง password กลับไป
	Role      string         `json:"role"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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
	db.AutoMigrate(&User{})
	fmt.Println("AutoMigrate สำเร็จ")
}

// สร้าง JWT token
func generateToken(userID uint, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // หมดอายุ 24 ชั่วโมง
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func main() {
	initDB()
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173", "http://localhost:5174", "http://localhost"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Authorization", "Content-Type"},
	}))

	// Register — สมัครสมาชิก
	r.POST("/api/auth/register", func(c *gin.Context) {
		var req RegisterRequest
		c.BindJSON(&req)

		// เข้ารหัส password
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		user := User{
			Email:    req.Email,
			Password: string(hashed),
			Role:     req.Role,
		}

		result := db.Create(&user)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email นี้มีอยู่แล้ว"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "สมัครสมาชิกสำเร็จ",
			"user":    user,
		})
	})

	// Login — เข้าสู่ระบบ
	r.POST("/api/auth/login", func(c *gin.Context) {
		var req LoginRequest
		c.BindJSON(&req)

		// หา user จาก email
		var user User
		result := db.Where("email = ?", req.Email).First(&user)
		if result.Error != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "email หรือ password ไม่ถูกต้อง"})
			return
		}

		// ตรวจ password
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "email หรือ password ไม่ถูกต้อง"})
			return
		}

		// สร้าง JWT token
		token, err := generateToken(user.ID, user.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user":  user,
		})
	})

	r.Run(":3002") // port 3002
}
