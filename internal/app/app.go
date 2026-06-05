package app

import (
	articlehandler "gocars-api/internal/articles/handler/http"
	articlesrepo "gocars-api/internal/articles/repository/postgresql"
	articlessvc "gocars-api/internal/articles/service"
	"gocars-api/internal/config"
	profilehandler "gocars-api/internal/profile/handler"
	profilerepo "gocars-api/internal/profile/repository/postgresql"
	profilesvc "gocars-api/internal/profile/service"
	roderhandler "gocars-api/internal/roder/handler"
	roderrepo "gocars-api/internal/roder/repository"
	rodersvc "gocars-api/internal/roder/service"
	vehiclehandler "gocars-api/internal/vehicle/handler"
	vehiclerepo "gocars-api/internal/vehicle/repository/postgresql"
	vehiclesvc "gocars-api/internal/vehicle/service"

	"gorm.io/gorm"
)

type App struct {
	// Articles
	ProductRepo  *articlesrepo.ProductRepository
	ProductSvc   *articlessvc.ProductService
	ProductHdl   *articlehandler.ProductHandler

	ArticleRepo  *articlesrepo.ArticleRepository
	FetchLogRepo *articlesrepo.APIFetchLogRepository
	ArticleSvc   *articlessvc.ArticleService
	ArticleHdl   *articlehandler.ArticleHandler

	OemRepo *articlesrepo.OemRepository
	OemSvc  *articlessvc.OemService
	OemHdl  *articlehandler.OemHandler

	CategoryRepo *articlesrepo.CategoryRepository
	CategoryHdl  *articlehandler.CategoryHandler

	DictRepo *articlesrepo.DictionaryRepository

	SearchHdl *articlehandler.SearchHandler
	ShopHdl   *articlehandler.ShopHandler

	// Roder
	OrderRepo *roderrepo.OrderRepository
	OrderSvc  *rodersvc.OrderService
	OrderHdl  *roderhandler.OrderHandler

	// Profile
	ProfileRepo *profilerepo.ProfileRepository
	ProfileSvc  *profilesvc.ProfileService
	ProfileHdl  *profilehandler.ProfileHandler

	// Vehicle
	VehicleHdl      *vehiclehandler.VehicleHandler
	EngineHdl       *vehiclehandler.EngineHandler
	ModelHdl        *vehiclehandler.ModelHandler
	ManufacturerHdl *vehiclehandler.ManufacturerHandler
	CrawlerHdl      *vehiclehandler.CrawlerHandler
	TodoHdl         *vehiclehandler.TodoHandler
	VinHdl          *vehiclehandler.VinHandler
}

func NewApp(db *gorm.DB, cfg config.Config) *App {
	// Articles
	productRepo  := articlesrepo.NewProductRepository(db)
	articleRepo  := articlesrepo.NewArticleRepository(db)
	fetchLogRepo := articlesrepo.NewAPIFetchLogRepository(db)
	oemRepo      := articlesrepo.NewOemRepository(db)
	categoryRepo := articlesrepo.NewCategoryRepository(db)
	dictRepo     := articlesrepo.NewDictionaryRepository(db)

	productSvc := articlessvc.NewProductService(productRepo)
	articleSvc := articlessvc.NewArticleService(articleRepo, fetchLogRepo, db)
	oemSvc     := articlessvc.NewOemService(oemRepo)

	productHdl  := articlehandler.NewProductHandler(productSvc)
	articleHdl  := articlehandler.NewArticleHandler(articleSvc)
	oemHdl      := articlehandler.NewOemHandler(oemSvc)
	categoryHdl := articlehandler.NewCategoryHandler(categoryRepo)
	searchHdl   := articlehandler.NewSearchHandler(articleRepo, articleSvc)
	shopHdl     := articlehandler.NewShopHandler(articleRepo, fetchLogRepo, articleSvc)

	// Roder
	orderRepo := roderrepo.NewOrderRepository(db)
	orderSvc  := rodersvc.NewOrderService(orderRepo)
	orderHdl  := roderhandler.NewOrderHandler(orderSvc)

	// Profile
	pRepo := profilerepo.NewProfileRepository(db)
	pSvc  := profilesvc.NewProfileService(pRepo)
	pHdl  := profilehandler.NewProfileHandler(pSvc)

	// Vehicle repositories
	engineRepo       := vehiclerepo.NewEngineRepository(db)
	modelRepo        := vehiclerepo.NewModelRepository(db)
	manufacturerRepo := vehiclerepo.NewManufacturerRepository(db)
	xyrRepo          := vehiclerepo.NewXyrRepository(db)

	// Vehicle services
	engineSvc       := vehiclesvc.NewEngineService(engineRepo)
	modelSvc        := vehiclesvc.NewModelService(modelRepo)
	manufacturerSvc := vehiclesvc.NewManufacturerService(manufacturerRepo)
	vehicleCoreSvc  := vehiclesvc.NewVehicleService(xyrRepo, engineRepo)

	// Vehicle services (extras)
	crawlerSvc := vehiclesvc.NewCrawlerService(cfg.GARAGE_HOST)
	vinSvc     := vehiclesvc.NewVinService()

	// Vehicle handlers
	vehicleHdl      := vehiclehandler.NewVehicleHandler(vehicleCoreSvc, engineSvc, modelSvc)
	engineHdl       := vehiclehandler.NewEngineHandler(engineSvc)
	modelHdl        := vehiclehandler.NewModelHandler(modelSvc)
	manufacturerHdl := vehiclehandler.NewManufacturerHandler(manufacturerSvc)
	crawlerHdl      := vehiclehandler.NewCrawlerHandler(crawlerSvc)
	todoHdl         := vehiclehandler.NewTodoHandler(vehicleCoreSvc, manufacturerSvc, modelSvc, engineSvc)
	vinHdl          := vehiclehandler.NewVinHandler(vinSvc)

	return &App{
		// Articles
		ProductRepo:  productRepo,
		ProductSvc:   productSvc,
		ProductHdl:   productHdl,

		ArticleRepo:  articleRepo,
		FetchLogRepo: fetchLogRepo,
		ArticleSvc:   articleSvc,
		ArticleHdl:   articleHdl,

		OemRepo: oemRepo,
		OemSvc:  oemSvc,
		OemHdl:  oemHdl,

		CategoryRepo: categoryRepo,
		CategoryHdl:  categoryHdl,

		DictRepo: dictRepo,

		SearchHdl: searchHdl,
		ShopHdl:   shopHdl,

		// Roder
		OrderRepo: orderRepo,
		OrderSvc:  orderSvc,
		OrderHdl:  orderHdl,

		// Profile
		ProfileRepo: pRepo,
		ProfileSvc:  pSvc,
		ProfileHdl:  pHdl,

		// Vehicle
		VehicleHdl:      vehicleHdl,
		EngineHdl:       engineHdl,
		ModelHdl:        modelHdl,
		ManufacturerHdl: manufacturerHdl,
		CrawlerHdl:      crawlerHdl,
		TodoHdl:         todoHdl,
		VinHdl:          vinHdl,
	}
}
