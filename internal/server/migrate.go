package server

import (
	"log"
	"strings"

	articles "gocars-api/internal/articles/repository/postgresql/model"
	profile "gocars-api/internal/profile/repository/postgresql/model"
	roder "gocars-api/internal/roder/repository/postgresql/model"
	vehicle "gocars-api/internal/vehicle/repository/postgresql/model"

	"gorm.io/gorm"
)

// migrate runs AutoMigrate for a named batch.
// SQLSTATE 42P07 ("relation already exists") is treated as a warning —
// the table was created on a previous run; GORM will still sync any
// missing columns/indexes on the next restart.
func migrate(db *gorm.DB, batch string, models ...interface{}) {
	if err := db.AutoMigrate(models...); err != nil {
		if strings.Contains(err.Error(), "42P07") || strings.Contains(err.Error(), "already exists") {
			log.Printf("autoMigrate %s: relation already exists — skipping\n", batch)
			return
		}
		log.Fatalf("autoMigrate %s failed: %v", batch, err)
	}
}

func autoMigrate(db *gorm.DB) {
	// Batch 1: tables with no foreign key dependencies
	migrate(db, "batch 1",
		&profile.Profile{}, // public.profiles — links to Supabase auth.users via UUID
		&articles.Category{},
		&articles.Oem{},
		&articles.Dictionary{},
		&vehicle.Manufacturer{},
		&vehicle.Model{},
		&vehicle.ModelFamily{},
	)

	// Batch 2: Engine must exist before EngineFamily and XyrVehicle
	migrate(db, "batch 2",
		&vehicle.Engine{},
	)

	// Batch 3: tables that reference Engine
	migrate(db, "batch 3",
		&vehicle.EngineFamily{},
		&vehicle.Xyr{},
	)

	// Batch 4: XyrVehicle references both Xyr and Engine
	migrate(db, "batch 4",
		&vehicle.XyrVehicle{},
	)

	// Batch 5: article tables with FKs
	migrate(db, "batch 5",
		&articles.ArticleAllSpecification{},
		&articles.ArticleVehicles{},
		&articles.APIFetchLog{},
	)

	// Batch 6: ArticleItem must exist before ArticleOem / ArticleCategory
	migrate(db, "batch 6",
		&articles.ArticleItem{},
	)

	// Batch 7: join tables that reference ArticleItem
	migrate(db, "batch 7",
		&articles.ArticleOem{},
		&articles.ArticleCategory{},
		&articles.DictionaryTranslation{},
	)

	// Batch 8: order tables
	migrate(db, "batch 8",
		&roder.Order{},
		&roder.OrderItem{},
		&roder.Invoice{},
	)
}
