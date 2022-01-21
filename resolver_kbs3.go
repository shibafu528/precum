package precum

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type kbs3Resolver struct{}

func NewKbS3Resolver() Resolver {
	return &kbs3Resolver{}
}

func (k kbs3Resolver) Resolve(ctx context.Context, url string) (*Material, error) {
	res, err := fetch(ctx, "GET", url)
	if err != nil {
		return nil, fmt.Errorf("kbs3Resolver(fetch): %w", err)
	}
	defer func() {
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
	}()
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
