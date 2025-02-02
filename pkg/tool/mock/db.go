package mock

import (
	"context"
	"sync"

	"github.com/Br0ce/opera/pkg/tool"
)

var _ tool.DB = (*ToolDB)(nil)

type ToolDB struct {
	AddFn        func(ctx context.Context, tool tool.Tool) error
	AddInvoked   bool
	GetFn        func(ctx context.Context, name string) (tool.Tool, error)
	GetInvoked   bool
	AllFn        func(ctx context.Context) ([]tool.Tool, error)
	AllInvoked   bool
	ClearFn      func(ctx context.Context) error
	ClearInvoked bool
	mu           sync.Mutex
}

func (t *ToolDB) Add(ctx context.Context, tool tool.Tool) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.AddInvoked = true
	return t.AddFn(ctx, tool)
}

func (t *ToolDB) Get(ctx context.Context, name string) (tool.Tool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.GetInvoked = true
	return t.GetFn(ctx, name)
}

func (t *ToolDB) All(ctx context.Context) ([]tool.Tool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.AllInvoked = true
	return t.AllFn(ctx)
}

func (t *ToolDB) Clear(ctx context.Context) error {
	t.ClearInvoked = true
	return t.ClearFn(ctx)
}
