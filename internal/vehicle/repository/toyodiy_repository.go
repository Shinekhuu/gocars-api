package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

func FetchVehicleInfo(vin string) (market, year, makeVal, model, frame string, err error) {
	url := fmt.Sprintf("https://www.toyodiy.com/parts/q?vin=%s", vin)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
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

	if err = chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second),
		chromedp.OuterHTML("html", &bodyHTML),
	); err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to get HTML: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(bodyHTML))
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	market = doc.Find("#wM a").Text()
	rawYear := doc.Find("#wYR a").Text()
	makeVal = doc.Find("#wMK a").Text()
	model = doc.Find("#wMD a").Text()
	frame = doc.Find("#wFR a").Text()

	if rawYear != "" && strings.Contains(rawYear, "/") {
		parts := strings.Split(rawYear, "/")
		if len(parts) == 2 {
			month := parts[0]
			yearPart := parts[1]
			if len(month) == 1 {
				month = "0" + month
			}
			year = fmt.Sprintf("%s-%s-01", yearPart, month)
		}
	}

	return market, year, makeVal, model, frame, nil
}
