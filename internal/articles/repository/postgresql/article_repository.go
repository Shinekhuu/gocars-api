package repository

import (
	"encoding/json"
	"log"

	articles "gocars-api/internal/articles/repository/postgresql/model"
	vehicle "gocars-api/internal/vehicle/repository/postgresql/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) FindArticle(id, articleID int) (articles.ArticleItem, error) {
	var article articles.ArticleItem

	q := r.db.
		Preload("AllSpecifications").
		Preload("AllOems.Oem")

	if id != 0 {
		q = q.Where("id = ?", id)
	} else {
		q = q.Where("article_id = ?", articleID)
	}

	err := q.First(&article).Error
	return article, err
}

func (r *ArticleRepository) UpdateArticleDictionary(productName string, dictionaryID uint) error {
	return r.db.
		Model(&articles.ArticleItem{}).
		Where("article_product_name = ?", productName).
		Update("dictionary_id", dictionaryID).Error
}

func (r *ArticleRepository) GetEngines(articleItemID uint, offset, limit int) ([]vehicle.Engine, int64, error) {
	var engines []vehicle.Engine
	var total int64

	r.db.Model(&articles.ArticleVehicles{}).
		Where("article_item_id = ?", articleItemID).
		Count(&total)

	err := r.db.
		Table("article_vehicles av").
		Select("e.*").
		Joins("LEFT JOIN engines e ON av.vehicle_id = e.vehicle_id").
		Where("av.article_item_id = ?", articleItemID).
		Offset(offset).
		Limit(limit).
		Find(&engines).Error

	return engines, total, err
}

