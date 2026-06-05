package service

import "strings"

type VinInfo struct {
	OriginalVIN  string `json:"original_vin"`
	FullVIN      string `json:"guessed_full_vin"`
	CarBrand     string `json:"car_brand"`
	ModelName    string `json:"model_name"`
	ModelYear    string `json:"model_year"`
	EngineType   string `json:"engine_type"`
	BodyType     string `json:"body_type"`
	Transmission string `json:"transmission"`
	PlantCode    string `json:"plant_code"`
	SerialNumber string `json:"serial_number"`
	Market       string `json:"market_region"`
}

var wmiMap = map[string]struct {
	Brand  string
	Market string
}{
	"WDB": {"Mercedes-Benz", "Germany"},
	"WDC": {"Mercedes-Benz", "Germany"},
	"WDD": {"Mercedes-Benz", "Germany"},
	"JTE": {"Toyota", "Japan"},
	"JTH": {"Lexus", "Japan"},
	"1HG": {"Honda", "USA"},
	"1N4": {"Nissan", "USA"},
	"MHU": {"Toyota", "Japan"},
	"5N1": {"Nissan", "USA"},
	"5T1": {"Toyota", "USA"},
	"5YJ": {"Tesla", "USA"},
}

var vdsMap = map[string]struct {
	Model        string
	Engine       string
	BodyType     string
	Transmission string
}{
	"MHU38": {"Harrier / RX400h / Highlander Hybrid", "3.3L Hybrid", "SUV", "Automatic"},
	"SALGR": {"Range Rover Sport", "4.4L V8", "SUV", "Automatic"},
	"ES331": {"Lexus ES 330", "3.3L V6", "Sedan", "Automatic"},
}

// yearMap covers model years 2001–2030 (VIN position 10).
var yearMap = map[rune]string{
	'1': "2001", '2': "2002", '3': "2003", '4': "2004", '5': "2005",
	'6': "2006", '7': "2007", '8': "2008", '9': "2009",
	'A': "2010", 'B': "2011", 'C': "2012", 'D': "2013", 'E': "2014",
	'F': "2015", 'G': "2016", 'H': "2017", 'J': "2018", 'K': "2019",
	'L': "2020", 'M': "2021", 'N': "2022", 'P': "2023", 'R': "2024",
	'S': "2025", 'T': "2026", 'V': "2027", 'W': "2028", 'X': "2029",
	'Y': "2030",
}

type VinService struct{}

func NewVinService() *VinService { return &VinService{} }

func (s *VinService) Decode(vin string) VinInfo {
	vin = strings.ToUpper(vin)
	info := VinInfo{OriginalVIN: vin}

	if len(vin) >= 3 {
		wmi := vin[0:3]
		if val, ok := wmiMap[wmi]; ok {
			info.CarBrand = val.Brand
			info.Market = val.Market
		} else {
			info.CarBrand = "Unknown"
			info.Market = "Unknown"
		}
	}

	vds := ""
	if len(vin) >= 8 {
		vds = vin[3:8]
	} else if len(vin) > 3 {
		vds = vin[3:]
	}

	for key, val := range vdsMap {
		if strings.HasPrefix(key, vds) || strings.HasPrefix(vds, key) {
			info.ModelName = val.Model
			info.EngineType = val.Engine
			info.BodyType = val.BodyType
			info.Transmission = val.Transmission
			break
		}
	}

	if info.ModelName == "" {
		info.ModelName = "Unknown"
		info.EngineType = "Unknown"
		info.BodyType = "Unknown"
		info.Transmission = "Unknown"
	}

	info.FullVIN = guessFullVIN(vin)

	if len(info.FullVIN) >= 17 {
		info.SerialNumber = info.FullVIN[11:]
	} else if len(info.FullVIN) > 11 {
		info.SerialNumber = info.FullVIN[11:]
	} else {
		info.SerialNumber = "Unknown"
	}

	if len(info.FullVIN) >= 11 {
		info.PlantCode = string(info.FullVIN[10])
	} else {
		info.PlantCode = "Unknown"
	}

	if len(info.FullVIN) >= 10 {
		if year, ok := yearMap[rune(info.FullVIN[9])]; ok {
			info.ModelYear = year
		} else {
			info.ModelYear = "Unknown"
		}
	} else {
		info.ModelYear = "Unknown"
	}

	return info
}

func guessFullVIN(partial string) string {
	partial = strings.ToUpper(partial)
	if len(partial) >= 17 {
		return partial
	}
	return partial + strings.Repeat("0", 17-len(partial))
}
