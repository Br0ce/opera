package tool

import (
	"context"
)

type DB interface {
	Add(ctx context.Context, tool Tool) error
	Get(ctx context.Context, name string) (Tool, error)
	All(ctx context.Context) ([]Tool, error)
}
