package models

import (
	"encoding/json"
	"log"
	"os"
)

type AiPart struct {
	ArticleProductName string      `json:"article_product_name"`
	SupplierName       string      `json:"supplier_name"`
	OEM                string      `json:"oem"`
	Vehicles           []AiVehicle `json:"vehicles"`
}

type AiVehicle struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Engine       string `json:"engine"`
	YearFrom     int    `json:"year_from"`
	YearTo       int    `json:"year_to"`
}

func SaveAICache(path string, cache map[string][]string) {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		log.Printf("❌ Cache marshal error: %v", err)
		return
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		log.Printf("❌ Cache write error: %v", err)
		return
	}

	_ = os.Rename(tmp, path)
}

func LoadAICache(path string) map[string][]string {
	cache := make(map[string][]string)

	data, err := os.ReadFile(path)
	if err != nil {
		return cache
	}

	if err := json.Unmarshal(data, &cache); err != nil {
		log.Printf("⚠️ Cache parse error: %v", err)
		return make(map[string][]string)
	}

	log.Printf("📂 Loaded cache: %d entries", len(cache))
	return cache
}
