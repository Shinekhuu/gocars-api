package meili

import (
	"fmt"

	"gocars-api/internal/search/models"

	"github.com/meilisearch/meilisearch-go"
	"go.uber.org/zap"
)

const indexName = "articles"

type MeiliService struct {
	client meilisearch.ServiceManager
	index  meilisearch.IndexManager
}

var Default *MeiliService

func Init(url, apiKey string) {
	client := meilisearch.New(url, meilisearch.WithAPIKey(apiKey))
	idx := client.Index(indexName)

	filterAttrs := &[]interface{}{"category_id", "supplier"}
	if _, err := idx.UpdateFilterableAttributes(filterAttrs); err != nil {
		zap.L().Warn("Meili: failed to update filterable attributes", zap.Error(err))
	}

	searchAttrs := &[]string{"article_no", "article_search_no", "product_name", "product_name_mn", "supplier"}
	if _, err := idx.UpdateSearchableAttributes(searchAttrs); err != nil {
		zap.L().Warn("Meili: failed to update searchable attributes", zap.Error(err))
	}

	Default = &MeiliService{client: client, index: idx}
	zap.L().Info("Meilisearch initialized", zap.String("url", url), zap.String("index", indexName))
}

func (s *MeiliService) IndexDocuments(docs []models.MeiliArticle) error {
	if len(docs) == 0 {
		return nil
	}
	pk := "id"
	task, err := s.index.AddDocuments(docs, &meilisearch.DocumentOptions{PrimaryKey: &pk})
	if err != nil {
		return fmt.Errorf("meili index: %w", err)
	}
	zap.L().Debug("Meili: enqueued index task", zap.Int64("taskUID", task.TaskUID), zap.Int("count", len(docs)))
	return nil
}

func (s *MeiliService) DeleteDocument(id uint) error {
	_, err := s.index.DeleteDocument(fmt.Sprintf("%d", id), nil)
	return err
}

type SearchResult struct {
	Hits  []models.MeiliArticle
	Total int64
}

func (s *MeiliService) Search(query string, categoryID *uint, page, limit int) (*SearchResult, error) {
	req := &meilisearch.SearchRequest{
		Limit:  int64(limit),
		Offset: int64((page - 1) * limit),
	}

	if categoryID != nil {
		req.Filter = fmt.Sprintf("category_id = %d", *categoryID)
	}

	resp, err := s.index.Search(query, req)
	if err != nil {
		return nil, fmt.Errorf("meili search: %w", err)
	}

	var hits []models.MeiliArticle
	for _, h := range resp.Hits {
		var article models.MeiliArticle
		if err := h.DecodeInto(&article); err == nil {
			hits = append(hits, article)
		}
	}

	return &SearchResult{Hits: hits, Total: resp.TotalHits}, nil
}
