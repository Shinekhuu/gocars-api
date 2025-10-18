package main

import (
	"log"

	"gocars-api/config"
	"gocars-api/controllers"
	"gocars-api/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("/home/ubuntu/project-go/gocars-api/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	config.InitDB()

	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"message": "Welcome"}) })
	router.POST("/signin", controllers.SignIn)
	router.POST("/signup", controllers.SignUp)
	router.POST("/verify-otp", controllers.VerifyOtp)
	router.POST("/resend-otp", controllers.ResendOtp)

	router.GET("/crawler", controllers.Garage)
	router.GET("/search", controllers.SearchOEM)
	router.GET("/shop", controllers.Shop)

	// Protected routes
	authorized := router.Group("/")
	authorized.Use(middleware.AuthRequired())
	authorized.GET("/profile", controllers.Profile)

	router.Run(":9000")
}
