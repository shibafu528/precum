package resolver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/shibafu528/precum/core"
)

const defaultUserAgent = "PrecumBot/1.0"

type ogpResolver struct{}

func NewOGPResolver() core.Resolver {
	return &ogpResolver{}
}

func (r *ogpResolver) Resolve(ctx context.Context, url string) (*core.Material, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("OGPResolver(http.NewRequest): %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OGPResolver(http.Client.Do): %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("OGPResolver: status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("OGPResolver(goquery.NewDocumentFromReader): %w", err)
	}

	m := &core.Material{Url: url}
	if s, ok := findMeta(doc, "meta[property=\"og:title\"]", "meta[property=\"twitter:title\"]"); ok {
		m.Title = s
	}
	if len(m.Title) == 0 {
		m.Title = doc.Find("title").First().Text()
	}

	if s, ok := findMeta(doc, "meta[property=\"og:description\"]", "meta[property=\"twitter:description]", "meta[name=\"description\"]"); ok {
		m.Description = s
	}

	if s, ok := findMeta(doc, "meta[property=\"og:image\"]", "meta[property=\"twitter:image\"]"); ok {
		m.Image = s
	}

	return m, nil
}

func findMeta(doc *goquery.Document, selectors ...string) (string, bool) {
	for _, sel := range selectors {
		if t := doc.Find(sel).First(); t.Length() != 0 {
			a, ok := t.Attr("content")
			if ok {
				return a, true
			}
		}
	}
	return "", false
}
