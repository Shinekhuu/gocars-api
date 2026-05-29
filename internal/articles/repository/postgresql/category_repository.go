package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	articles "gocars-api/internal/articles/repository/postgresql/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) GetCategoryTreeIDs(categoryID uint) ([]uint, error) {
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

	if err := r.db.Raw(query, categoryID).Scan(&categoryIDs).Error; err != nil {
		return nil, err
	}

	return categoryIDs, nil
}

func (r *CategoryRepository) SaveCategoryRecursive(list []*articles.Category) error {
	for _, cat := range list {
		if err := r.db.
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "category_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"category_name", "level", "thumbnail", "parent_id"}),
			}).
			Create(cat).Error; err != nil {
			return err
		}
		if len(cat.Children) > 0 {
			if err := r.SaveCategoryRecursive(cat.Children); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *CategoryRepository) UpdateCategoryNamesBatch(categoryMap map[uint]string) error {
	if len(categoryMap) == 0 {
		return nil
	}
	query := "UPDATE categories SET category_name_mn = CASE category_id "
	args := make([]interface{}, 0, len(categoryMap)*2)
	ids := make([]uint, 0, len(categoryMap))
	for id, name := range categoryMap {
		query += "WHEN ? THEN ? "
		args = append(args, id, name)
		ids = append(ids, id)
	}
	query += "END WHERE category_id IN ?"
	args = append(args, ids)
	return r.db.Exec(query, args...).Error
}

const seedBatchSize = 500

func (r *CategoryRepository) SeedFromFile(ctx context.Context, jsonPath string) error {
	f, err := os.Open(jsonPath)
	if err != nil {
		return fmt.Errorf("cannot open JSON file: %w", err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("cannot read JSON: %w", err)
	}

	var root map[string]articles.CategoryJSON
	if err := json.Unmarshal(content, &root); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	categories := articles.ParseCategoryMap(root, nil)
	var flat []*articles.Category
	flattenCategories(categories, &flat)

	fmt.Printf("Preparing to upsert %d categories in batches of %d\n", len(flat), seedBatchSize)

	for i := 0; i < len(flat); i += seedBatchSize {
		end := i + seedBatchSize
		if end > len(flat) {
			end = len(flat)
		}
		batch := flat[i:end]

		if err := r.db.WithContext(ctx).
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "category_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"category_name", "level", "thumbnail", "parent_id"}),
			}).
			Create(batch).Error; err != nil {
			return fmt.Errorf("batch upsert error at items %d-%d: %w", i, end, err)
		}
		fmt.Printf("Upserted batch %d-%d\n", i, end)
	}

	return nil
}

func flattenCategories(list []*articles.Category, out *[]*articles.Category) {
	for _, cat := range list {
		*out = append(*out, cat)
		if len(cat.Children) > 0 {
			flattenCategories(cat.Children, out)
		}
	}
}
