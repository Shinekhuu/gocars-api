package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	articles "gocars-api/internal/articles/repository/postgresql/model"
	handlrdto "gocars-api/internal/articles/handler/dto"
	"gocars-api/internal/articles/jobs"
	repo "gocars-api/internal/articles/repository/postgresql"

	"gorm.io/gorm"
)

const FetchTTL = 30 * 24 * time.Hour

var refreshLocks sync.Map

type ArticleService struct {
	articleRepo  *repo.ArticleRepository
	fetchLogRepo *repo.APIFetchLogRepository
	db           *gorm.DB
}

func NewArticleService(articleRepo *repo.ArticleRepository, fetchLogRepo *repo.APIFetchLogRepository, db *gorm.DB) *ArticleService {
	return &ArticleService{
		articleRepo:  articleRepo,
		fetchLogRepo: fetchLogRepo,
		db:           db,
	}
}

func (s *ArticleService) GetArticleDetail(id int, articleID int, page int, limit int) (*handlrdto.ArticleResponse, error) {
	if id == 0 && articleID == 0 {
		return nil, errors.New("missing id or article_id")
	}

	offset := (page - 1) * limit

	var (
		article articles.ArticleItem
		err     error
	)

	article, err = s.articleRepo.FindArticle(id, articleID)

	if err == nil && article.ID != 0 {
		if article.IsFetched || article.ArticleID == nil {
			engines, total, _ := s.articleRepo.GetEngines(article.ID, offset, limit)
			return handlrdto.ToDBResponse(article, engines, total, page, limit), nil
		}

		log.Printf("retry fetching article_id=%d", articleID)
	}

	if articleID == 0 {
		return handlrdto.ToDBResponse(article, nil, 0, page, limit), nil
	}

	return GetArticleCompleteDetailFromRapidAPI(articleID, page, limit, offset)
}

func (s *ArticleService) ShouldRefetch(vehicleID uint, categoryID uint) bool {
	fetchLog, err := s.fetchLogRepo.GetAPIFetchLog(vehicleID, categoryID)

	if err != nil {
		return true
	}

	if fetchLog == nil {
		return true
	}

	return time.Since(fetchLog.LastFetchedAt) > FetchTTL
}

func (s *ArticleService) RefreshArticlesAsync(vehicleID uint, categoryID uint) {
	key := fmt.Sprintf("%d:%d", vehicleID, categoryID)

	_, loaded := refreshLocks.LoadOrStore(key, true)

	if loaded {
		return
	}

	go func() {
		defer refreshLocks.Delete(key)
		_, _ = s.GetArticleItemsFromRapidAPI(vehicleID, categoryID)
	}()
}

func GetArticleCompleteDetailFromRapidAPI(articleID int, page int, limit int, offset int) (*handlrdto.ArticleResponse, error) {
	if articleID == 0 {
		return nil, errors.New("invalid article_id")
	}

	url := fmt.Sprintf(
		"https://%s/articles/article-complete-details/type-id/1?articleId=%d&langId=4&countryFilterId=125",
		os.Getenv("X_RAPIDAPI_HOST"), articleID,
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

	var apiResp articles.RapidAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	api := apiResp.Article
	log.Printf("API response article_id=%d", api.ArticleID)

	articleIDCopy := new(uint)
	*articleIDCopy = api.ArticleID

	article := articles.ArticleItem{
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

	select {
	case jobs.ArticleQueue <- article:
		log.Printf("queued article_id=%d", articleID)
	default:
		log.Println("queue full, skip")
	}

	return handlrdto.ToAPIResponse(article, page, limit, offset), nil
}
