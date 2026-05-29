package jobs

import (
	"log"

	articles "gocars-api/internal/articles/repository/postgresql/model"
)

var ArticleQueue = make(chan articles.ArticleItem, 100)

func StartWorker() {
	go func() {
		for a := range ArticleQueue {
			log.Println("processing:", a.ArticleID)
			processArticle(a)
		}
	}()
}
