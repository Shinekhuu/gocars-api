package dto

type APIFetchLogDTO struct {
	VehicleID     int    `json:"vehicleId"`
	CategoryID    int    `json:"categoryId"`
	LastFetchedAt string `json:"lastFetchedAt"`
	IsExpired     bool   `json:"isExpired"`
}
