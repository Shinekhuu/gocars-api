package models

import (
	"encoding/json"
	"fmt"
	"gocars-api/database"
	"io"
	"net/http"
	"os"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ArticleItem represents a single article item
type ArticleItem struct {
	gorm.Model
	ArticleID            uint   `json:"articleId" gorm:"column:article_id;uniqueIndex"`
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
	ArticleID     uint   `json:"articleId" gorm:"column:article_id"`
	ArticleItemID uint   `json:"articleItemId" gorm:"column:article_item_id;index;"`
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
	ArticleID     uint `json:"articleId" gorm:"column:article_id;index;"`
	VehicleID     uint `json:"vehicleId" gorm:"column:vehicle_id;index;"`
	ArticleItemID uint `json:"articleItemId" gorm:"column:article_item_id;index;"`
}

type VehicleArticlesResponse struct {
	VehicleID  uint          `json:"vehicleId"`
	CategoryID uint          `json:"categoryId"` // optional
	Articles   []ArticleItem `json:"articles"`   // must match JSON
}

type OemArticleResponse struct {
	Articles []ArticleItem `json:"articles"` // must match JSON
}

type Oem struct {
	gorm.Model
	Brand     string `json:"brand" gorm:"type:varchar(255)"`
	DisplayNo string `json:"displayNo" gorm:"type:varchar(255);index"`
}

type ArticleOem struct {
	gorm.Model
	ArticleID     uint `gorm:"index"`
	OemID         uint `gorm:"index"`
	ArticleItemID uint `json:"articleItemId" gorm:"column:article_item_id;index"`

	// Add this so GORM can preload the OEM
	Oem Oem `gorm:"foreignKey:OemID;references:ID;constraint:OnDelete:CASCADE"`
}

type ArticleCategory struct {
	gorm.Model
	ArticleID     uint `json:"articleId" gorm:"column:article_id;index"`
	CategoryID    uint `json:"categoryId" gorm:"column:category_id;index"`
	ArticleItemID uint `json:"articleItemId" gorm:"column:article_item_id;index"`

	// Add this so GORM can preload the Category
	Category Category `gorm:"foreignKey:CategoryID;references:CategoryID;constraint:OnDelete:CASCADE"`
}

type RapidAPIResponse struct {
	Article ArticleItem `json:"article"`
}

func GetArticleItemsByVehicleIdAndCategoryId(vehicleID uint, categoryID uint, page int, limit int) (*[]ArticleItem, int64, error) {
	var dbArticleItems []ArticleItem
	var total int64

	// 1️⃣ Count total items for pagination
	if err := database.DB.
		Table("article_categories AS ac").
		Joins("INNER JOIN article_vehicles AS av ON av.article_item_id = ac.article_item_id").
		Where("ac.category_id = ? AND av.vehicle_id = ?", categoryID, vehicleID).
		Select("COUNT(DISTINCT av.article_item_id)").
		Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// 2️⃣ Pagination logic
	offset := (page - 1) * limit

	// 3️⃣ First, fetch article IDs only
	var articleIDs []uint
	if err := database.DB.
		Table("article_categories AS ac").
		Select("DISTINCT ac.article_item_id").
		Joins("INNER JOIN article_vehicles AS av ON av.article_item_id = ac.article_item_id").
		Where("ac.category_id = ? AND av.vehicle_id = ?", categoryID, vehicleID).
		Order("ac.article_item_id DESC").
		Limit(limit).
		Offset(offset).
		Pluck("ac.article_item_id", &articleIDs).Error; err != nil {
		return nil, 0, err
	}

	// 4️⃣ Fetch article details + OEMs only for these IDs
	if len(articleIDs) > 0 {
		if err := database.DB.
			Table("article_items AS ai").
			Select("ai.*, GROUP_CONCAT(DISTINCT o.display_no SEPARATOR ', ') AS article_search_no").
			Joins("LEFT JOIN article_oems AS ao ON ai.id = ao.article_item_id").
			Joins("LEFT JOIN oems AS o ON ao.oem_id = o.id").
			Where("ai.id IN ?", articleIDs).
			Group("ai.id").
			Find(&dbArticleItems).Error; err != nil {
			return nil, 0, err
		}
	}

	return &dbArticleItems, total, nil
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

func GetArticleItemsFromRapidAPI(vehicleID uint, categoryID uint) (*VehicleArticlesResponse, error) {
	// 1️⃣ Fetch from HTTP API
	url := fmt.Sprintf(
		"https://tecdoc-catalog.p.rapidapi.com/articles/list/type-id/1/vehicle-id/%d/category-id/%d/lang-id/4",
		vehicleID,
		categoryID,
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
		return nil, fmt.Errorf("error reading response: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var vehicleArticlesResponse VehicleArticlesResponse
	if err := json.Unmarshal(body, &vehicleArticlesResponse); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	// Save to DB for future requests
	go func() {
		for i := range vehicleArticlesResponse.Articles {
			article := &vehicleArticlesResponse.Articles[i]
			_ = database.DB.
				Where(ArticleItem{ArticleID: article.ArticleID}).
				Assign(article).
				FirstOrCreate(article)

			av := ArticleVehicles{
				ArticleItemID: article.ID,
				VehicleID:     vehicleArticlesResponse.VehicleID,
			}
			_ = database.DB.
				Where("vehicle_id = ? AND article_item_id = ?", av.VehicleID, av.ArticleItemID).
				Assign(av).
				FirstOrCreate(&av)

			ac := ArticleCategory{
				ArticleItemID: article.ID,
				CategoryID:    vehicleArticlesResponse.CategoryID,
			}
			_ = database.DB.
				Where("category_id = ? AND article_item_id = ?", ac.CategoryID, av.ArticleItemID).
				Assign(ac).
				FirstOrCreate(&ac)
		}
	}()

	return &vehicleArticlesResponse, nil
}

func GetArticleItemsByOemFromRapidAPI(oem string) ([]ArticleItem, error) {
	var articles []ArticleItem

	// 1️⃣ Fetch from HTTP API
	url := fmt.Sprintf(
		"https://tecdoc-catalog.p.rapidapi.com/articles-oem/search-by-article-oem-no/lang-id/4/article-oem-no/%s",
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
	if err := json.Unmarshal(body, &articles); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	// 3️⃣ Save async
	// Save asynchronously
	go func() {
		err := database.DB.Transaction(func(tx *gorm.DB) error {
			for i := range articles {
				article := articles[i]

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

	return articles, nil
}

// Fetch from RapidAPI and save asynchronously
func GetArticleCompleteDetailFromRapidAPI(articleID int) (*ArticleItem, error) {
	url := fmt.Sprintf(
		"https://tecdoc-catalog.p.rapidapi.com/articles/article-complete-details/type-id/1/article-id/%d/lang-id/4/country-filter-id/125",
		articleID,
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
		return nil, fmt.Errorf("error calling API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Unmarshal into wrapper struct
	var apiResp RapidAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	articleDetail := apiResp.Article

	// Save asynchronously
	go func(a ArticleItem) {
		err := database.DB.Transaction(func(tx *gorm.DB) error {
			// 1) Upsert ArticleItem
			if err := tx.
				Where("article_id = ?", a.ArticleID).
				Assign(map[string]interface{}{"is_fetched": true}).
				FirstOrCreate(&a).Error; err != nil {
				return err
			}

			// 2) Upsert Specifications
			for _, s := range a.AllSpecifications {
				spec := ArticleAllSpecification{
					ArticleItemID: a.ID,
					CriteriaName:  s.CriteriaName,
					CriteriaValue: s.CriteriaValue,
				}

				if err := tx.
					Where("article_item_id = ? AND criteria_name = ? AND criteria_value = ?", spec.ArticleItemID, spec.CriteriaName, spec.CriteriaValue).
					Assign(&spec).
					FirstOrCreate(&spec).Error; err != nil {
					return err
				}
			}

			// 3) Upsert OEMs + linking table
			for _, o := range a.OemResponses {

				newOem := Oem{
					Brand:     o.Brand,
					DisplayNo: o.DisplayNo,
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
					ArticleItemID: a.ID,
					OemID:         newOem.ID,
				}

				if err := tx.
					Where("article_item_id = ? AND oem_id = ?", link.ArticleItemID, link.OemID).
					Assign(&link).
					FirstOrCreate(&link).Error; err != nil {
					return err
				}
			}

			// 4) Engines + linking table
			for _, c := range a.CompatibleCarsResponse {

				// Resolve manufacturer ID (only ID)
				tx.Model(&Manufacturer{}).
					Select("manufacturer_id").
					Where("manufacturer_name = ?", c.ManufacturerName).
					Find(&c.ManufacturerID)

				// Upsert Engine
				engine := Engine{
					VehicleID:                 c.VehicleID,
					ManufacturerID:            c.ManufacturerID,
					ManufacturerName:          c.ManufacturerName,
					ModelID:                   c.ModelID,
					ModelName:                 c.ModelName,
					TypeEngineName:            c.TypeEngineName,
					ConstructionIntervalStart: c.ConstructionIntervalStart,
					ConstructionIntervalEnd:   c.ConstructionIntervalEnd,
				}

				if err := tx.Clauses(clause.OnConflict{
					Columns: []clause.Column{{Name: "vehicle_id"}, {Name: "model_id"}},
					DoUpdates: clause.AssignmentColumns([]string{
						"manufacturer_id",
						"manufacturer_name",
						"model_name",
						"type_engine_name",
						"construction_interval_start",
						"construction_interval_end",
					}),
				}).Create(&engine).Error; err != nil {
					return err
				}

				// Link Article <-> Vehicle (Engine)
				link := ArticleVehicles{
					ArticleItemID: a.ID,
					VehicleID:     engine.VehicleID,
				}

				if err := tx.
					Where("article_item_id = ? AND vehicle_id = ?", link.ArticleItemID, link.VehicleID).
					Assign(&link).
					FirstOrCreate(&link).Error; err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			fmt.Println("Async article save failed:", err)
		}
	}(articleDetail)

	return &articleDetail, nil
}
