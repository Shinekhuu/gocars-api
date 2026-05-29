package jobs

import (
	"log"

	articles "gocars-api/internal/articles/repository/postgresql/model"
	db "gocars-api/internal/database/mysql"
	vehicle "gocars-api/internal/vehicle/repository/postgresql/model"
)

func saveEngines(a articles.ArticleItem) {
	log.Println("saveEngines START article_item_id=", a.ID)

	if a.ID == 0 {
		log.Println("skip engines: article_item_id is 0")
		return
	}

	log.Println("cars count:", len(a.CompatibleCarsResponse))

	if len(a.CompatibleCarsResponse) == 0 {
		log.Println("no compatible cars → skip")
		return
	}

	var engines []vehicle.Engine
	var links []articles.ArticleVehicles

	for _, c := range a.CompatibleCarsResponse {
		if c.VehicleID == 0 {
			log.Println("skip invalid vehicle_id=0")
			continue
		}

		engines = append(engines, vehicle.Engine{
			VehicleID:        c.VehicleID,
			ModelID:          c.ModelID,
			ManufacturerName: c.ManufacturerName,
			ModelName:        c.ModelName,
		})

		links = append(links, articles.ArticleVehicles{
			ArticleItemID: a.ID,
			VehicleID:     c.VehicleID,
		})
	}

	gdb := db.DB

	for i := 0; i < len(engines); i += 200 {
		end := i + 200
		if end > len(engines) {
			end = len(engines)
		}

		batch := engines[i:end]

		query := "INSERT IGNORE INTO engines (vehicle_id, model_id, model_name, manufacturer_name, created_at, updated_at) VALUES "
		args := []interface{}{}

		for j, e := range batch {
			if j > 0 {
				query += ","
			}
			query += "(?, ?, ?, ?, NOW(), NOW())"
			args = append(args, e.VehicleID, e.ModelID, e.ModelName, e.ManufacturerName)
		}

		res := gdb.Exec(query, args...)

		if res.Error != nil {
			log.Printf("engine bulk FAILED batch[%d:%d] err=%v", i, end, res.Error)
		} else {
			log.Printf("engine bulk OK batch[%d:%d] rows=%d", i, end, res.RowsAffected)
		}
	}

	for i := 0; i < len(links); i += 200 {
		end := i + 200
		if end > len(links) {
			end = len(links)
		}

		batch := links[i:end]

		query := "INSERT IGNORE INTO article_vehicles (article_item_id, vehicle_id, created_at, updated_at) VALUES "
		args := []interface{}{}

		for j, l := range batch {
			if j > 0 {
				query += ","
			}
			query += "(?, ?, NOW(), NOW())"
			args = append(args, l.ArticleItemID, l.VehicleID)
		}

		res := gdb.Exec(query, args...)

		if res.Error != nil {
			log.Printf("link bulk FAILED batch[%d:%d] err=%v", i, end, res.Error)
		} else {
			log.Printf("link batch[%d:%d] inserted=%d skipped=%d",
				i, end, res.RowsAffected, len(batch)-int(res.RowsAffected))
		}
	}
}
