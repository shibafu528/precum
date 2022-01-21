package precum

import (
	"context"
	"fmt"
	"io"

	"github.com/PuerkitoBio/goquery"
)

type ogpResolver struct{}

func NewOGPResolver() Resolver {
	return &ogpResolver{}
}

func (r *ogpResolver) Resolve(ctx context.Context, url string) (*Material, error) {
	res, err := fetch(ctx, "GET", url)
	if err != nil {
		return nil, fmt.Errorf("OGPResolver(fetch): %w", err)
	}
	defer func() {
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
	}()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("OGPResolver: status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("OGPResolver(goquery.NewDocumentFromReader): %w", err)
	}

	m := &Material{Url: url}
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
