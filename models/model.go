package models

import (
	"encoding/json"
	"fmt"
	"gocars-api/database"
	"gocars-api/utils"
	"io"
	"net/http"
	"os"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ModelResponse struct {
	CountModels int     `json:"countModels"`
	Models      []Model `json:"models"`
}

type Model struct {
	gorm.Model
	ManufacturerID int     `json:"manufacturerId"`
	ModelID        int     `gorm:"uniqueIndex" json:"modelId"` // unique constraint
	ModelName      string  `json:"modelName"`
	ModelYearFrom  string  `json:"modelYearFrom"`
	ModelYearTo    *string `json:"modelYearTo"` // nullable field (can be null)
}

func GetModelByName(manufacturerID int, modelName string, buildDate string) (*Model, error) {
	var dbModel Model
	modelName = strings.ToUpper(modelName)
	modelName = utils.SplitModelName(modelName)

	// 1️⃣ Try fetching from DB first
	err := database.DB.Where("manufacturer_id = ?", manufacturerID).
		Where("model_name LIKE ?", modelName+"%").
		Where("model_year_from <= ?", buildDate).
		Where("(model_year_to >= ? OR (model_year_from <= ? AND model_year_to IS NULL))", buildDate, buildDate).
		Order("model_year_from DESC").
		First(&dbModel).Error

	if err == nil {
		// Found in DB, return it
		return &dbModel, nil
	}

	// 2️⃣ Not found: fetch from RapidAPI
	modelResponse, apiErr := GetModelsFromRapidAPI(manufacturerID)
	if apiErr != nil {
		return nil, fmt.Errorf("DB lookup failed: %v; RapidAPI fetch failed: %v", err, apiErr)
	}

	// 3️⃣ Search fetched data for first matching model
	for i := range modelResponse.Models {
		m := modelResponse.Models[i]

		// Check model name matches and buildDate within range
		if strings.HasPrefix(m.ModelName, modelName) {
			if m.ModelYearFrom <= buildDate && (m.ModelYearTo == nil || *m.ModelYearTo >= buildDate) {
				return &m, nil
			}
		}
	}

	return nil, fmt.Errorf("model not found in DB or RapidAPI for manufacturer %d and name %s", manufacturerID, modelName)
}

func GetModelsFromRapidAPI(manufacturerID int) (*ModelResponse, error) {
	url := fmt.Sprintf(
		"https://tecdoc-catalog.p.rapidapi.com/models/list/type-id/1/manufacturer-id/%d/lang-id/4/country-filter-id/125",
		manufacturerID,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-rapidapi-key", os.Getenv("X_RAPIDAPI_KEY"))
	req.Header.Set("x-rapidapi-host", os.Getenv("X_RAPIDAPI_HOST"))

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var modelResponse ModelResponse
	if err := json.Unmarshal(body, &modelResponse); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	// 2️⃣ Add ManufacturerID to each model before saving
	for i := range modelResponse.Models {
		modelResponse.Models[i].ManufacturerID = manufacturerID
	}

	// 3️⃣ Upsert into DB to avoid duplicates (ModelID is unique)
	if len(modelResponse.Models) > 0 {
		database.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "model_id"}}, // unique key
			UpdateAll: true,                                // update all fields if exists
		}).Create(&modelResponse.Models)
	}

	return &modelResponse, nil
}
