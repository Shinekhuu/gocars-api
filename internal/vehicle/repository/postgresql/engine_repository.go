package repository

import (
	"regexp"
	"strings"

	"gocars-api/internal/vehicle/repository/postgresql/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EngineRepository struct {
	db *gorm.DB
}

func NewEngineRepository(db *gorm.DB) *EngineRepository {
	return &EngineRepository{db: db}
}

func (r *EngineRepository) GetByModelID(modelID uint) ([]model.Engine, error) {
	var engines []model.Engine
	err := r.db.Where("model_id = ? AND is_fetched = ?", modelID, true).Find(&engines).Error
	return engines, err
}

func (r *EngineRepository) GetByNameLike(manufacturerID, modelID uint, frame string) ([]model.Engine, error) {
	frame = strings.ToUpper(frame)
	var engines []model.Engine
	err := r.db.
		Where("manufacturer_id = ?", manufacturerID).
		Where("model_id = ?", modelID).
		Where("UPPER(type_engine_name) LIKE ?", "%"+frame+"%").
		Find(&engines).Error
	return engines, err
}

var nonAlnumRegex = regexp.MustCompile(`[^\p{L}\p{N}]`)

func (r *EngineRepository) GetByTypeEngineNames(names []string) ([]model.Engine, error) {
	var conditions []string
	var args []interface{}

	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		clean := strings.ToUpper(nonAlnumRegex.ReplaceAllString(name, ""))
		if clean == "" {
			continue
		}
		conditions = append(conditions,
			"UPPER(REGEXP_REPLACE(type_engine_name, '[^A-Za-z0-9]', '')) LIKE ?",
		)
		args = append(args, "%"+clean+"%")
	}

	var engines []model.Engine
	if len(conditions) == 0 {
		return engines, nil
	}

	err := r.db.Where(strings.Join(conditions, " OR "), args...).Find(&engines).Error
	return engines, err
}

func (r *EngineRepository) UpsertMany(engines []model.Engine) error {
	if len(engines) == 0 {
		return nil
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "vehicle_id"}},
		UpdateAll: true,
	}).Create(&engines).Error
}
