package repositories

import "gocars-api/database"

func GetCategoryTreeIDs(categoryID uint) ([]uint, error) {
	var categoryIDs []uint

	query := `
		WITH RECURSIVE category_tree AS (
			SELECT category_id FROM categories WHERE category_id = ?
			UNION ALL
			SELECT c.category_id
			FROM categories c
			INNER JOIN category_tree ct ON c.parent_id = ct.category_id
		)
		SELECT category_id FROM category_tree
	`

	if err := database.DB.Raw(query, categoryID).Scan(&categoryIDs).Error; err != nil {
		return nil, err
	}

	return categoryIDs, nil
}
