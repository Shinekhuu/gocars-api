package main

import (
	"log"

	"gocars-api/controllers"
	"gocars-api/database"
	"gocars-api/middleware"
	"gocars-api/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// For Production
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())   // keep logs (recommended)
	router.Use(gin.Recovery()) // crash protection

	// For Development
	// gin.SetMode(gin.DebugMode)
	// router := gin.Default()

	// ✅ Trusted proxies (important for production)
	router.SetTrustedProxies(nil)

	// Configure CORS
	config := cors.Config{
		AllowOrigins:     []string{"https://gocars.mn", "http://localhost:5173"}, // your frontend domain
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}
	router.Use(cors.New(config))

	err := godotenv.Load("/home/ubuntu/project-go/gocars-api/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	database.InitDB()

	// Auto-migrate models here
	if err := database.DB.AutoMigrate(
		&models.Category{},
		&models.Oem{},
		&models.Xyr{},
		&models.User{},
		&models.Otp{},
		&models.Manufacturer{},
		&models.Model{},
		&models.EngineFamily{},
		&models.ModelFamily{},
		&models.ArticleAllSpecification{},
		&models.ArticleVehicles{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Auto-migrate models here
	if err := database.DB.AutoMigrate(
		&models.Engine{},
		&models.ArticleOem{},
		&models.ArticleCategory{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Auto-migrate models here
	if err := database.DB.AutoMigrate(
		&models.ArticleItem{},
		&models.XyrVehicle{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Auto-migrate models here
	if err := database.DB.AutoMigrate(
		&models.Order{},
		&models.OrderItem{},
		&models.Invoice{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	router.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"message": "Welcome"}) })
	router.POST("/signin", controllers.SignIn)
	router.POST("/signup", controllers.SignUp)
	router.POST("/verify-otp", controllers.VerifyOtp)
	router.POST("/resend-otp", controllers.ResendOtp)

	router.GET("/manufacturers", controllers.GetManufacturers)
	// router.GET("/manufacturers-seeder", controllers.FillData)
	router.GET("/model", controllers.GetModels)
	router.GET("/engine", controllers.GetEngines)

	router.GET("/vehicle", controllers.FetchData)
	router.GET("/search", controllers.SearchOEM)
	router.GET("/shop", controllers.Shop)

	// router.GET("/decode", controllers.Decode)

	// router.GET("/partsouq", scraper.FetchVehicleInfoPartsouq)

	// router.GET("/article-seeder", controllers.FillArticleItemData)
	router.GET("/article", controllers.Article)

	// router.GET("/category-seeder", controllers.FillCategories)

	// Protected routes
	authorized := router.Group("/")
	authorized.Use(middleware.AuthRequired())
	authorized.GET("/profile", controllers.Profile)
	authorized.POST("/order", controllers.CreateOrder)

	router.Run(":9000")
}
