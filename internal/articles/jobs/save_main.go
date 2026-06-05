package jobs

import (
	"errors"
	"log"

	articles "gocars-api/internal/articles/repository/postgresql/model"

	"gorm.io/gorm"
)

func saveMain(a articles.ArticleItem) error {
	return gdb.Transaction(func(tx *gorm.DB) error {
		var existing articles.ArticleItem
		err := tx.Where("article_id = ?", a.ArticleID).First(&existing).Error

		if err == nil {
			res := tx.Model(&articles.ArticleItem{}).
				Where("id = ?", existing.ID).
				Updates(map[string]interface{}{
					"is_fetched":              true,
					"article_no":              a.ArticleNo,
					"article_product_name":    a.ArticleProductName,
					"supplier_id":             a.SupplierID,
					"supplier_name":           a.SupplierName,
					"product_id":              a.ProductID,
					"article_media_type":      a.ArticleMediaType,
					"article_media_file_name": a.ArticleMediaFileName,
					"s3_image":                a.S3Image,
				})
			if res.Error != nil {
				return res.Error
			}
			log.Println("Rows:", res.RowsAffected)
			a.ID = existing.ID

		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			a.IsFetched = true
			if err := tx.Create(&a).Error; err != nil {
				return err
			}
		} else {
			return err
		}

		for _, s := range a.AllSpecifications {
			spec := articles.ArticleAllSpecification{
				ArticleItemID: a.ID,
				CriteriaName:  s.CriteriaName,
				CriteriaValue: s.CriteriaValue,
			}
			tx.Where("article_item_id=? AND criteria_name=? AND criteria_value=?",
				a.ID, s.CriteriaName, s.CriteriaValue).
				FirstOrCreate(&spec)
		}

		for _, o := range a.OemResponses {
			oem := articles.Oem{Brand: o.Brand, DisplayNo: o.DisplayNo}
			tx.Where("brand=? AND display_no=?", o.Brand, o.DisplayNo).
				FirstOrCreate(&oem)

			link := articles.ArticleOem{ArticleItemID: a.ID, OemID: oem.ID}
			tx.Where("article_item_id=? AND oem_id=?", a.ID, oem.ID).
				FirstOrCreate(&link)
		}

		return nil
	})
}
