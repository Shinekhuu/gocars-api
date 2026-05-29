package repository

import (
	articles "gocars-api/internal/articles/repository/postgresql/model"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

type ProductFilter struct {
	Search     *string
	CategoryID *uint
	Page       int
	Limit      int
}

func (r *ProductRepository) GetProducts(filter ProductFilter) ([]articles.Product, int64, error) {
	var products []articles.Product
	var total int64

	offset := (filter.Page - 1) * filter.Limit

	countQuery := `
	SELECT COUNT(DISTINCT ai.id)
	FROM article_items ai
	LEFT JOIN (
		SELECT article_item_id, MAX(id) AS max_id
		FROM article_categories
		GROUP BY article_item_id
	) ac_max ON ac_max.article_item_id = ai.id
	LEFT JOIN article_categories ac ON ac.id = ac_max.max_id
	LEFT JOIN categories c ON ac.category_id = c.category_id
	WHERE 1=1
	`

	args := []interface{}{}

	if filter.Search != nil {
		countQuery += " AND ai.article_no LIKE ?"
		args = append(args, "%"+*filter.Search+"%")
	}

	if filter.CategoryID != nil {
		countQuery += " AND c.category_id = ?"
		args = append(args, *filter.CategoryID)
	}

	if err := r.db.Raw(countQuery, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	mainQuery := `
	SELECT
		ai.id,
		ai.article_id,
		ai.article_no,
		ai.article_search_no,
		ai.article_product_name,
		ai.s3_image,
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
		END AS oems
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
			REGEXP_REPLACE(o.display_no, '[^A-Za-z0-9]', '') AS display_no_clean
		FROM article_oems ao
		JOIN oems o ON ao.oem_id = o.id
		WHERE o.display_no IS NOT NULL
	) o ON o.article_item_id = ai.id
	WHERE 1=1
	`

	mainArgs := []interface{}{}

	if filter.Search != nil {
		mainQuery += " AND ai.article_no LIKE ?"
		mainArgs = append(mainArgs, "%"+*filter.Search+"%")
	}

	if filter.CategoryID != nil {
		mainQuery += " AND c.category_id = ?"
		mainArgs = append(mainArgs, *filter.CategoryID)
	}

	mainQuery += `
	GROUP BY ai.id, c.category_id
	ORDER BY ai.id DESC
	LIMIT ? OFFSET ?
	`

	mainArgs = append(mainArgs, filter.Limit, offset)

	if err := r.db.Raw(mainQuery, mainArgs...).Scan(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
