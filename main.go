package main

import (
	"log"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	"github.com/geziyor/geziyor/export"
	"github.com/playwright-community/playwright-go"
)

func main() {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
		Args:     []string{"--remote-debugging-port=9222"},
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	defer browser.Close()
	defer pw.Stop()

	time.Sleep(time.Second * 2)

	getProduct("B0D85BHLB5")
}

func getProduct(asin string) {
	geziyor.NewGeziyor(&geziyor.Options{
		StartRequestsFunc: func(g *geziyor.Geziyor) {
			g.GetRendered("https://www.amazon.com/product-reviews/"+asin+"?pageNumber=1", g.Opt.ParseFunc)
		},
		ParseFunc:       parseReviews,
		Exporters:       []export.Exporter{&export.JSON{FileName: "reviews.json"}},
		BrowserEndpoint: "ws://localhost:9222",
		RequestDelay:    time.Second * 2,
	}).Start()
}

func parseReviews(g *geziyor.Geziyor, r *client.Response) {
	r.HTMLDoc.Find("div[data-hook='review']").Each(func(_ int, s *goquery.Selection) {
		g.Exports <- map[string]interface{}{
			"title":    s.Find("a[data-hook='review-title']").Text(),
			"rating":   s.Find("i[data-hook='review-star-rating']").Text(),
			"review":   s.Find("span[data-hook='review-body'] span").Text(),
			"date":     s.Find("span[data-hook='review-date']").Text(),
			"author":   s.Find("span.a-profile-name").Text(),
			"verified": s.Find("span[data-hook='avp-badge']").Length() > 0,
		}
	})

	if nextPage, exists := r.HTMLDoc.Find("li.a-last a").Attr("href"); exists {
		g.Get(r.JoinURL(nextPage), parseReviews)
	}
}
