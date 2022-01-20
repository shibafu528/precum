package precum

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/bitly/go-simplejson"
)

var komifloComicsPagePattern = regexp.MustCompile("komiflo\\.com(?:/#!)?/comics/(\\d+)")

type komifloResolver struct{}

func NewKomifloResolver() Resolver {
	return &komifloResolver{}
}

func (k komifloResolver) Resolve(ctx context.Context, url string) (*Material, error) {
	matches := komifloComicsPagePattern.FindStringSubmatch(url)
	if matches == nil {
		return nil, fmt.Errorf("komifloResolver: unmatched URL pattern: %s", url)
	}
	id := matches[1]

	client := &http.Client{
		Timeout: defaultTimeout,
	}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.komiflo.com/content/id/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("komifloResolver(http.NewRequest): %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("komifloResolver(http.Client.Do): %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("komifloResolver: status code error: %d %s", res.StatusCode, res.Status)
	}

	j, err := simplejson.NewFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("komifloResolver(simplejson.NewFromReader): %w", err)
	}

	m := &Material{
		Url:         url,
		Title:       j.GetPath("content", "data", "title").MustString(""),
		Description: fmt.Sprintf("%s - %s", j.GetPath("content", "attributes", "artists", "children").GetIndex(0).GetPath("data", "name").MustString("?"), j.GetPath("content", "parents").GetIndex(0).GetPath("data", "title").MustString("?")),
		Image:       "https://t.komiflo.com/564_mobile_large_3x/" + j.GetPath("content", "named_imgs", "cover", "filename").MustString(""),
	}

	return m, nil
}
