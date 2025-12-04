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

// 🟩 Xyr model
type Xyr struct {
	gorm.Model
	PlateNumber   string      `gorm:"column:plate_number;type:varchar(20);uniqueIndex" json:"plateNumber"`
	CabinNumber   string      `json:"cabinNumber"`
	CountryName   string      `json:"countryName"`
	MarkName      string      `json:"markName"`
	ModelName     string      `json:"modelName"`
	BuildYear     int         `json:"buildYear"`
	ColorName     string      `json:"colorName"`
	Type          string      `json:"type"`
	OwnerType     string      `json:"ownerType"`
	Intent        *string     `json:"intent"`
	ClassName     string      `json:"className"`
	MotorNumber   *string     `json:"motorNumber"`
	ImportDate    *CustomTime `json:"importDate"`
	FuelType      string      `json:"fuelType"`
	ManCount      int         `json:"manCount"`
	AxleCount     int         `json:"axleCount"`
	Capacity      float64     `json:"capacity"`
	Mass          float64     `json:"mass"`
	Weight        float64     `json:"weight"`
	Length        float64     `json:"length"`
	Width         float64     `json:"width"`
	Height        float64     `json:"height"`
	Transmission  *string     `json:"transmission"`
	WheelPosition string      `json:"wheelPosition"`
	RFID          *string     `json:"rfid"`
}

type XyrVehicle struct {
	gorm.Model
	XyrID     uint
	VehicleID uint

	Xyr    Xyr    `gorm:"foreignKey:XyrID;references:ID;constraint:OnDelete:CASCADE;"`
	Engine Engine `gorm:"foreignKey:VehicleID;references:VehicleID;constraint:OnDelete:CASCADE;"`
}

func (Xyr) TableName() string        { return "xyrs" }
func (XyrVehicle) TableName() string { return "xyr_vehicles" }
