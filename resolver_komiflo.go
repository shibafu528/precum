package precum

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"

	"github.com/tidwall/gjson"
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

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("komifloResolver(io.ReadAll): %w", err)
	}
	if !gjson.ValidBytes(bytes) {
		return nil, fmt.Errorf("komifloResolver(gjson.ValidBytes): invalid json")
	}

	j := gjson.ParseBytes(bytes)
	artistName := j.Get("content.attributes.artists.children.0.data.name").String()
	if artistName == "" {
		artistName = "?"
	}

	magazineName := j.Get("content.parents.0.data.title").String()
	if magazineName == "" {
		magazineName = "?"
	}

	m := &Material{
		Url:         url,
		Title:       j.Get("content.data.title").String(),
		Description: fmt.Sprintf("%s - %s", artistName, magazineName),
		Image:       "https://t.komiflo.com/564_mobile_large_3x/" + j.Get("content.named_imgs.cover.filename").String(),
	}

	for _, artist := range j.Get("content.attributes.artists.children.#.data.name").Array() {
		m.Tags = append(m.Tags, artist.String())
	}

	var tags []string
	for _, tag := range j.Get("content.attributes.tags.children.#.data.name").Array() {
		tags = append(tags, tag.String())
	}
	sort.Strings(tags)
	m.Tags = append(m.Tags, tags...)

	return m, nil
}
