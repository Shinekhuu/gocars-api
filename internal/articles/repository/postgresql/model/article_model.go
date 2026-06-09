package model

import "gorm.io/gorm"

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
	IsFetched            bool  `gorm:"type:boolean;default:false"`
	DictionaryID         *uint `gorm:"index"`

	AllSpecifications []ArticleAllSpecification `gorm:"foreignKey:ArticleItemID;references:ID;constraint:OnDelete:CASCADE;"`
	AllOems           []ArticleOem              `gorm:"foreignKey:ArticleItemID;references:ID;constraint:OnDelete:CASCADE;"`
	AllCategories     []ArticleCategory         `gorm:"foreignKey:ArticleItemID;references:ID;constraint:OnDelete:CASCADE;"`
	AllVehicles       []ArticleVehicles         `gorm:"foreignKey:ArticleItemID;references:ID;constraint:OnDelete:CASCADE;"`
	Dictionary        *Dictionary               `gorm:"foreignKey:DictionaryID;references:ID;constraint:OnDelete:SET NULL;"`

	OemResponses           []ArticleOemResponse     `json:"oemNo" gorm:"-"`
	CompatibleCarsResponse []CompatibleCarsResponse `json:"compatibleCars" gorm:"-"`
}

type ArticleAllSpecification struct {
	gorm.Model
	CriteriaName  string `json:"criteriaName" gorm:"type:varchar(255)"`
	CriteriaValue string `json:"criteriaValue" gorm:"type:varchar(255)"`
	ArticleItemID uint   `json:"articleItemId" gorm:"column:article_item_id;index"`
}

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
}

type VehicleArticlesResponse struct {
	VehicleID  string        `json:"vehicleId"`
	CategoryID string        `json:"categoryId"`
	Articles   []ArticleItem `json:"articles"`
}

type ArticleOem struct {
	gorm.Model
	ArticleItemID uint `json:"articleItemId" gorm:"column:article_item_id;uniqueIndex:idx_article_oem"`
	OemID         uint `json:"oemId" gorm:"column:oem_id;uniqueIndex:idx_article_oem"`

	Oem Oem `gorm:"foreignKey:OemID;references:ID;constraint:OnDelete:CASCADE"`
}

type ArticleCategory struct {
	gorm.Model
	CategoryID    uint `json:"categoryId" gorm:"column:category_id;uniqueIndex:idx_article_category"`
	ArticleItemID uint `json:"articleItemId" gorm:"column:article_item_id;uniqueIndex:idx_article_category"`

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
	OemNo             []ArticleOemResponse      `json:"oemNo"`
	CompatibleCars    []CompatibleCarsResponse  `json:"compatibleCars"`
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
