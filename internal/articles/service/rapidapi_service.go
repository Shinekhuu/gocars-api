package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	articles "gocars-api/internal/articles/repository/postgresql/model"
	repo "gocars-api/internal/articles/repository/postgresql"
)

func (s *ArticleService) GetArticleItemsFromRapidAPI(vehicleID uint, categoryID uint) (*articles.VehicleArticlesResponse, error) {
	resp, err := fetchRapidAPI(vehicleID, categoryID)
	if err != nil {
		return nil, err
	}

	go s.PersistFetchedArticles(resp, vehicleID, categoryID)

	return resp, nil
}

func fetchRapidAPI(vehicleID uint, categoryID uint) (*articles.VehicleArticlesResponse, error) {
	url := fmt.Sprintf(
		"https://%s/articles/list/type-id/1/vehicle-id/%d/category-id/%d/lang-id/4",
		os.Getenv("X_RAPIDAPI_HOST"), vehicleID, categoryID,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-rapidapi-key", os.Getenv("X_RAPIDAPI_KEY"))
	req.Header.Set("x-rapidapi-host", os.Getenv("X_RAPIDAPI_HOST"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("rapidapi status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result articles.VehicleArticlesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *ArticleService) PersistFetchedArticles(data *articles.VehicleArticlesResponse, vehicleID uint, categoryID uint) {
	tx := s.db.Begin()

	fetchLogRepoTx := repo.NewAPIFetchLogRepository(tx)

	for i := range data.Articles {
		article := &data.Articles[i]

		if err := tx.
			Where(articles.ArticleItem{ArticleID: article.ArticleID}).
			Assign(article).
			FirstOrCreate(article).Error; err != nil {
			tx.Rollback()
			return
		}

		vehicleRel := articles.ArticleVehicles{
			ArticleItemID: article.ID,
			VehicleID:     vehicleID,
		}
		if err := tx.
			Where("vehicle_id=? AND article_item_id=?", vehicleID, article.ID).
			FirstOrCreate(&vehicleRel).Error; err != nil {
			tx.Rollback()
			return
		}

		categoryRel := articles.ArticleCategory{
			ArticleItemID: article.ID,
			CategoryID:    categoryID,
		}
		if err := tx.
			Where("category_id=? AND article_item_id=?", categoryID, article.ID).
			FirstOrCreate(&categoryRel).Error; err != nil {
			tx.Rollback()
			return
		}
	}

	if err := fetchLogRepoTx.TouchAPIFetchLog(tx, vehicleID, categoryID); err != nil {
		tx.Rollback()
		return
	}

	tx.Commit()
}

func (s *ArticleService) GetByOemFromRapidAPI(oem string) ([]articles.ArticleItem, error) {
	url := fmt.Sprintf(
		"https://%s/artlookup/search-articles-by-article-no?lang-id=4&articleNo=%s&articleType=OENumber",
		os.Getenv("X_RAPIDAPI_HOST"), oem,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-rapidapi-key", os.Getenv("X_RAPIDAPI_KEY"))
	req.Header.Set("x-rapidapi-host", os.Getenv("X_RAPIDAPI_HOST"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var rapidOEMResponse articles.RapidOEMResponse
	if err := json.Unmarshal(body, &rapidOEMResponse); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	if len(rapidOEMResponse.Articles) > 0 {
		go func() {
			if err := s.articleRepo.PersistRapidOemArticles(rapidOEMResponse.Articles, oem); err != nil {
				fmt.Println("Async article & oem save failed:", err)
			}
		}()
	}

	return rapidOEMResponse.Articles, nil
}
