package models

import (
	"encoding/json"
	"fmt"
	"gocars-api/database"
	"os"

	"gorm.io/gorm/clause"
)

type Category struct {
	CategoryID     uint   `json:"categoryId" gorm:"column:category_id;primaryKey"`
	CategoryName   string `json:"categoryName"  gorm:"type:varchar(255)"`
	CategoryNameMn string `json:"categoryNameMn"  gorm:"column:category_name_mn;type:varchar(255)"`
	Level          int    `json:"level"`
	Thumbnail      string `json:"thumbnail"  gorm:"type:text"`

	// Self-referencing FK
	ParentID *uint `json:"parentId" gorm:"column:parent_id;index"`

	// Parent category (for easy access)
	Parent *Category `json:"parent,omitempty" gorm:"foreignKey:ParentID;references:CategoryID"`

	// Children of this category
	Children []*Category `json:"children" gorm:"foreignKey:ParentID;references:CategoryID"`
}

type CategoryJSON struct {
	CategoryID   uint             `json:"categoryId"`
	CategoryName string           `json:"categoryName"`
	Level        int              `json:"level"`
	Thumbnail    string           `json:"thumbnail"`
	Children     CategoryChildren `json:"children"`
}

type CategoryChildren map[string]CategoryJSON

func (c *CategoryChildren) UnmarshalJSON(data []byte) error {
	// Case 1: empty array → children = empty map
	if string(data) == "[]" {
		*c = make(map[string]CategoryJSON)
		return nil
	}

	// Try to unmarshal as map
	var asMap map[string]CategoryJSON
	if err := json.Unmarshal(data, &asMap); err == nil {
		*c = asMap
		return nil
	}

	// Try to unmarshal as array (ignore elements)
	var asArray []CategoryJSON
	if err := json.Unmarshal(data, &asArray); err == nil {
		*c = make(map[string]CategoryJSON)
		return nil
	}

	return fmt.Errorf("children must be object or array, got: %s", string(data))
}

func ParseCategoryMap(m map[string]CategoryJSON, parent *Category) []*Category {
	categories := make([]*Category, 0, len(m))

	for _, v := range m {
		cat := &Category{
			CategoryID:   v.CategoryID,
			CategoryName: v.CategoryName,
			Level:        v.Level,
			Thumbnail:    v.Thumbnail,
		}

		if parent != nil {
			cat.ParentID = &parent.CategoryID
		}

		if len(v.Children) > 0 {
			cat.Children = ParseCategoryMap(v.Children, cat)
		}

		categories = append(categories, cat)
	}

	return categories
}

func SaveCategoryRecursive(list []*Category) error {
	for _, cat := range list {
		// Upsert by CategoryID
		if err := database.DB.
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "category_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"category_name", "level", "thumbnail", "parent_id"}),
			}).
			Create(cat).Error; err != nil {
			return err
		}

		// Save children recursively
		if len(cat.Children) > 0 {
			if err := SaveCategoryRecursive(cat.Children); err != nil {
				return err
			}
		}
	}

	return nil
}

func SeedCategories(jsonPath string) error {
	content, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("cannot read JSON: %w", err)
	}

	var root map[string]CategoryJSON
	if err := json.Unmarshal(content, &root); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	categories := ParseCategoryMap(root, nil) // []*Category
	return SaveCategoryRecursive(categories)
}

func flattenCategoriesMN(cat *CategoryJSON, result map[uint]string) {
	if cat == nil {
		return
	}

	// Optional: skip empty names
	if cat.CategoryID != 0 && cat.CategoryName != "" {
		result[cat.CategoryID] = cat.CategoryName
	}

	for _, child := range cat.Children {
		flattenCategoriesMN(&child, result)
	}
}

func LoadCategoryMap(filePath string) (map[uint]string, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var raw map[string]*CategoryJSON
	if err := json.Unmarshal(file, &raw); err != nil {
		return nil, err
	}

	result := make(map[uint]string, 1000) // small optimization

	for _, root := range raw {
		flattenCategoriesMN(root, result)
	}

	return result, nil
}

func UpdateCategoryNamesBatch(categoryMap map[uint]string) error {
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

	// ✅ FIX HERE
	query += "END WHERE category_id IN ?"

	// combine args
	args = append(args, ids)

	return database.DB.Exec(query, args...).Error
}
