package repository

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	vehicle "gocars-api/internal/vehicle/repository/postgresql/model"
)

type BodyResponse struct {
	Data    json.RawMessage `json:"data"`
	Iv      *string         `json:"iv"`
	Message string          `json:"message"`
	Success bool            `json:"success"`
}

func DecryptAES256CBC(encDataBase64, ivBase64, key string) (*vehicle.Vehicle, error) {
	encData, err := base64.StdEncoding.DecodeString(encDataBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode data: %v", err)
	}

	iv, err := base64.StdEncoding.DecodeString(ivBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode IV: %v", err)
	}

	keyBytes := []byte(key)
	if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	} else if len(keyBytes) < 32 {
		padding := make([]byte, 32-len(keyBytes))
		keyBytes = append(keyBytes, padding...)
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}

	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("invalid IV length: %d", len(iv))
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(encData))
	mode.CryptBlocks(decrypted, encData)

	paddingLen := int(decrypted[len(decrypted)-1])
	if paddingLen > 0 && paddingLen <= aes.BlockSize {
		decrypted = decrypted[:len(decrypted)-paddingLen]
	}

	var vehicle vehicle.Vehicle
	if err := json.Unmarshal([]byte(decrypted), &vehicle); err != nil {
		return nil, fmt.Errorf("failed to parse vehicle data: %w", err)
	}

	return &vehicle, nil
}

func FetchVehicleData(plate string) (*vehicle.Vehicle, error) {
	if plate == "" {
		return nil, fmt.Errorf("plate number is required")
	}

	visited := make(map[string]bool)

	startURL := os.Getenv("GARAGE_HOST") + "platenew?platenumber=" + url.QueryEscape(plate)

	bodyResponse, err := crawlGarage(startURL, visited)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch garage API: %w", err)
	}

	if bodyResponse == nil {
		return nil, fmt.Errorf("garage API returned nil body")
	}

	if bodyResponse.Data == nil {
		return nil, fmt.Errorf("garage API returned empty data")
	}

	if bodyResponse.Iv == nil || *bodyResponse.Iv == "" {
		var vehicleInfo vehicle.Vehicle
		if err := json.Unmarshal(bodyResponse.Data, &vehicleInfo); err != nil {
			return nil, fmt.Errorf("failed to parse vehicle data: %w", err)
		}
		return &vehicleInfo, nil
	}

	var encrypted string
	if err := json.Unmarshal(bodyResponse.Data, &encrypted); err != nil {
		return nil, fmt.Errorf("failed to parse encrypted data: %w", err)
	}

	vehicleInfo, err := DecryptAES256CBC(encrypted, *bodyResponse.Iv, os.Getenv("GARAGE_KEY"))
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return vehicleInfo, nil
}

func crawlGarage(pageURL string, visited map[string]bool) (*BodyResponse, error) {
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

	fmt.Println(res)

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("Response body: %s\n", string(bodyBytes))

	var bodyResponse BodyResponse
	if err := json.Unmarshal(bodyBytes, &bodyResponse); err != nil {
		log.Println(err)
		return nil, fmt.Errorf("failed to parse JSON: %w; raw: %s", err, string(bodyBytes))
	}

	return &bodyResponse, nil
}
