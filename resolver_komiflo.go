package precum

import (
	"context"
	"fmt"
	"io"
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

	res, err := fetch(ctx, "GET", "https://api.komiflo.com/content/id/"+id)
	if err != nil {
		return nil, fmt.Errorf("komifloResolver(fetch): %w", err)
	}
	defer func() {
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
	}()
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