func (r *ArticleRepository) GetArticleItemsByVehicleIdAndCategoryId(
	vehicleID uint,
	categoryID uint,
	page int,
	limit int,
) (*[]articles.ArticleItemWithCategory, int64, error) {

	var results []articles.ArticleItemWithCategory
	var total int64

	offset := (page - 1) * limit

	baseCTE := `
		WITH RECURSIVE category_tree AS (
			SELECT category_id FROM categories WHERE category_id = ?
			UNION ALL
			SELECT c.category_id
			FROM categories c
			INNER JOIN category_tree ct ON c.parent_id = ct.category_id
		)
	`

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

	if err := r.db.Raw(countQuery, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

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
	if err := r.db.Raw(idQuery, args...).Scan(&articleIDs).Error; err != nil {
		return nil, 0, err
	}

	if len(articleIDs) == 0 {
		return &results, total, nil
	}

	err := r.db.
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
		Joins(`LEFT JOIN article_categories ac ON ac.id = ac_max.max_id`).
		Joins(`LEFT JOIN categories c ON ac.category_id = c.category_id`).
		Where("ai.id IN ?", articleIDs).
		Group("ai.id, c.category_id").
		Order("ai.id DESC").
		Find(&results).Error

	if err != nil {
		return nil, 0, err
	}

	return &results, total, nil
}

func (r *ArticleRepository) UpdateArticle(article *articles.ArticleItem) error {
	return r.db.Save(article).Error
}

func (r *ArticleRepository) FetchArticlesForIndex(offset, limit int) ([]articles.ArticleItemWithCategory, error) {
	var rows []articles.ArticleItemWithCategory

	err := r.db.
		Table("article_items AS ai").
		Select(`
			ai.id,
			ai.article_no,
			ai.article_search_no,
			ai.article_product_name,
			ai.supplier_name,
			ai.s3_image,

			c.category_id,
			c.category_name,
			c.category_name_mn
		`).
		Joins(`LEFT JOIN article_categories ac ON ac.article_item_id = ai.id`).
		Joins(`LEFT JOIN categories c ON ac.category_id = c.category_id`).
		Order("ai.id ASC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error

	return rows, err
}

func (r *ArticleRepository) GetByOem(oem string, page, limit int) (*[]articles.ArticleItem, int64, error) {
	var items []articles.ArticleItem
	var total int64

	matchExpr := "UPPER(REGEXP_REPLACE(oems.display_no, '[^A-Za-z0-9]', '', 'g')) = UPPER(REGEXP_REPLACE(?, '[^A-Za-z0-9]', '', 'g'))"

	if err := r.db.
		Table("oems").
		Joins("LEFT JOIN article_oems ON oems.id = article_oems.oem_id").
		Joins("LEFT JOIN article_items ON article_oems.article_item_id = article_items.id").
		Where(matchExpr, oem).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	if err := r.db.
		Table("oems").
		Select("article_items.*").
		Joins("LEFT JOIN article_oems ON oems.id = article_oems.oem_id").
		Joins("LEFT JOIN article_items ON article_oems.article_item_id = article_items.id").
		Where(matchExpr, oem).
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return &items, total, nil
}

func (r *ArticleRepository) SaveArticleItem(article *articles.ArticleItem, vehicleIDs []uint, categories string) (*articles.ArticleItem, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Model(&articles.ArticleItem{}).
			Select("id").
			Where("article_no = ?", article.ArticleNo).
			Limit(1).
			Scan(&article.ID).Error; err != nil {
			return err
		}

		exists := article.ID != 0

		oem := articles.Oem{
			Brand:     article.SupplierName,
			DisplayNo: article.ArticleNo,
		}
		if err := tx.
			Where(articles.Oem{Brand: oem.Brand, DisplayNo: oem.DisplayNo}).
			FirstOrCreate(&oem).Error; err != nil {
			return err
		}

		if !exists {
			if err := tx.Create(article).Error; err != nil {
				return err
			}
		}

		if err := tx.
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "article_item_id"}, {Name: "oem_id"}},
				DoNothing: true,
			}).
			Create(&articles.ArticleOem{ArticleItemID: article.ID, OemID: oem.ID}).Error; err != nil {
			return err
		}

		if len(vehicleIDs) > 0 {
			seen := make(map[uint]struct{}, len(vehicleIDs))
			vehicles := make([]articles.ArticleVehicles, 0, len(vehicleIDs))
			for _, vid := range vehicleIDs {
				if vid == 0 {
					continue
				}
				if _, ok := seen[vid]; ok {
					continue
				}
				seen[vid] = struct{}{}
				vehicles = append(vehicles, articles.ArticleVehicles{ArticleItemID: article.ID, VehicleID: vid})
			}
			if len(vehicles) > 0 {
				if err := tx.
					Clauses(clause.OnConflict{
						Columns:   []clause.Column{{Name: "article_item_id"}, {Name: "vehicle_id"}},
						DoNothing: true,
					}).
					Create(&vehicles).Error; err != nil {
					return err
				}
			}
		}

		if categories != "" && categories != "null" {
			var categoryIDs []uint
			if err := json.Unmarshal([]byte(categories), &categoryIDs); err != nil {
				log.Println("invalid categories:", categories)
				return err
			}
			if len(categoryIDs) > 0 {
				seen := make(map[uint]struct{}, len(categoryIDs))
				cats := make([]articles.ArticleCategory, 0, len(categoryIDs))
				for _, cid := range categoryIDs {
					if cid == 0 {
						continue
					}
					if _, ok := seen[cid]; ok {
						continue
					}
					seen[cid] = struct{}{}
					cats = append(cats, articles.ArticleCategory{ArticleItemID: article.ID, CategoryID: cid})
				}
				if len(cats) > 0 {
					if err := tx.
						Clauses(clause.OnConflict{
							Columns:   []clause.Column{{Name: "article_item_id"}, {Name: "category_id"}},
							DoNothing: true,
						}).
						Create(&cats).Error; err != nil {
						return err
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		log.Println(article.ArticleNo, "Save failed:", err)
		return nil, err
	}

	return article, nil
}

func (r *ArticleRepository) SearchProducts(filter articles.ProductFilter) ([]articles.Product, int64, error) {
	var products []articles.Product
	var total int64

	offset := (filter.Page - 1) * filter.Limit

	search := ""
	if filter.Search != nil {
		search = *filter.Search
	}
	like := "%" + search + "%"

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
		COALESCE(
			jsonb_agg(jsonb_build_object('displayNo', o.display_no_clean, 'brand', o.brand))
			FILTER (WHERE o.display_no_clean IS NOT NULL),
			'[]'::jsonb
		) AS oems,
		string_agg(o.display_no_clean || ' ' || o.brand, ' ') AS oems_text
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
			REGEXP_REPLACE(o.display_no, '[^A-Za-z0-9]', '', 'g') AS display_no_clean
		FROM article_oems ao
		INNER JOIN oems o ON ao.oem_id = o.id
		WHERE o.display_no IS NOT NULL
	) o ON o.article_item_id = ai.id
	WHERE 1=1
	`

	baseArgs := []interface{}{}

	if filter.VehicleID != nil {
		baseSubQuery += `
		AND EXISTS (
			SELECT 1 FROM article_vehicles av
			WHERE av.article_item_id = ai.id AND av.vehicle_id = ?
		)`
		baseArgs = append(baseArgs, *filter.VehicleID)
	}

	baseSubQuery += "\nGROUP BY ai.id, c.category_id\n"

	searchCond := `(
		? = '' OR
		results.article_product_name ILIKE ? OR
		results.category_name ILIKE ? OR
		results.category_name_mn ILIKE ? OR
		results.supplier_name ILIKE ? OR
		results.oems_text ILIKE ? OR
		results.oems_text ILIKE '%' || REGEXP_REPLACE(?, '[^[:alnum:]]', '', 'g') || '%'
	)`
	searchArgs := []interface{}{search, like, like, like, like, like, search}

	countQuery := "SELECT COUNT(*) FROM (\n" + baseSubQuery + "\n) AS results\nWHERE " + searchCond
	countArgs := append(append([]interface{}{}, baseArgs...), searchArgs...)
	if err := r.db.Raw(countQuery, countArgs...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	priorityExpr := `
	CASE
		WHEN ? != '' AND results.oems_text ILIKE '%' || REGEXP_REPLACE(?, '[^[:alnum:]]', '', 'g') || '%' THEN 1
		WHEN ? != '' AND results.oems_text ILIKE ? THEN 2
		WHEN ? != '' AND results.category_name ILIKE ? THEN 3
		WHEN ? != '' AND results.category_name_mn ILIKE ? THEN 4
		WHEN ? != '' AND results.article_product_name ILIKE ? THEN 5
		WHEN ? != '' AND results.supplier_name ILIKE ? THEN 6
		ELSE 7
	END AS priority`
	priorityArgs := []interface{}{search, search, search, like, search, like, search, like, search, like, search, like}

	mainQuery := "SELECT results.*, " + priorityExpr + " FROM (\n" + baseSubQuery + "\n) AS results\nWHERE " + searchCond + "\nORDER BY priority ASC, id DESC\nLIMIT ? OFFSET ?"
	mainArgs := append(append([]interface{}{}, priorityArgs...), baseArgs...)
	mainArgs = append(mainArgs, searchArgs...)
	mainArgs = append(mainArgs, filter.Limit, offset)

	if err := r.db.Raw(mainQuery, mainArgs...).Scan(&products).Error; err != nil {
		return nil, 0, err
	}

	if products == nil {
		products = []articles.Product{}
	}

	for i := range products {
		if len(products[i].OEMsRaw) > 0 {
			if err := json.Unmarshal(products[i].OEMsRaw, &products[i].OEMs); err != nil {
				products[i].OEMs = []articles.OEM{}
			}
		} else {
			products[i].OEMs = []articles.OEM{}
		}
	}

	return products, total, nil
}

func (r *ArticleRepository) PersistRapidOemArticles(rapidArticles []articles.ArticleItem, oem string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i := range rapidArticles {
			article := rapidArticles[i]

			if err := tx.
				Where("article_id = ?", article.ArticleID).
				Assign(article).
				FirstOrCreate(&article).Error; err != nil {
				return err
			}

			newOem := articles.Oem{Brand: article.SupplierName, DisplayNo: oem}
			if err := tx.
				Where("brand = ? AND display_no = ?", newOem.Brand, newOem.DisplayNo).
				Assign(&newOem).
				FirstOrCreate(&newOem).Error; err != nil {
				return err
			}

			link := articles.ArticleOem{ArticleItemID: article.ID, OemID: newOem.ID}
			if err := tx.
				Where("article_item_id = ? AND oem_id = ?", link.ArticleItemID, link.OemID).
				Assign(&link).
				FirstOrCreate(&link).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
