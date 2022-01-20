package precum

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type kbs3Resolver struct{}

func NewKbS3Resolver() Resolver {
	return &kbs3Resolver{}
}

func (k kbs3Resolver) Resolve(ctx context.Context, url string) (*Material, error) {
	client := &http.Client{
		Timeout: defaultTimeout,
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("kbs3Resolver(http.NewRequest): %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kbs3Resolver(http.Client.Do): %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("kbs3Resolver: status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("kbs3Resolver(goquery.NewDocumentFromReader): %w", err)
	}

	info := doc.Find("div.post-info").First()

	m := &Material{
		Url:         url,
		Title:       info.Find("h1").First().Text(),
		Description: strings.TrimSpace(info.Find("p.summary").First().Text()),
	}

	info.Find("ul.tags > li.tag > a").Each(func(i int, s *goquery.Selection) {
		t := s.Text()
		switch t {
		case "R-15", "R-18":
			// skip
		default:
			m.Tags = append(m.Tags, t)
		}
	})

	return m, nil
}
