package models

type MeiliArticle struct {
	ID            uint   `json:"id"`
	ArticleNo     string `json:"article_no"`
	SearchNo      string `json:"article_search_no"`
	ProductName   string `json:"product_name"`
	ProductNameMN string `json:"product_name_mn,omitempty"`
	Supplier      string `json:"supplier"`
	Image         string `json:"image"`

	CategoryID     uint   `json:"category_id"`
	CategoryName   string `json:"category_name"`
	CategoryNameMN string `json:"category_name_mn"`
}
