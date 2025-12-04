package models

// 🟩 Vehicle model
type Vehicle struct {
	CarID     uint    `json:"carId"`
	CarName   string  `json:"carName"`
	ManuID    *uint   `json:"manuId"`
	ManuName  string  `json:"manuName"`
	ModelID   *uint   `json:"modelId"`
	ModelName string  `json:"modelName"`
	MotorCode *string `json:"motorCode"`
	MotorType *string `json:"motorType"`
}
