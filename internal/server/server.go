package server

import (
	"log"

	"gocars-api/internal/app"
	"gocars-api/internal/articles/jobs"
	"gocars-api/internal/config"
	pgdb "gocars-api/internal/database/postgres"
	"gocars-api/internal/middleware"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Run(cfg config.Config, application *app.App) {
	var router *gin.Engine

	if cfg.MODE == "PRODUCTION" {
		gin.SetMode(gin.ReleaseMode)
		router = gin.New()
		router.Use(gin.Logger())
		router.Use(gin.Recovery())
	} else {
		gin.SetMode(gin.DebugMode)
		router = gin.Default()
	}

	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatalf("failed to set trusted proxies: %v", err)
	}

	_ = sentry.CurrentHub()
	router.Use(sentrygin.New(sentrygin.Options{}))

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://gocars.mn", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	if cfg.MODE == "PRODUCTION" {
		autoMigrate(pgdb.DB)
	}

	jobs.StartWorker(pgdb.DB)

	router.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok", "service": "gocars-api"}) })
	router.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"message": "Welcome"}) })

	router.GET("/manufacturers", application.ManufacturerHdl.GetManufacturers)
	router.GET("/model", application.ModelHdl.GetModels)
	router.GET("/engine", application.EngineHdl.GetEngines)

	router.GET("/vehicle", application.VehicleHdl.FetchData)
	router.GET("/search", application.SearchHdl.Search)
	router.GET("/shop", application.ShopHdl.Shop)
	router.GET("/oems", application.OemHdl.GetOEMs)

	router.GET("/article", application.ArticleHdl.Article)
	router.POST("/order", application.OrderHdl.CreateOrder)
	router.GET("/order/:id", application.OrderHdl.GetOrder)
	router.GET("/orders/:id/pdf", application.OrderHdl.GetOrderPDF)

	router.GET("/products", application.ProductHdl.GetProducts)

	// Profile — auth is handled by gocars-auth (port 9001); gocars-api owns profile data
	authorized := router.Group("/")
	authorized.Use(middleware.AuthRequired())
	authorized.GET("/profile", application.ProfileHdl.Profile)
	authorized.PUT("/profile", application.ProfileHdl.UpdateProfile)

	router.Run(":9000")
}
