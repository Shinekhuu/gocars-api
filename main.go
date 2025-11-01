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
	err := godotenv.Load("/home/ubuntu/project-go/gocars-api/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	database.InitDB()

	// Auto-migrate models here
	if err := database.DB.AutoMigrate(
		&models.User{},
		&models.Otp{},
		&models.Xyr{},
		&models.Manufacturer{},
		&models.Model{},
		&models.Engine{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	router := gin.Default()
	// Configure CORS
	config := cors.Config{
		AllowOrigins:     []string{"https://gocars.mn", "http://localhost:5173"}, // your frontend domain
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}
	router.Use(cors.New(config))

	router.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"message": "Welcome"}) })
	router.POST("/signin", controllers.SignIn)
	router.POST("/signup", controllers.SignUp)
	router.POST("/verify-otp", controllers.VerifyOtp)
	router.POST("/resend-otp", controllers.ResendOtp)

	router.GET("/manufactures", controllers.GetManufacturers)
	router.GET("/model", controllers.GetModels)
	router.GET("/engine", controllers.GetEngines)

	router.GET("/crawler", controllers.Garage)
	router.GET("/search", controllers.SearchOEM)
	router.GET("/shop", controllers.Shop)

	router.GET("/decode", controllers.Decode)
	router.POST("/vehicle", controllers.GetXyrData)

	// Protected routes
	authorized := router.Group("/")
	authorized.Use(middleware.AuthRequired())
	authorized.GET("/profile", controllers.Profile)

	router.Run(":9000")
}
