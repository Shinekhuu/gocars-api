package services

import (
	"encoding/json"
	"fmt"
	"gocars-api/database"
	"gocars-api/models"
	"gocars-api/repositories"
	"io"
	"net/http"
	"os"
)

func GetArticleItemsFromRapidAPI(
	vehicleID uint,
	categoryID uint,
) (
	*models.VehicleArticlesResponse,
	error,
) {

	resp, err :=
		fetchRapidAPI(
			vehicleID,
			categoryID,
		)

	if err != nil {
		return nil, err
	}

	go PersistFetchedArticles(
		resp,
		vehicleID,
		categoryID,
	)

	return resp, nil
}

func fetchRapidAPI(
	vehicleID uint,
	categoryID uint,
) (
	*models.VehicleArticlesResponse,
	error,
) {

	url := fmt.Sprintf(
		"https://auto-parts-catalog.p.rapidapi.com/articles/list/type-id/1/vehicle-id/%d/category-id/%d/lang-id/4",
		vehicleID,
		categoryID,
	)

	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"x-rapidapi-key",
		os.Getenv("X_RAPIDAPI_KEY"),
	)

	req.Header.Set(
		"x-rapidapi-host",
		os.Getenv("X_RAPIDAPI_HOST"),
	)

	resp, err :=
		http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil,
			fmt.Errorf(
				"rapidapi status %d",
				resp.StatusCode,
			)
	}

	body, err := io.ReadAll(
		resp.Body,
	)

	if err != nil {
		return nil, err
	}

	var result models.VehicleArticlesResponse

	err = json.Unmarshal(
		body,
		&result,
	)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func PersistFetchedArticles(
	data *models.VehicleArticlesResponse,
	vehicleID uint,
	categoryID uint,
) {

	tx := database.DB.Begin()

	for i := range data.Articles {

		article :=
			&data.Articles[i]

		err := tx.
			Where(
				models.ArticleItem{
					ArticleID: article.ArticleID,
				},
			).
			Assign(article).
			FirstOrCreate(article).
			Error

		if err != nil {
			tx.Rollback()
			return
		}

		vehicleRel :=
			models.ArticleVehicles{
				ArticleItemID: article.ID,
				VehicleID:     vehicleID,
			}

		err = tx.
			Where(
				"vehicle_id=? AND article_item_id=?",
				vehicleID,
				article.ID,
			).
			FirstOrCreate(
				&vehicleRel,
			).Error

		if err != nil {
			tx.Rollback()
			return
		}

		categoryRel :=
			models.ArticleCategory{
				ArticleItemID: article.ID,
				CategoryID:    categoryID,
			}

		err = tx.
			Where(
				"category_id=? AND article_item_id=?",
				categoryID,
				article.ID,
			).
			FirstOrCreate(
				&categoryRel,
			).Error

		if err != nil {
			tx.Rollback()
			return
		}
	}

	err := repositories.TouchAPIFetchLog(
		tx,
		vehicleID,
		categoryID,
	)

	if err != nil {
		tx.Rollback()
		return
	}

	tx.Commit()
}
