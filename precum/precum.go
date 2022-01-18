package precum

import (
	"context"
	"errors"
	"regexp"

	"github.com/shibafu528/teaser/precum/core"
	"github.com/shibafu528/teaser/precum/resolver"
)

var (
	// 処理に全く対応していないURLが渡された時のエラー
	ErrUnsupportedUrl = errors.New("unsupported url")
	// resolverがサイトのコンテンツに対応しておらず、他のresolverに委譲する時のエラー
	ErrUnsupportedContent = errors.New("unsupported content")
)

type Material = core.Material

var registry = []resolverPattern{
	{regexp.MustCompile(".*"), func() core.Resolver { return resolver.NewOGPResolver() }},
}

type resolverPattern struct {
	pattern *regexp.Regexp
	factory func() core.Resolver
}

var cache = map[string]*Material{}

func Resolve(ctx context.Context, url string) (*Material, error) {
	if m, ok := cache[url]; ok {
		return m, nil
	}
	for _, e := range registry {
		if e.pattern.MatchString(url) {
			m, err := e.factory().Resolve(ctx, url)
			if errors.Is(err, ErrUnsupportedContent) {
				continue
			}
			if err == nil {
				cache[url] = m
			}
			return m, err
		}
	}
	return nil, ErrUnsupportedUrl
}
