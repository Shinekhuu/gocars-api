package models

import (
	"encoding/json"
	"fmt"
	"gocars-api/database"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EngineResponse struct {
	ModelType       string   `json:"modelType"`
	CountModelTypes int      `json:"countModelTypes"`
	Engines         []Engine `json:"ModelTypes"`
}

type Engine struct {
	gorm.Model
	ManufacturerID            uint           `json:"manufacturerId"`
	ModelID                   uint           `json:"modelId"`
	VehicleID                 uint           `gorm:"uniqueIndex" json:"vehicleId"`
	ManufacturerName          string         `json:"manufacturerName"`
	ModelName                 string         `json:"modelName"`
	TypeEngineName            string         `json:"typeEngineName"`
	ConstructionIntervalStart string         `json:"constructionIntervalStart"`
	ConstructionIntervalEnd   string         `json:"constructionIntervalEnd"`
	PowerKw                   string         `json:"powerKw"`
	PowerPs                   string         `json:"powerPs"`
	CapacityTax               *string        `json:"capacityTax"` // nullable
	FuelType                  string         `json:"fuelType"`
	BodyType                  string         `json:"bodyType"`
	NumberOfCylinders         int            `json:"numberOfCylinders"`
	CapacityLt                string         `json:"capacityLt"`
	CapacityTech              string         `json:"capacityTech"`
	EngineCodes               string         `json:"engineCodes"`
	EngID                     uint           `json:"engId"`
	IsFetched                 bool           `gorm:"type:tinyint(1);default:0"`
	Families                  []EngineFamily `gorm:"foreignKey:EngineID;constraint:OnDelete:CASCADE;" json:"families"`
}

type EngineFamily struct {
	gorm.Model
	EngineID   uint   `gorm:"column:engine_id" json:"engineId"`
	FamilyCode string `gorm:"column:family_code;type:varchar(50)" json:"familyCode"`
}

func GetEngineByName(manufacturerID uint, modelID uint, frame string) (*[]Engine, error) {
	frame = strings.ToUpper(frame)
	var dbEngines []Engine

	// 1️⃣ Try fetching from DB first
	err := database.DB.Where("manufacturer_id = ?", manufacturerID).
		Where("model_id = ?", modelID).
		Where("UPPER(type_engine_name) LIKE ?", "%"+frame+"%").
		Find(&dbEngines).Error

	if err == nil && len(dbEngines) > 0 {
		// Found in DB, return it
		return &dbEngines, nil
	}

	// 2️⃣ Not found: fetch from RapidAPI
	engineResponse, apiErr := GetEnginesFromRapidAPI(manufacturerID, modelID)
	if apiErr != nil {
		return nil, fmt.Errorf("DB lookup failed: %v; RapidAPI fetch failed: %v", err, apiErr)
	}

	// 3️⃣ Filter fetched data for engines matching frame
	var matchedEngines []Engine
	for i := range engineResponse.Engines {
		e := engineResponse.Engines[i]
		name := strings.ToUpper(e.TypeEngineName)

		if strings.Contains(name, frame) {
			matchedEngines = append(matchedEngines, e)
		}
	}

	if len(matchedEngines) > 0 {
		return &matchedEngines, nil
	}

	return nil, fmt.Errorf("engine not found in DB or RapidAPI for manufacturer %d, model %d, frame %s", manufacturerID, modelID, frame)
}

func GetEnginesFromRapidAPI(ManufacturerID uint, ModelID uint) (*EngineResponse, error) {
	url := fmt.Sprintf(
		"https://tecdoc-catalog.p.rapidapi.com/types/type-id/1/list-vehicles-types/%d/lang-id/4/country-filter-id/125",
		ModelID,
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

	var engineResponse EngineResponse
	if err := json.Unmarshal(body, &engineResponse); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	// 2️⃣ Add ManufacturerID and ModelID to each model before saving
	for i := range engineResponse.Engines {
		engineResponse.Engines[i].ManufacturerID = ManufacturerID
		engineResponse.Engines[i].ModelID = ModelID
		engineResponse.Engines[i].IsFetched = true
	}

	go func() {
		// 3️⃣ Upsert into DB to avoid duplicates (ModelID is unique)
		if len(engineResponse.Engines) > 0 {
			database.DB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "vehicle_id"}}, // unique key
				UpdateAll: true,                                  // update all fields if exists
			}).Create(&engineResponse.Engines)
		}
	}()

	return &engineResponse, nil
}

func GetEnginesByModelId(modelID uint) (*EngineResponse, error) {
	// 1️⃣ Try to load from database
	var engineResponse EngineResponse
	if err := database.DB.Where("model_id = ? AND is_fetched = ?", modelID, true).Find(&engineResponse.Engines).Error; err != nil {
		return nil, err
	}

	// 2️⃣ Set count
	engineResponse.CountModelTypes = len(engineResponse.Engines)

	// 3️⃣ Return result (even if empty)
	return &engineResponse, nil
}

var nonAlnumRegex = regexp.MustCompile(`[^\p{L}\p{N}]`)

// =========================
// DB QUERY
// =========================
func GetEnginesByTypeEngineNames(engineNames []string) ([]Engine, error) {
	var engines []Engine

	var conditions []string
	var args []interface{}

	for _, name := range engineNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		// Normalize input
		clean := nonAlnumRegex.ReplaceAllString(name, "")
		clean = strings.ToUpper(clean)

		if clean == "" {
			continue
		}

		conditions = append(conditions,
			"UPPER(REGEXP_REPLACE(type_engine_name, '[^A-Za-z0-9]', '')) LIKE ?",
		)

		args = append(args, "%"+clean+"%")
	}

	if len(conditions) == 0 {
		return engines, nil
	}

	query := database.DB.Where(strings.Join(conditions, " OR "), args...)

	if err := query.Find(&engines).Error; err != nil {
		return nil, err
	}

	return engines, nil
}
