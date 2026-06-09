package jobs

import (
	"log"

	articles "gocars-api/internal/articles/repository/postgresql/model"

	"gorm.io/gorm"
)

var (
	ArticleQueue = make(chan articles.ArticleItem, 100)
	gdb          *gorm.DB
)

func StartWorker(db *gorm.DB) {
	gdb = db
	go func() {
		for a := range ArticleQueue {
			log.Println("processing:", a.ArticleID)
			processArticle(a)
		}
	}()
}
