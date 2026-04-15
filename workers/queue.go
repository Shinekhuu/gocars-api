package workers

import (
	"log"

	"gocars-api/models"
)

var ArticleQueue = make(chan models.ArticleItem, 100)

func StartWorker() {
	go func() {
		for a := range ArticleQueue {
			log.Println("🟢 processing:", a.ArticleID)
			processArticle(a)
		}
	}()
}
