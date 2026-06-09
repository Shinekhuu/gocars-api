package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

func FetchVehicleInfoPartsouq(vin string) (string, error) {
	url := fmt.Sprintf("https://partsouq.com/en/search/all?q=%s", vin)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	var html string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second),
		chromedp.OuterHTML("html", &html),
	); err != nil {
		return "", err
	}

	return html, nil
}
