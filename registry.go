package precum

import (
	"context"
	"regexp"
)

type Resolver interface {
	Resolve(ctx context.Context, url string) (*Material, error)
}

type rule struct {
	pattern *regexp.Regexp
	factory func() Resolver
}

var registry = []rule{
	{regexp.MustCompile(".*"), NewOGPResolver},
}
