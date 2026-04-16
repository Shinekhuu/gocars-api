package repositories

import (
	"gocars-api/database"
	"gocars-api/models"
)

func FindArticle(id, articleID int) (models.ArticleItem, error) {
	var article models.ArticleItem

	db := database.DB.
		Preload("AllSpecifications").
		Preload("AllOems.Oem")

	if id != 0 {
		db = db.Where("id = ?", id)
	} else {
		db = db.Where("article_id = ?", articleID)
	}

	err := db.First(&article).Error
	return article, err
}

func GetEngines(articleItemID uint, offset, limit int) ([]models.Engine, int64, error) {
	var engines []models.Engine
	var total int64

	database.DB.Model(&models.ArticleVehicles{}).
		Where("article_item_id = ?", articleItemID).
		Count(&total)

	err := database.DB.
		Table("article_vehicles av").
		Select("e.*").
		Joins("LEFT JOIN engines e ON av.vehicle_id = e.vehicle_id").
		Where("av.article_item_id = ?", articleItemID).
		Offset(offset).
		Limit(limit).
		Find(&engines).Error

	return engines, total, err
}

func GetArticleItemsByVehicleIdAndCategoryId(
	vehicleID uint,
	categoryID uint,
	page int,
	limit int,
) (*[]models.ArticleItemWithCategory, int64, error) {

	var results []models.ArticleItemWithCategory
	var total int64

	offset := (page - 1) * limit

	// ==========================
	// 1️⃣ COMMON PART
	// ==========================
	baseCTE := `
		WITH RECURSIVE category_tree AS (
			SELECT category_id FROM categories WHERE category_id = ?
			UNION ALL
			SELECT c.category_id
			FROM categories c
			INNER JOIN category_tree ct ON c.parent_id = ct.category_id
		)
	`

	// ==========================
	// 2️⃣ COUNT QUERY
	// ==========================
	countQuery := baseCTE + `
		SELECT COUNT(*)
		FROM article_items ai
		WHERE EXISTS (
			SELECT 1
			FROM article_categories ac
			JOIN category_tree ct ON ct.category_id = ac.category_id
			WHERE ac.article_item_id = ai.id
		)
	`

	args := []interface{}{categoryID}

	// 👉 vehicle filter dynamic
	if vehicleID != 0 {
		countQuery += `
		AND EXISTS (
			SELECT 1
			FROM article_vehicles av
			WHERE av.article_item_id = ai.id
			AND av.vehicle_id = ?
		)`
		args = append(args, vehicleID)
	}

	if err := database.DB.Raw(countQuery, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// ==========================
	// 3️⃣ ID QUERY
	// ==========================
	idQuery := baseCTE + `
		SELECT ai.id
		FROM article_items ai
		WHERE EXISTS (
			SELECT 1
			FROM article_categories ac
			JOIN category_tree ct ON ct.category_id = ac.category_id
			WHERE ac.article_item_id = ai.id
		)
	`

	args = []interface{}{categoryID}

	if vehicleID != 0 {
		idQuery += `
		AND EXISTS (
			SELECT 1
			FROM article_vehicles av
			WHERE av.article_item_id = ai.id
			AND av.vehicle_id = ?
		)`
		args = append(args, vehicleID)
	}

	idQuery += `
		ORDER BY ai.id DESC
		LIMIT ? OFFSET ?
	`

	args = append(args, limit, offset)

	var articleIDs []uint

	if err := database.DB.Raw(idQuery, args...).Scan(&articleIDs).Error; err != nil {
		return nil, 0, err
	}

	if len(articleIDs) == 0 {
		return &results, total, nil
	}

	// ==========================
	// 4️⃣ DETAIL QUERY
	// ==========================
	err := database.DB.
		Table("article_items AS ai").
		Select(`
			ai.id,
			ai.article_id,
			ai.article_no,
			ai.article_search_no,
			ai.article_product_name,
			ai.supplier_name,
			ai.s3_image,

			c.category_id,
			c.category_name,
			c.category_name_mn,
			c.level,
			c.thumbnail,
			c.parent_id
		`).
		Joins(`
			LEFT JOIN (
				SELECT article_item_id, MAX(id) AS max_id
				FROM article_categories
				GROUP BY article_item_id
			) ac_max 
			ON ac_max.article_item_id = ai.id
		`).
		Joins(`
			LEFT JOIN article_categories ac 
			ON ac.id = ac_max.max_id
		`).
		Joins(`
			LEFT JOIN categories c 
			ON ac.category_id = c.category_id
		`).
		Where("ai.id IN ?", articleIDs).
		Group("ai.id, c.category_id").
		Order("ai.id DESC").
		Find(&results).Error

	if err != nil {
		return nil, 0, err
	}

	return &results, total, nil
}
