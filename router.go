package precum

import (
	"context"
	"errors"
	"regexp"
)

type Rule struct {
	Pattern *regexp.Regexp
	Factory func() Resolver
}

type Router struct {
	Rules []Rule
}

func (r *Router) Resolve(ctx context.Context, url string) (*Material, error) {
	for _, e := range r.Rules {
		if e.Pattern.MatchString(url) {
			m, err := e.Factory().Resolve(ctx, url)
			if errors.Is(err, ErrUnsupportedContent) {
				continue
			}
			return m, err
		}
	}
	return nil, ErrUnsupportedUrl
}

var DefaultRouter = &Router{
	Rules: []Rule{
		{regexp.MustCompile("komiflo\\.com(/#!)?/comics/(\\d+)"), NewKomifloResolver},
		{regexp.MustCompile("ss\\.kb10uy\\.org/posts/\\d+$"), NewKbS3Resolver},
		{regexp.MustCompile(".*"), NewOGPResolver},
	},
}
