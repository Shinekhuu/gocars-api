package models

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const defaultBatchSize = 500 // tune as needed (250 – 1000)

func flattenCategories(list []*Category, out *[]*Category) {
	for _, cat := range list {
		*out = append(*out, cat)
		if len(cat.Children) > 0 {
			flattenCategories(cat.Children, out)
		}
	}
}

func SeedCategoriesFromFile(ctx context.Context, db *gorm.DB, jsonPath string) error {
	f, err := os.Open(jsonPath)
	if err != nil {
		return fmt.Errorf("cannot open JSON file: %w", err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("cannot read JSON: %w", err)
	}

	var root map[string]CategoryJSON
	if err := json.Unmarshal(content, &root); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	categories := ParseCategoryMap(root, nil) // []*Category (with Children hierarchy)
	var flat []*Category
	flattenCategories(categories, &flat)

	// optional: log count
	fmt.Printf("Preparing to upsert %d categories in batches of %d\n", len(flat), defaultBatchSize)

	// batch upserts
	for i := 0; i < len(flat); i += defaultBatchSize {
		end := i + defaultBatchSize
		if end > len(flat) {
			end = len(flat)
		}
		batch := flat[i:end]

		if err := db.WithContext(ctx).
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "category_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"category_name", "level", "thumbnail", "parent_id"}),
			}).
			Create(batch).Error; err != nil {
			return fmt.Errorf("batch upsert error at items %d–%d: %w", i, end, err)
		}
		fmt.Printf("Upserted batch %d–%d\n", i, end)
	}

	return nil
}
