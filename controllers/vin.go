package controllers

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// VinInfo stores structured VIN information
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

// Local WMI map for Japanese and US VINs
var localWMIMap = map[string]struct {
	Brand  string
	Market string
}{
	// Mercedes-Benz
	"WDB": {"Mercedes-Benz", "Germany"},
	"WDC": {"Mercedes-Benz", "Germany"},
	"WDD": {"Mercedes-Benz", "Germany"},
	"W1K": {"Mercedes-Benz", "Germany"},
	"W1N": {"Mercedes-Benz", "Germany"},
	"W1X": {"Mercedes-Benz", "Germany"},
	"W1Y": {"Mercedes-Benz", "Germany"},
	"W1Z": {"Mercedes-Benz", "Germany"},
	"W2W": {"Mercedes-Benz", "USA"},
	"W2X": {"Mercedes-Benz", "USA"},
	"W2Y": {"Mercedes-Benz", "USA"},
	"W2Z": {"Mercedes-Benz", "USA"},
	"W3W": {"Mercedes-Benz", "USA"},
	"W3X": {"Mercedes-Benz", "USA"},
	"W3Y": {"Mercedes-Benz", "USA"},
	"W3Z": {"Mercedes-Benz", "USA"},

	// Existing entries
	"JTE": {"Toyota", "Japan"},
	"JTH": {"Lexus", "Japan"},
	"1HG": {"Honda", "USA"},
	"1N4": {"Nissan", "USA"},
	"MHU": {"Toyota", "Japan"},

	// Additional entries
	"2FA": {"Ford", "Canada"},
	"2FB": {"Ford", "Canada"},
	"2G1": {"Chevrolet", "Canada"},
	"2G2": {"Pontiac", "Canada"},
	"3G1": {"Chevrolet", "Mexico"},
	"3G2": {"Pontiac", "Mexico"},
	"3N1": {"Nissan", "Mexico"},
	"4M2": {"Mercury", "USA"},
	"4S1": {"Isuzu", "USA"},
	"5F1": {"Honda", "USA"},
	"5N1": {"Nissan", "USA"},
	"5T1": {"Toyota", "USA"},
	"5YJ": {"Tesla", "USA"},
	"6G2": {"Pontiac", "Australia"},
	"6H1": {"Holden", "Australia"},
	"6MM": {"Mitsubishi", "Australia"},
	"6T1": {"Toyota", "Australia"},
	"8A1": {"Renault", "Argentina"},
	"8AF": {"Ford", "Argentina"},
	"8AD": {"Peugeot", "Argentina"},
	"8GD": {"Peugeot", "Chile"},
	"9BG": {"Chevrolet", "Brazil"},
	"9BD": {"Fiat", "Brazil"},
	"9BF": {"Ford", "Brazil"},
	"93H": {"Honda", "Brazil"},
	"93Y": {"Renault", "Brazil"},
	"9BS": {"Scania", "Brazil"},
	"93R": {"Toyota", "Brazil"},
	"9BW": {"Volkswagen", "Brazil"},
	"9FB": {"Renault", "Colombia"},
}

// Local VDS map for common models
var localVDSMap = map[string]struct {
	Model        string
	Engine       string
	BodyType     string
	Transmission string
}{
	"MHU38": {"Harrier / RX400h / Highlander Hybrid", "3.3L Hybrid", "SUV", "Automatic"},
	"SALGR": {"Range Rover Sport", "4.4L V8", "SUV", "Automatic"},
	"ES331": {"Lexus ES 330", "3.3L V6", "Sedan", "Automatic"},
}

// Year map for 10th character
var yearMap = map[rune]string{
	'1': "2001", '2': "2002", '3': "2003", '4': "2004", '5': "2005",
	'6': "2006", '7': "2007", '8': "2008", '9': "2009",
	'A': "2010", 'B': "2011", 'C': "2012", 'D': "2013", 'E': "2014",
	'F': "2015", 'G': "2016", 'H': "2017",
}

// GuessFullVIN fills missing characters to make 17-character VIN
func GuessFullVIN(partial string) string {
	partial = strings.ToUpper(partial)
	if len(partial) >= 17 {
		return partial
	}
	missing := 17 - len(partial)
	return partial + strings.Repeat("0", missing)
}

// DecodePartialVIN decodes VIN using local maps
func DecodePartialVIN(vin string) (VinInfo, error) {
	vin = strings.ToUpper(vin)
	info := VinInfo{
		OriginalVIN: vin,
	}

	// --- WMI (first 3 characters) ---
	if len(vin) >= 3 {
		wmi := vin[0:3]
		if val, ok := localWMIMap[wmi]; ok {
			info.CarBrand = val.Brand
			info.Market = val.Market
		} else {
			info.CarBrand = "Unknown"
			info.Market = "Unknown"
		}
	}

	// --- VDS (characters 4–8 if available) ---
	vds := ""
	if len(vin) >= 8 {
		vds = vin[3:8] // positions 4–8
	} else if len(vin) > 3 {
		vds = vin[3:] // take whatever is available
	}

	// Match local VDS map using partial key
	for key, val := range localVDSMap {
		if strings.HasPrefix(key, vds) || strings.HasPrefix(vds, key) {
			info.ModelName = val.Model
			info.EngineType = val.Engine
			info.BodyType = val.BodyType
			info.Transmission = val.Transmission
			break
		}
	}

	// --- Fill missing Unknowns ---
	if info.ModelName == "" {
		info.ModelName = "Unknown"
		info.EngineType = "Unknown"
		info.BodyType = "Unknown"
		info.Transmission = "Unknown"
	}

	// --- Fill full VIN to 17 chars ---
	info.FullVIN = GuessFullVIN(vin)

	// --- Serial number (last 6 chars) ---
	if len(info.FullVIN) >= 17 {
		info.SerialNumber = info.FullVIN[11:]
	} else if len(info.FullVIN) > 11 {
		info.SerialNumber = info.FullVIN[11:]
	} else {
		info.SerialNumber = "Unknown"
	}

	// --- Plant code (11th char if available) ---
	if len(info.FullVIN) >= 11 {
		info.PlantCode = string(info.FullVIN[10])
	} else {
		info.PlantCode = "Unknown"
	}

	// --- Model year (10th char if available) ---
	if len(info.FullVIN) >= 10 {
		yearChar := rune(info.FullVIN[9])
		if year, ok := yearMap[yearChar]; ok {
			info.ModelYear = year
		} else {
			info.ModelYear = "Unknown"
		}
	} else {
		info.ModelYear = "Unknown"
	}

	return info, nil
}

// Decode is the Gin handler
func Decode(c *gin.Context) {
	partialVIN := c.DefaultQuery("vin", "MHU382076138")

	info, err := DecodePartialVIN(partialVIN)
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("Error decoding VIN: %v", err),
		})
		return
	}

	// Return JSON response
	c.JSON(200, info)
}
