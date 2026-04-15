package app

import (
	"gocars-api/controllers"
	"gocars-api/repositories"
	"gocars-api/services"

	"gorm.io/gorm"
)

type App struct {
	ProductHdl *controllers.ProductHandler
}

func NewApp(db *gorm.DB) *App {
	repo := repositories.NewProductRepository(db)
	svc := services.NewProductService(repo)
	hdl := controllers.NewProductHandler(svc)

	return &App{
		ProductHdl: hdl,
	}
}
