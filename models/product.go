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
	// ✅ BASE SUBQUERY
	// =========================
	baseSubQuery := `
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

		GROUP_CONCAT(o.display_no_clean, ' ', o.brand) AS oems_text

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
	`

	baseArgs := []interface{}{}

	// 🚗 vehicle optional
	if filter.VehicleID != nil {
		baseSubQuery += `
		AND EXISTS (
			SELECT 1
			FROM article_vehicles av
			WHERE av.article_item_id = ai.id
			AND av.vehicle_id = ?
		)`
		baseArgs = append(baseArgs, *filter.VehicleID)
	}

	baseSubQuery += `
	GROUP BY ai.id, c.category_id
	`

	// =========================
	// ✅ COUNT QUERY
	// =========================
	countQuery := `
	SELECT COUNT(*) FROM (
		` + baseSubQuery + `
	) AS results

	WHERE (
		? = '' OR
		results.article_product_name LIKE ? OR
		results.category_name LIKE ? OR
		results.category_name_mn LIKE ? OR
		results.supplier_name LIKE ? OR
		results.oems_text LIKE ? OR
		results.oems_text LIKE CONCAT('%', REGEXP_REPLACE(?, '[^[:alnum:]]', ''), '%')
	)
	`

	countArgs := append([]interface{}{}, baseArgs...)
	countArgs = append(countArgs,
		search,
		like, like, like, like,
		like,
		search, // 🔥 normalized OEM
	)

	if err := database.DB.Raw(countQuery, countArgs...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// =========================
	// ✅ MAIN QUERY
	// =========================
	mainQuery := `
	SELECT
		results.id,
		results.article_id,
		results.article_no,
		results.article_search_no,
		results.article_product_name,
		results.s3_image,
		results.supplier_name,

		results.category_id,
		results.category_name,
		results.category_name_mn,
		results.level,
		results.thumbnail,
		results.parent_id,

		results.oems,

		CASE 
			WHEN ? != '' AND results.oems_text LIKE CONCAT('%', REGEXP_REPLACE(?, '[^[:alnum:]]', ''), '%') THEN 1
			WHEN ? != '' AND results.oems_text LIKE ? THEN 2
			WHEN ? != '' AND results.category_name LIKE ? THEN 3
			WHEN ? != '' AND results.category_name_mn LIKE ? THEN 4
			WHEN ? != '' AND results.article_product_name LIKE ? THEN 5
			WHEN ? != '' AND results.supplier_name LIKE ? THEN 6
			ELSE 7
		END AS priority

	FROM (
		` + baseSubQuery + `
	) AS results

	WHERE (
		? = '' OR
		results.article_product_name LIKE ? OR
		results.category_name LIKE ? OR
		results.category_name_mn LIKE ? OR
		results.supplier_name LIKE ? OR
		results.oems_text LIKE ? OR
		results.oems_text LIKE CONCAT('%', REGEXP_REPLACE(?, '[^[:alnum:]]', ''), '%')
	)

	ORDER BY priority ASC, id DESC
	LIMIT ? OFFSET ?
	`

	mainArgs := []interface{}{}

	// 1️⃣ priority
	mainArgs = append(mainArgs,
		search, search, // normalized OEM
		search, like,
		search, like,
		search, like,
		search, like,
		search, like,
	)

	// 2️⃣ vehicle
	mainArgs = append(mainArgs, baseArgs...)

	// 3️⃣ WHERE
	mainArgs = append(mainArgs,
		search,
		like, like, like, like,
		like,
		search, // normalized OEM
	)

	// 4️⃣ pagination
	mainArgs = append(mainArgs,
		filter.Limit,
		offset,
	)

	if err := database.DB.Raw(mainQuery, mainArgs...).Scan(&products).Error; err != nil {
		return nil, 0, err
	}

	if products == nil {
		products = []Product{}
	}

	// OEM parse
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
