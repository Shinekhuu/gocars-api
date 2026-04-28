package models

import (
	"encoding/json"
	"fmt"
	"gocars-api/database"
	"io"
	"log"
	"net/http"
	"os"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ArticleItem represents a single article item
type ArticleItem struct {
	gorm.Model
	ArticleID            *uint  `json:"articleId" gorm:"column:article_id;index"`
	ArticleSearchNo      string `json:"articleSearchNo" gorm:"column:article_search_no;type:text"`
	ArticleNo            string `json:"articleNo" gorm:"column:article_no;type:text"`
	ArticleProductName   string `json:"articleProductName" gorm:"column:article_product_name;type:text"`
	ProductID            int    `json:"productId" gorm:"column:product_id"`
	ManufacturerID       int    `json:"manufacturerId" gorm:"column:manufacturer_id"`
	ManufacturerName     string `json:"manufacturerName" gorm:"column:manufacturer_name;type:text"`
	SupplierID           int    `json:"supplierId" gorm:"column:supplier_id"`
	SupplierName         string `json:"supplierName" gorm:"column:supplier_name;type:text"`
	ArticleMediaType     string `json:"articleMediaType" gorm:"column:article_media_type;type:text"`
	ArticleMediaFileName string `json:"articleMediaFileName" gorm:"column:article_media_file_name;type:text"`
	S3Image              string `json:"s3image" gorm:"column:s3_image;type:text"`
	Price                float64
	IsFetched            bool `gorm:"type:tinyint(1);default:0"`

	// MUST ADD THIS
	AllSpecifications []ArticleAllSpecification `gorm:"foreignKey:ArticleItemID;references:ID;constraint:OnDelete:CASCADE"`
	AllOems           []ArticleOem              `gorm:"foreignKey:ArticleItemID;references:ID;constraint:OnDelete:CASCADE;"`
	AllCategories     []ArticleCategory         `gorm:"foreignKey:ArticleItemID;references:ID;constraint:OnDelete:CASCADE;"`
	AllVehicles       []ArticleVehicles         `gorm:"foreignKey:ArticleItemID;references:ID;constraint:OnDelete:CASCADE;"`

	// For JSON unmarshalling from API (not stored directly)
	OemResponses           []ArticleOemResponse     `json:"oemNo" gorm:"-"`
	CompatibleCarsResponse []CompatibleCarsResponse `json:"compatibleCars" gorm:"-"`
}

type ArticleAllSpecification struct {
	gorm.Model
	CriteriaName  string `json:"criteriaName" gorm:"type:varchar(255)"`
	CriteriaValue string `json:"criteriaValue" gorm:"type:varchar(255)"`
	ArticleItemID uint   `json:"articleItemId" gorm:"column:article_item_id;index"`
}

// ArticleOemResponse is only for API JSON unmarshalling
type ArticleOemResponse struct {
	Brand     string `json:"oemBrand"`
	DisplayNo string `json:"oemDisplayNo"`
}

type CompatibleCarsResponse struct {
	VehicleID                 uint   `json:"vehicleId"`
	ModelID                   uint   `json:"modelId"`
	ManufacturerID            uint   `json:"manufacturerId"`
	ManufacturerName          string `json:"manufacturerName"`
	ModelName                 string `json:"modelName"`
	TypeEngineName            string `json:"typeEngineName"`
	ConstructionIntervalStart string `json:"constructionIntervalStart"`
	ConstructionIntervalEnd   string `json:"constructionIntervalEnd"`
}

type ArticleVehicles struct {
	gorm.Model
	VehicleID     uint `json:"vehicleId" gorm:"column:vehicle_id;uniqueIndex:idx_article_vehicle;"`
	ArticleItemID uint `json:"articleItemId" gorm:"column:article_item_id;uniqueIndex:idx_article_vehicle;"`

	Engine Engine `gorm:"foreignKey:VehicleID;references:VehicleID;constraint:OnDelete:CASCADE"`
}

type VehicleArticlesResponse struct {
	VehicleID  string        `json:"vehicleId"`
	CategoryID string        `json:"categoryId"` // optional
	Articles   []ArticleItem `json:"articles"`   // must match JSON
}

type ArticleOem struct {
	gorm.Model
	ArticleItemID uint `json:"articleItemId" gorm:"column:article_item_id;uniqueIndex:idx_article_oem"`
	OemID         uint `json:"oemId" gorm:"column:oem_id;uniqueIndex:idx_article_oem"`

	// Add this so GORM can preload the OEM
	Oem Oem `gorm:"foreignKey:OemID;references:ID;constraint:OnDelete:CASCADE"`
}

type ArticleCategory struct {
	gorm.Model
	CategoryID    uint `json:"categoryId" gorm:"column:category_id;uniqueIndex:idx_article_category"`
	ArticleItemID uint `json:"articleItemId" gorm:"column:article_item_id;uniqueIndex:idx_article_category"`

	// Add this so GORM can preload the Category
	Category Category `gorm:"foreignKey:CategoryID;references:CategoryID;constraint:OnDelete:CASCADE"`
}

type RapidAPIResponse struct {
	Article ArticleAPI `json:"article"`
}

type ArticleAPI struct {
	ArticleID            uint   `json:"articleId"`
	ArticleNo            string `json:"articleNo"`
	ArticleProductName   string `json:"articleProductName"`
	SupplierName         string `json:"supplierName"`
	SupplierID           uint   `json:"supplierId"`
	ProductID            int    `json:"productId"`
	ArticleMediaType     string `json:"articleMediaType"`
	ArticleMediaFileName string `json:"articleMediaFileName"`
	S3Image              string `json:"s3image"`

	AllSpecifications []ArticleAllSpecification `json:"allSpecifications"`

	OemNo []ArticleOemResponse `json:"oemNo"`

	CompatibleCars []CompatibleCarsResponse `json:"compatibleCars"`
}

type ArticleItemWithCategory struct {
	ID                 uint    `json:"ID"`
	ArticleID          uint    `json:"articleId"`
	ArticleNo          string  `json:"articleNo"`
	ArticleSearchNo    string  `json:"articleSearchNo"`
	ArticleProductName string  `json:"articleProductName"`
	S3Image            string  `json:"s3image"`
	SupplierName       string  `json:"supplierName"`
	CategoryID         *uint   `json:"categoryId"`
	CategoryName       *string `json:"categoryName"`
	CategoryNameMn     *string `json:"categoryNameMN"`
	Level              *int    `json:"level"`
	Thumbnail          *string `json:"thumbnail"`
	ParentID           *uint   `json:"parentId"`
}

type RapidOEMResponse struct {
	CountArticles int           `json:"countArticles"`
	Articles      []ArticleItem `json:"articles"`
}

func GetArticleItemsByOem(oem string, page int, limit int) (*[]ArticleItem, int64, error) {
	var dbArticleItems []ArticleItem
	var total int64

	// Count total items for pagination
	if err := database.DB.
		Table("oems").
		Joins("LEFT JOIN article_oems ON oems.id = article_oems.oem_id").
		Joins("LEFT JOIN article_items ON article_oems.article_item_id = article_items.id").
		Where(
			"UPPER(REGEXP_REPLACE(oems.display_no, '[^A-Za-z0-9]', '')) = "+
				"UPPER(REGEXP_REPLACE(?, '[^A-Za-z0-9]', ''))",
			oem,
		).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination logic
	offset := (page - 1) * limit

	// Fetch paginated records
	if err := database.DB.
		Table("oems").
		Select("article_items.*").
		Joins("LEFT JOIN article_oems ON oems.id = article_oems.oem_id").
		Joins("LEFT JOIN article_items ON article_oems.article_item_id = article_items.id").
		Where(
			"UPPER(REGEXP_REPLACE(oems.display_no, '[^A-Za-z0-9]', '')) = "+
				"UPPER(REGEXP_REPLACE(?, '[^A-Za-z0-9]', ''))",
			oem,
		).
		Limit(limit).
		Offset(offset).
		Find(&dbArticleItems).Error; err != nil {
		return nil, 0, err
	}

	return &dbArticleItems, total, nil
}

// func GetArticleItemsFromRapidAPI(vehicleID uint, categoryID uint) (*VehicleArticlesResponse, error) {
// 	// 1️⃣ Fetch from HTTP API
// 	url := fmt.Sprintf(
// 		"https://auto-parts-catalog.p.rapidapi.com/articles/list/type-id/1/vehicle-id/%d/category-id/%d/lang-id/4",
// 		vehicleID,
// 		categoryID,
// 	)

// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("error creating request: %w", err)
// 	}
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("x-rapidapi-key", os.Getenv("X_RAPIDAPI_KEY"))
// 	req.Header.Set("x-rapidapi-host", os.Getenv("X_RAPIDAPI_HOST"))

// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return nil, fmt.Errorf("error reading response: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("error reading response: %w", err)
// 	}

// 	var vehicleArticlesResponse VehicleArticlesResponse
// 	if err := json.Unmarshal(body, &vehicleArticlesResponse); err != nil {
// 		return nil, fmt.Errorf("error parsing JSON: %w", err)
// 	}

// 	// Save to DB for future requests
// 	go func() {
// 		for i := range vehicleArticlesResponse.Articles {
// 			article := &vehicleArticlesResponse.Articles[i]
// 			_ = database.DB.
// 				Where(ArticleItem{ArticleID: article.ArticleID}).
// 				Assign(article).
// 				FirstOrCreate(article)

// 			av := ArticleVehicles{
// 				ArticleItemID: article.ID,
// 				VehicleID:     utils.AtoiUint(vehicleArticlesResponse.VehicleID),
// 			}
// 			_ = database.DB.
// 				Where("vehicle_id = ? AND article_item_id = ?", av.VehicleID, av.ArticleItemID).
// 				Assign(av).
// 				FirstOrCreate(&av)

// 			ac := ArticleCategory{
// 				ArticleItemID: article.ID,
// 				CategoryID:    utils.AtoiUint(vehicleArticlesResponse.CategoryID),
// 			}
// 			_ = database.DB.
// 				Where("category_id = ? AND article_item_id = ?", ac.CategoryID, av.ArticleItemID).
// 				Assign(ac).
// 				FirstOrCreate(&ac)
// 		}
// 	}()

// 	return &vehicleArticlesResponse, nil
// }

func GetArticleItemsByOemFromRapidAPI(oem string) ([]ArticleItem, error) {
	var rapidOEMResponse RapidOEMResponse

	// 1️⃣ Fetch from HTTP API
	url := fmt.Sprintf(
		"https://auto-parts-catalog.p.rapidapi.com/artlookup/search-articles-by-article-no??lang-id=4&articleNo=%s&articleType=OENumber",
		oem,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-rapidapi-key", os.Getenv("X_RAPIDAPI_KEY"))
	req.Header.Set("x-rapidapi-host", os.Getenv("X_RAPIDAPI_HOST"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// 2️⃣ Parse JSON --> articles (slice)
	if err := json.Unmarshal(body, &rapidOEMResponse); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	// 3️⃣ Save async
	// Save asynchronously
	go func() {
		err := database.DB.Transaction(func(tx *gorm.DB) error {
			for i := range rapidOEMResponse.Articles {
				article := rapidOEMResponse.Articles[i]

				// MUST USE POINTER
				if err := tx.
					Where("article_id = ?", article.ArticleID).
					Assign(article).
					FirstOrCreate(&article).Error; err != nil {
					return err
				}

				newOem := Oem{
					Brand:     article.ManufacturerName,
					DisplayNo: oem,
				}

				// Upsert Oem
				if err := tx.
					Where("brand = ? AND display_no = ?", newOem.Brand, newOem.DisplayNo).
					Assign(&newOem).
					FirstOrCreate(&newOem).Error; err != nil {
					return err
				}

				// Link Article <-> OEM
				link := ArticleOem{
					ArticleItemID: article.ID,
					OemID:         newOem.ID,
				}

				if err := tx.
					Where("article_item_id = ? AND oem_id = ?", link.ArticleItemID, link.OemID).
					Assign(&link).
					FirstOrCreate(&link).Error; err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			fmt.Println("Async article & oem save failed:", err)
		}
	}()

	return rapidOEMResponse.Articles, nil
}

// categories is a JSON string like "[1,2,3]"
func SaveArticleItemToDB(article *ArticleItem, vehicleIDs []uint, categories string) (*ArticleItem, error) {

	err := database.DB.Transaction(func(tx *gorm.DB) error {

		// =========================
		// 1. CHECK EXISTING ARTICLE
		// =========================
		err := tx.
			Model(&ArticleItem{}).
			Select("id").
			Where("article_no = ?", article.ArticleNo).
			Limit(1).
			Scan(&article.ID).Error

		if err != nil {
			return err
		}

		exists := article.ID != 0

		// =========================
		// 2. OEM (FirstOrCreate)
		// =========================
		oem := Oem{
			Brand:     article.SupplierName,
			DisplayNo: article.ArticleNo,
		}

		if err := tx.
			Where(Oem{
				Brand:     oem.Brand,
				DisplayNo: oem.DisplayNo,
			}).
			FirstOrCreate(&oem).Error; err != nil {
			return err
		}

		// =========================
		// 3. ARTICLE (CREATE IF NOT EXISTS)
		// =========================
		if !exists {
			if err := tx.Create(article).Error; err != nil {
				return err
			}
		}

		// =========================
		// 4. LINK ARTICLE <-> OEM
		// =========================
		if err := tx.
			Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "article_item_id"},
					{Name: "oem_id"},
				},
				DoNothing: true,
			}).
			Create(&ArticleOem{
				ArticleItemID: article.ID,
				OemID:         oem.ID,
			}).Error; err != nil {
			return err
		}

		// =========================
		// 5. VEHICLES (DEDUP + BULK)
		// =========================
		if len(vehicleIDs) > 0 {
			vMap := make(map[uint]struct{}, len(vehicleIDs))
			vehicles := make([]ArticleVehicles, 0, len(vehicleIDs))

			for _, vid := range vehicleIDs {
				if vid == 0 {
					continue
				}
				if _, ok := vMap[vid]; ok {
					continue
				}
				vMap[vid] = struct{}{}

				vehicles = append(vehicles, ArticleVehicles{
					ArticleItemID: article.ID,
					VehicleID:     vid,
				})
			}

			if len(vehicles) > 0 {
				if err := tx.
					Clauses(clause.OnConflict{
						Columns: []clause.Column{
							{Name: "article_item_id"},
							{Name: "vehicle_id"},
						},
						DoNothing: true,
					}).
					Create(&vehicles).Error; err != nil {
					return err
				}
			}
		}

		// =========================
		// 6. CATEGORIES (DEDUP + BULK)
		// =========================
		if categories != "" && categories != "null" {
			var categoryIDs []uint

			if err := json.Unmarshal([]byte(categories), &categoryIDs); err != nil {
				log.Println("⚠️ invalid categories:", categories)
				return err
			}

			if len(categoryIDs) > 0 {
				cMap := make(map[uint]struct{}, len(categoryIDs))
				cats := make([]ArticleCategory, 0, len(categoryIDs))

				for _, cid := range categoryIDs {
					if cid == 0 {
						continue
					}
					if _, ok := cMap[cid]; ok {
						continue
					}
					cMap[cid] = struct{}{}

					cats = append(cats, ArticleCategory{
						ArticleItemID: article.ID,
						CategoryID:    cid,
					})
				}

				if len(cats) > 0 {
					if err := tx.
						Clauses(clause.OnConflict{
							Columns: []clause.Column{
								{Name: "article_item_id"},
								{Name: "category_id"},
							},
							DoNothing: true,
						}).
						Create(&cats).Error; err != nil {
						return err
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		log.Println(article.ArticleNo, "❌ Save failed:", err)
		return nil, err
	}

	return article, nil
}
