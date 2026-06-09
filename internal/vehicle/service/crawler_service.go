package service

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type CrawlerService struct {
	garageHost string
}

func NewCrawlerService(garageHost string) *CrawlerService {
	return &CrawlerService{garageHost: garageHost}
}

func (s *CrawlerService) FetchGarageByPlate(plate string) (string, error) {
	encoded := url.QueryEscape(plate)

	if _, err := crawlPage(s.garageHost + "platenew?platenumber=" + encoded); err != nil {
		return "", err
	}

	return crawlPage(s.garageHost + "plate?platenumber=" + encoded)
}

func crawlPage(pageURL string) (string, error) {
	res, err := http.Get(pageURL)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", pageURL, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch %s: status %d", pageURL, res.StatusCode)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	return string(b), nil
}
