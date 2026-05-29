package dto

type ArticleResponse struct {
	Article           ArticleDTO    `json:"article"`
	Oems              []OemDTO      `json:"oems"`
	AllSpecifications interface{}   `json:"allSpecifications"`
	CompatibleCars    PaginationDTO `json:"compatibleCars"`
}

type ArticleDTO struct {
	ID                   *uint  `json:"ID"`
	ArticleID            *uint  `json:"articleId"`
	ArticleSearchNo      string `json:"articleSearchNo"`
	ArticleNo            string `json:"articleNo"`
	ArticleProductName   string `json:"articleProductName"`
	ManufacturerID       uint   `json:"manufacturerId"`
	ManufacturerName     string `json:"manufacturerName"`
	SupplierID           uint   `json:"supplierId"`
	SupplierName         string `json:"supplierName"`
	ArticleMediaType     string `json:"articleMediaType"`
	ArticleMediaFileName string `json:"articleMediaFileName"`
	S3Image              string `json:"s3image"`
}

type OemDTO struct {
	Brand     string `json:"brand"`
	DisplayNo string `json:"displayNo"`
}

type PaginationDTO struct {
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
	Total int64       `json:"total"`
	Data  interface{} `json:"data"`
}

type AllSpecificationDTO struct {
	CriteriaName  string `json:"criteriaName"`
	CriteriaValue string `json:"CriteriaValue"`
}

type APIFetchLogDTO struct {
	VehicleID     int    `json:"vehicleId"`
	CategoryID    int    `json:"categoryId"`
	LastFetchedAt string `json:"lastFetchedAt"`
	IsExpired     bool   `json:"isExpired"`
}
