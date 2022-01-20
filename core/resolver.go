package core

import "context"

type Resolver interface {
	Resolve(ctx context.Context, url string) (*Material, error)
}
