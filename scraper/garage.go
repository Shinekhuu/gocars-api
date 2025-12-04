package scraper

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gocars-api/models"
	"io"
	"net/http"
	"net/url"
	"os"
)

// BodyResponse represents the JSON returned from the API before decryption
type BodyResponse struct {
	Data    string `json:"data"`
	Iv      string `json:"iv"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

var visited = make(map[string]bool)

// DecryptAES256CBC decrypts base64 AES-256-CBC encrypted data and returns VehicleInfo
func DecryptAES256CBC(encDataBase64, ivBase64, key string) (*models.Vehicle, error) {
	// Decode Base64 data
	encData, err := base64.StdEncoding.DecodeString(encDataBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode data: %v", err)
	}

	// Decode Base64 IV
	iv, err := base64.StdEncoding.DecodeString(ivBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode IV: %v", err)
	}

	// Prepare key: ensure 32 bytes for AES-256
	keyBytes := []byte(key)
	if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	} else if len(keyBytes) < 32 {
		padding := make([]byte, 32-len(keyBytes))
		keyBytes = append(keyBytes, padding...)
	}

	// Create AES cipher
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}

	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("invalid IV length: %d", len(iv))
	}

	// Decrypt
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(encData))
	mode.CryptBlocks(decrypted, encData)

	// Remove PKCS7 padding
	paddingLen := int(decrypted[len(decrypted)-1])
	if paddingLen > 0 && paddingLen <= aes.BlockSize {
		decrypted = decrypted[:len(decrypted)-paddingLen]
	}

	// Parse inner body into VehicleInfo
	var vehicle models.Vehicle
	if err := json.Unmarshal([]byte(decrypted), &vehicle); err != nil {
		return nil, fmt.Errorf("failed to parse vehicle data: %w", err)
	}

	return &vehicle, nil
}

func FetchVehicleData(plate string) (*models.Vehicle, error) {
	if plate == "" {
		return nil, fmt.Errorf("plate number is required")
	}

	// Reset visited map for each call
	visited = make(map[string]bool)

	startURL := os.Getenv("GARAGE_HOST") + "platenew?platenumber=" + url.QueryEscape(plate)
	bodyResponse, err := crawl(startURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch garage API: %w", err)
	}

	if bodyResponse == nil {
		return nil, fmt.Errorf("garage API returned nil body")
	}

	if bodyResponse.Data == "" || bodyResponse.Iv == "" {
		return nil, fmt.Errorf("garage API returned empty data or IV")
	}

	vehicleInfo, err := DecryptAES256CBC(bodyResponse.Data, bodyResponse.Iv, os.Getenv("GARAGE_KEY"))
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return vehicleInfo, nil
}

// crawl fetches the API response and parses JSON into BodyResponse
func crawl(pageURL string) (*BodyResponse, error) {
	if visited[pageURL] {
		return nil, fmt.Errorf("URL already visited: %s", pageURL)
	}
	visited[pageURL] = true

	res, err := http.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %w", pageURL, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 HTTP status: %d", res.StatusCode)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var bodyResponse BodyResponse
	if err := json.Unmarshal(bodyBytes, &bodyResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w; raw: %s", err, string(bodyBytes))
	}

	return &bodyResponse, nil
}
