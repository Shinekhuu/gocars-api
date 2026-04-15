package models

import (
	"encoding/json"
	"gocars-api/database"
)

type Product struct {
	ID                 uint    `json:"ID" gorm:"column:id"`
	ArticleID          uint    `json:"articleId" gorm:"column:article_id"`
	ArticleNo          string  `json:"articleNo" gorm:"column:article_no"`
	ArticleSearchNo    string  `json:"articleSearchNo" gorm:"column:article_search_no"`
	ArticleProductName string  `json:"articleProductName" gorm:"column:article_product_name"`
	S3Image            *string `json:"s3image" gorm:"column:s3_image"`
	SupplierName       string  `json:"supplierName" gorm:"column:supplier_name"`

	CategoryID     *uint   `json:"categoryId" gorm:"column:category_id"`
	CategoryName   *string `json:" categoryName" gorm:"column:category_name"`
	CategoryNameMN *string `json:"categoryNameMN" gorm:"column:category_name_mn"`
	Level          *int    `json:"level" gorm:"column:level"`
	Thumbnail      *string `json:"thumbnail" gorm:"column:thumbnail"`
	ParentID       *uint   `json:"parentId" gorm:"column:parent_id"`

	OEMsRaw []byte `json:"-" gorm:"column:oems"`
	OEMs    []OEM  `json:"oems" gorm:"-"`
}

type OEM struct {
	DisplayNo string `json:"displayNo"`
	Brand     string `json:"brand"`
}

type ProductFilter struct {
	Search     *string
	VehicleID  *uint
	CategoryID *uint
	Page       int
	Limit      int
}

func GetProducts(filter ProductFilter) ([]Product, int64, error) {
	var products []Product
	var total int64

	offset := (filter.Page - 1) * filter.Limit

	search := ""
	if filter.Search != nil {
		search = *filter.Search
	}

	like := "%" + search + "%"

	// =========================
	// ✅ COUNT QUERY
	// =========================
	countQuery := `
	SELECT COUNT(DISTINCT ai.id)
	FROM article_items ai
	LEFT JOIN article_categories ac ON ac.article_item_id = ai.id
	LEFT JOIN categories c ON ac.category_id = c.category_id
	LEFT JOIN article_oems ao ON ao.article_item_id = ai.id
	LEFT JOIN oems o ON ao.oem_id = o.id
	WHERE 1=1
	AND (
		? = '' OR
		ai.article_product_name LIKE ? OR
		c.category_name LIKE ? OR
		c.category_name_mn LIKE ? OR
		o.display_no_clean LIKE CONCAT('%', REGEXP_REPLACE(?, '[^A-Za-z0-9]', ''), '%') OR
		o.brand LIKE ?
	)
	`

	countArgs := []interface{}{
		search,
		like, like, like, search, like,
	}

	if err := database.DB.Raw(countQuery, countArgs...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// =========================
	// ✅ MAIN QUERY
	// =========================
	mainQuery := `
	SELECT 
		ai.id,
		ai.article_id,
		ai.article_no,
		ai.article_search_no,
		ai.article_product_name,
		ai.s3_image,
		ai.supplier_name,

		c.category_id,
		c.category_name,
		c.category_name_mn,
		c.level,
		c.thumbnail,
		c.parent_id,

		CASE 
			WHEN COUNT(o.display_no_clean) = 0 THEN JSON_ARRAY()
			ELSE JSON_ARRAYAGG(
				JSON_OBJECT(
					'display_no', o.display_no_clean,
					'brand', o.brand
				)
			)
		END AS oems,

		CASE 
			WHEN ? != '' AND MAX(
        		o.display_no_clean LIKE CONCAT('%', REGEXP_REPLACE(?, '[^A-Za-z0-9]', ''), '%')
    		) THEN 1
			WHEN ? != '' AND MAX(c.category_name LIKE ?) THEN 2
			WHEN ? != '' AND MAX(c.category_name_mn LIKE ?) THEN 3
			WHEN ? != '' AND MAX(ai.article_product_name LIKE ?) THEN 4
			WHEN ? != '' AND MAX(o.brand LIKE ?) THEN 5
			ELSE 6
		END AS priority

	FROM article_items ai

	LEFT JOIN (
		SELECT article_item_id, MAX(id) AS max_id
		FROM article_categories
		GROUP BY article_item_id
	) ac_max ON ac_max.article_item_id = ai.id

	LEFT JOIN article_categories ac ON ac.id = ac_max.max_id
	LEFT JOIN categories c ON ac.category_id = c.category_id

	LEFT JOIN (
		SELECT DISTINCT 
			ao.article_item_id,
			o.brand,
			o.display_no_clean
		FROM article_oems ao
		INNER JOIN oems o ON ao.oem_id = o.id
		WHERE o.display_no_clean IS NOT NULL
	) o ON o.article_item_id = ai.id

	WHERE 1=1
	AND (
		? = '' OR
		ai.article_product_name LIKE ? OR
		c.category_name LIKE ? OR
		c.category_name_mn LIKE ? OR
		o.display_no_clean LIKE CONCAT('%', REGEXP_REPLACE(?, '[^A-Za-z0-9]', ''), '%') OR
		o.brand LIKE ?
	)

	GROUP BY ai.id, c.category_id
	ORDER BY priority ASC, ai.id DESC
	LIMIT ? OFFSET ?
	`

	args := []interface{}{
		// priority (10 args)
		search, search,
		search, like,
		search, like,
		search, like,
		search, like,

		// WHERE
		search,
		like, like, like,
		search, // OEM raw (cleaned in SQL)
		like,

		// pagination
		filter.Limit,
		offset,
	}

	if err := database.DB.Raw(mainQuery, args...).Scan(&products).Error; err != nil {
		return nil, 0, err
	}

	// =========================
	// ✅ PARSE OEM JSON
	// =========================
	for i := range products {
		if len(products[i].OEMsRaw) > 0 {
			if err := json.Unmarshal(products[i].OEMsRaw, &products[i].OEMs); err != nil {
				return nil, 0, err
			}
		} else {
			products[i].OEMs = []OEM{}
		}
	}

	return products, total, nil
}
