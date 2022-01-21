package precum

import (
	"context"
	"errors"
)

var (
	// 処理に全く対応していないURLが渡された時のエラー
	ErrUnsupportedUrl = errors.New("unsupported url")
	// resolverがサイトのコンテンツに対応しておらず、他のresolverに委譲する時のエラー
	ErrUnsupportedContent = errors.New("unsupported content")
)

type Material struct {
	Url         string
	Title       string
	Description string
	Image       string
	Tags        []string
}

type Resolver interface {
	Resolve(ctx context.Context, url string) (*Material, error)
}

var cache = map[string]*Material{}

func Resolve(ctx context.Context, url string) (*Material, error) {
	if m, ok := cache[url]; ok {
		return m, nil
	}
	m, err := DefaultRouter.Resolve(ctx, url)
	if err == nil {
		cache[url] = m
	}
	return m, err
}
