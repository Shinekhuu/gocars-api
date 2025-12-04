package scraper

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
)

// FetchVehicleInfo fetches Toyota/Lexus VIN info from ToyoDIY using headless Chrome
// func FetchVehicleInfoPartsouq(c *gin.Context) (market, year, makeVal, model, frame string, err error) {

func FetchVehicleInfoPartsouq(c *gin.Context) {
	vin := "GG2W0004204"
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
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second), // wait for JS
		chromedp.OuterHTML("html", &html),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": err,
		})
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
