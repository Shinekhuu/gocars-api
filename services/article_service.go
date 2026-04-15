package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"gocars-api/dto"
	"gocars-api/mappers"
	"gocars-api/models"
	"gocars-api/repositories"
	"gocars-api/workers"
	"io"
	"log"
	"net/http"
	"os"
)

func GetArticleDetail(id int, articleID int, page int, limit int) (*dto.ArticleResponse, error) {

	if id == 0 && articleID == 0 {
		return nil, errors.New("missing id or article_id")
	}

	offset := (page - 1) * limit

	var (
		article models.ArticleItem
		err     error
	)

	// ==========================
	// 1️⃣ Try DB
	// ==========================
	article, err = repositories.FindArticle(id, articleID)

	if err == nil && article.ID != 0 {

		if article.IsFetched || article.ArticleID == nil {
			engines, total, _ := repositories.GetEngines(article.ID, offset, limit)
			return mappers.ToDBResponse(article, engines, total, page, limit), nil
		}

		// 🔥 FIX: allow retry instead of blocking forever
		log.Printf("🔁 retry fetching article_id=%d", articleID)
	}

	// ==========================
	// 2️⃣ Call API (ONLY ONCE)
	// ==========================
	if articleID == 0 {
		return nil, errors.New("invalid article_id")
	}

	apiData, err := getArticleCompleteDetailFromRapidAPI(articleID)
	log.Printf("API call → article_id=%d, err=%v", articleID, err)

	if err != nil || apiData == nil {
		if article.ID != 0 {
			return mappers.ToDBResponse(article, nil, 0, page, limit), nil
		}
		return nil, errors.New("failed to fetch article from API")
	}

	// ==========================
	// 3️⃣ PUSH TO WORKER (SAFE)
	// ==========================
	select {
	case workers.ArticleQueue <- *apiData:
		log.Printf("📥 queued article_id=%d", articleID)
	default:
		log.Println("⚠️ queue full, skip")
	}

	return mappers.ToAPIResponse(*apiData, page, limit, offset), nil
}

// ==========================
// API FETCH
// ==========================
func getArticleCompleteDetailFromRapidAPI(articleID int) (*models.ArticleItem, error) {

	url := fmt.Sprintf(
		"https://tecdoc-catalog.p.rapidapi.com/articles/article-complete-details/type-id/1?articleId=%d&langId=4&countryFilterId=125",
		articleID,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-rapidapi-key", os.Getenv("X_RAPIDAPI_KEY"))
	req.Header.Set("x-rapidapi-host", os.Getenv("X_RAPIDAPI_HOST"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var apiResp models.RapidAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	api := apiResp.Article
	log.Printf("API response article_id=%d", api.ArticleID)

	// 🔥 KEEP YOUR POINTER MODEL
	articleIDCopy := new(uint)
	*articleIDCopy = api.ArticleID

	article := models.ArticleItem{
		ArticleID:            articleIDCopy,
		ArticleNo:            api.ArticleNo,
		ArticleProductName:   api.ArticleProductName,
		SupplierID:           int(api.SupplierID),
		SupplierName:         api.SupplierName,
		ProductID:            api.ProductID,
		ArticleMediaType:     api.ArticleMediaType,
		ArticleMediaFileName: api.ArticleMediaFileName,
		S3Image:              api.S3Image,
		IsFetched:            true,

		AllSpecifications:      api.AllSpecifications,
		OemResponses:           api.OemNo,
		CompatibleCarsResponse: api.CompatibleCars,
	}

	return &article, nil
}
