package models

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// CustomTime handles timestamps without timezone and stores DATE in DB
type CustomTime struct {
	time.Time
}

const ctLayout = "2006-01-02T15:04:05"

// UnmarshalJSON parses JSON date like "2021-04-21T00:00:00"
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := string(bytes.Trim(b, `"`))
	if s == "" || s == "null" {
		return nil
	}
	t, err := time.Parse(ctLayout, s)
	if err != nil {
		return err
	}
	ct.Time = t
	return nil
}

// Value implements driver.Valuer (store as DATE in DB)
func (ct CustomTime) Value() (driver.Value, error) {
	if ct.IsZero() {
		return nil, nil
	}
	// Store only date (YYYY-MM-DD)
	return ct.Format("2006-01-02"), nil
}

// Scan implements sql.Scanner (read from DB)
func (ct *CustomTime) Scan(value interface{}) error {
	if value == nil {
		ct.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		ct.Time = v
		return nil
	case []byte:
		t, err := time.Parse("2006-01-02", string(v))
		if err != nil {
			return err
		}
		ct.Time = t
		return nil
	case string:
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return err
		}
		ct.Time = t
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into CustomTime", value)
	}
}

// IntFromFloat parses JSON floats into int
type IntFromFloat int

func (i *IntFromFloat) UnmarshalJSON(b []byte) error {
	var f float64
	if err := json.Unmarshal(b, &f); err != nil {
		return err
	}
	*i = IntFromFloat(int(f))
	return nil
}

type Xyr struct {
	gorm.Model
	PlateNumber   string      `gorm:"column:plate_number;type:varchar(20);unique"`
	CabinNumber   string      `gorm:"column:cabin_number;type:varchar(20)"`
	CountryName   string      `gorm:"column:country_name;type:varchar(50)"`
	MarkName      string      `gorm:"column:mark_name;type:varchar(50)"`
	ModelName     string      `gorm:"column:model_name;type:varchar(50)"`
	BuildYear     int         `gorm:"column:build_year"`
	ColorName     string      `gorm:"column:color_name;type:varchar(50)"`
	Type          string      `gorm:"column:type;type:varchar(50)"`
	OwnerType     string      `gorm:"column:owner_type;type:varchar(50)"`
	Intent        *string     `gorm:"column:intent;type:varchar(255);default:NULL"`
	ClassName     string      `gorm:"column:class_name;type:varchar(10)"`
	MotorNumber   *string     `gorm:"column:motor_number;type:varchar(50);default:NULL"`
	ImportDate    *CustomTime `json:"importDate" gorm:"type:date"` // DATE type
	FuelType      string      `gorm:"column:fuel_type;type:varchar(50)"`
	ManCount      int         `gorm:"column:man_count"`
	AxleCount     int         `gorm:"column:axle_count"`
	Capacity      float64     `json:"capacity"`
	Mass          float64     `json:"mass"`
	Weight        float64     `json:"weight"`
	Length        float64     `json:"length"`
	Width         float64     `json:"width"`
	Height        float64     `json:"height"`
	Transmission  *string     `gorm:"column:transmission;type:varchar(50);default:NULL"`
	WheelPosition string      `gorm:"column:wheel_position;type:varchar(50)"`
	RFID          *string     `gorm:"column:rfid;type:varchar(50);default:NULL"`
}

// Optional: TableName overrides default table name
func (Xyr) TableName() string {
	return "xyr"
}
