package model

type Product struct {
	ID                 uint    `json:"ID" gorm:"column:id"`
	ArticleID          uint    `json:"articleId" gorm:"column:article_id"`
	ArticleNo          string  `json:"articleNo" gorm:"column:article_no"`
	ArticleSearchNo    string  `json:"articleSearchNo" gorm:"column:article_search_no"`
	ArticleProductName string  `json:"articleProductName" gorm:"column:article_product_name"`
	S3Image            *string `json:"s3image" gorm:"column:s3_image"`
	SupplierName       string  `json:"supplierName" gorm:"column:supplier_name"`

	CategoryID     *uint   `json:"categoryId" gorm:"column:category_id"`
	CategoryName   *string `json:"categoryName" gorm:"column:category_name"`
	CategoryNameMN *string `json:"categoryNameMN" gorm:"column:category_name_mn"`
	Level          *int    `json:"-" gorm:"column:level"`
	Thumbnail      *string `json:"thumbnail" gorm:"column:thumbnail"`
	ParentID       *uint   `json:"-" gorm:"column:parent_id"`

	OEMsRaw []byte `json:"-" gorm:"column:oems"`
	OEMs    []OEM  `json:"-" gorm:"-"`

	Priority int `json:"-" gorm:"column:priority"`
}

type OEM struct {
	DisplayNo string `json:"displayNo"`
	Brand     string `json:"brand"`
}

type ProductFilter struct {
	Search    *string
	VehicleID *uint
	Page      int
	Limit     int
}
