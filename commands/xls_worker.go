package commands

import (
	"fmt"
	"gocars-api/models"
	"gocars-api/services"
	"gocars-api/utils"
	"log"
	"strings"

	"github.com/xuri/excelize/v2"
)

func Sync() {
	filePath := "/home/ubuntu/project-go/gocars-api/data/latest-data.xlsx"
	cacheFile := "/home/ubuntu/project-go/gocars-api/cache/ai_cache.json"

	// =========================
	// LOAD AI CACHE
	// =========================
	aiCache := models.LoadAICache(cacheFile)

	// =========================
	// OPEN EXCEL
	// =========================
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Fatalf("❌ Failed to open file: %v", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)

	rows, err := f.GetRows(sheetName)
	if err != nil {
		log.Fatalf("❌ Failed to read rows: %v", err)
	}

	log.Printf("📊 Total rows: %d", len(rows))

	// =========================
	// NOT FOUND FILE
	// =========================
	notFoundFile := excelize.NewFile()
	notFoundSheet := "Sheet1"

	headers := []string{"OEM", "Description", "Model", "Brand", "Categories"}
	notFoundFile.SetSheetRow(notFoundSheet, "A1", &headers)

	notFoundRow := 2

	// =========================
	// LOOP
	// =========================
	for i, row := range rows {
		if i == 0 {
			continue // skip header row
		}

		if len(row) < 7 {
			continue
		}

		oem := strings.TrimSpace(row[1])
		name := strings.TrimSpace(row[2])
		modelStr := strings.ToUpper(strings.TrimSpace(row[3]))
		s3Image := strings.TrimSpace(row[4])
		supplierName := strings.TrimSpace(row[5])
		categories := strings.TrimSpace(row[7])

		if modelStr == "" {
			continue
		}

		log.Printf("🔍 Processing [%d/%d]: OEM: %s | Model: %s | Categories: %s", i+1, len(rows), oem, modelStr, categories)

		// =========================
		// CACHE KEY
		// =========================
		key := oem

		var engineNames []string

		// =========================
		// 1. CACHE HIT & AI FALLBACK
		// =========================
		if cached, ok := aiCache[key]; ok {
			engineNames = cached
			log.Printf("⚡ Cache hit → %v", engineNames)
		} else {
			// =========================
			// AI FALLBACK
			// =========================
			aiResult, err := services.MapWithAI(oem, name, modelStr, supplierName)
			if err != nil {
				log.Printf("❌ AI failed: %v", err)
			} else if len(aiResult) > 0 {
				engineNames = utils.Unique(aiResult)
				aiCache[key] = engineNames

				models.SaveAICache(cacheFile, aiCache) // ✅ persist

				log.Printf("🤖 AI mapping → %v %s", modelStr, aiResult)
			}
		}

		// =========================
		// 2. DB LOOKUP
		// =========================
		engines, err := models.GetEnginesByTypeEngineNames(engineNames)
		if err != nil {
			log.Printf("❌ DB Error: %v", err)
			continue
		}

		// =========================
		// 3. SAVE
		// =========================
		if len(engines) > 0 {
			var vehicleIDs []uint
			for _, e := range engines {
				vehicleIDs = append(vehicleIDs, e.VehicleID)
			}

			article := models.ArticleItem{
				ArticleSearchNo:    oem,
				ArticleNo:          oem,
				ArticleProductName: name,
				S3Image:            s3Image,
				SupplierName:       supplierName,
				Price:              0,
				IsFetched:          false,
			}

			// ✅ NO goroutine (important)
			_, err := models.SaveArticleItemToDB(&article, vehicleIDs, categories)
			if err != nil {
				log.Printf("❌ Save failed: %v", err)
			} else {
				log.Printf("%s: ✅ Saved (%d engines)", oem, len(vehicleIDs))
			}

			continue
		}

		// =========================
		// 4. NOT FOUND
		// =========================
		notFoundFile.SetSheetRow(
			notFoundSheet,
			fmt.Sprintf("A%d", notFoundRow),
			&[]interface{}{oem, name, modelStr, supplierName, categories},
		)

		log.Printf("❌ Not found: %s", modelStr)
		notFoundRow++
	}

	// =========================
	// SAVE NOT FOUND FILE
	// =========================
	if err := notFoundFile.SaveAs("/home/ubuntu/gocars-api/not/not_found.xlsx"); err != nil {
		log.Fatalf("❌ Failed to save not_found file: %v", err)
	}

	log.Println("✅ Sync completed")
}
