package dto

type EngineDTO struct {
	ManufacturerName         string  `json:"manufacturerName"`
	ModelName                string  `json:"modelName"`
	TypeEngineName           string  `json:"typeEngineName"`
	ContructionIntervalStart string  `json:"contructionIntervalStart"`
	ContructionIntervalEnd   string  `json:"contructionIntervalEnd"`
	PowerKw                  *string `json:"powerKw"`
	PowerPs                  *string `json:"powerPs"`
	FuelType                 *string `json:"fuelType"`
	BodyType                 *string `json:"bodyType"`
	NumberOfCylinders        *int    `json:"numberOfCylinders"`
	CapacityLt               *string `json:"capacityLt"`
	CapacityTech             string  `json:"capacityTech"`
	EngineCodes              string  `json:"engineCodes"`
}
