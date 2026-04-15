package workers

import (
	"log"

	"gocars-api/database"
	"gocars-api/models"
)

func processArticle(a models.ArticleItem) {

	// 1️⃣ Save main (insert/update article)
	if err := saveMain(a); err != nil {
		log.Println("❌ saveMain failed:", err)
		return
	}

	// 2️⃣ Reload ONLY to get ID
	var dbArticle models.ArticleItem

	if err := database.DB.
		Where("article_id = ?", *a.ArticleID).
		First(&dbArticle).Error; err != nil {

		log.Println("❌ failed to reload article:", err)
		return
	}

	log.Printf("✅ article loaded with ID=%d", dbArticle.ID)

	// 🔥 IMPORTANT FIX: keep API data, only inject ID
	a.ID = dbArticle.ID

	// 3️⃣ Run engines (NO goroutine for debug stability)
	go saveEngines(a)
}
