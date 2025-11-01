package models

import "gorm.io/gorm"

// ArticleItem represents a single article item
type ArticleItem struct {
	gorm.Model
	ArticleID            int    `json:"articleId"`
	ArticleSearchNo      string `json:"articleSearchNo"`
	ArticleNo            string `json:"articleNo"`
	ArticleProductName   string `json:"articleProductName"`
	ManufacturerID       int    `json:"manufacturerId"`
	ManufacturerName     string `json:"manufacturerName"`
	SupplierID           int    `json:"supplierId"`
	SupplierName         string `json:"supplierName"`
	ArticleMediaType     string `json:"articleMediaType"`
	ArticleMediaFileName string `json:"articleMediaFileName"`
	S3Image              string `json:"s3image"`
}

// ArticleList is just a slice of ArticleItem
type ArticleList []ArticleItem

type VehicleArticles struct {
	VehicleId     int           `json:"vehicleId"`
	CategoryId    int           `json:"categoryId"`
	CountArticles int           `json:"countArticles"`
	Articles      []ArticleItem `json:"articles"`
}
