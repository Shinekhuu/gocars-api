package model

import (
	"encoding/json"
	"fmt"
	"os"
)

type Category struct {
	CategoryID     uint   `json:"categoryId" gorm:"column:category_id;primaryKey"`
	CategoryName   string `json:"categoryName" gorm:"type:varchar(255)"`
	CategoryNameMn string `json:"categoryNameMn" gorm:"column:category_name_mn;type:varchar(255)"`
	Level          int    `json:"level"`
	Thumbnail      string `json:"thumbnail" gorm:"type:text"`

	ParentID *uint `json:"parentId" gorm:"column:parent_id;index"`

	Parent   *Category  `json:"parent,omitempty" gorm:"foreignKey:ParentID;references:CategoryID"`
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
	if string(data) == "[]" {
		*c = make(map[string]CategoryJSON)
		return nil
	}
	var asMap map[string]CategoryJSON
	if err := json.Unmarshal(data, &asMap); err == nil {
		*c = asMap
		return nil
	}
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

func LoadCategoryMap(filePath string) (map[uint]string, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var raw map[string]*CategoryJSON
	if err := json.Unmarshal(file, &raw); err != nil {
		return nil, err
	}
	result := make(map[uint]string, 1000)
	for _, root := range raw {
		flattenCategoriesMN(root, result)
	}
	return result, nil
}

func flattenCategoriesMN(cat *CategoryJSON, result map[uint]string) {
	if cat == nil {
		return
	}
	if cat.CategoryID != 0 && cat.CategoryName != "" {
		result[cat.CategoryID] = cat.CategoryName
	}
	for _, child := range cat.Children {
		flattenCategoriesMN(&child, result)
	}
}
