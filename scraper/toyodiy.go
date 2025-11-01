package scraper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// FetchVehicleInfo fetches Toyota/Lexus VIN info from ToyoDIY using headless Chrome
func FetchVehicleInfo(vin string) (market, year, makeVal, model, frame string, err error) {
	url := fmt.Sprintf("https://www.toyodiy.com/parts/q?vin=%s", vin)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true), // run in background
		chromedp.Flag("disable-gpu", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64)"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var bodyHTML string

	// Navigate and wait a bit for JS to render
	if err = chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second),         // give JS time to render table
		chromedp.OuterHTML("html", &bodyHTML), // capture full HTML
	); err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to get HTML: %w", err)
	}

	// Parse HTML with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(bodyHTML))
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract values by div id
	market = doc.Find("#wM a").Text()
	rawYear := doc.Find("#wYR a").Text()
	makeVal = doc.Find("#wMK a").Text()
	model = doc.Find("#wMD a").Text()
	frame = doc.Find("#wFR a").Text()

	// Convert "MM/YYYY" → "YYYY-MM-DD"
	if rawYear != "" && strings.Contains(rawYear, "/") {
		parts := strings.Split(rawYear, "/")
		if len(parts) == 2 {
			month := parts[0]
			yearPart := parts[1]
			// Ensure valid 2-digit month
			if len(month) == 1 {
				month = "0" + month
			}
			year = fmt.Sprintf("%s-%s-01", yearPart, month)
		}
	}

	return market, year, makeVal, model, frame, nil
}
