package tool

import (
	"context"
)

type Discovery interface {
	Get(ctx context.Context, name string) (Tool, error)
	All(ctx context.Context) ([]Tool, error)
	Refresh(ctx context.Context) error
}
