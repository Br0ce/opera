package inmem

import (
	"context"
	"sync"

	"github.com/Br0ce/opera/pkg/tool"
	"github.com/Br0ce/opera/pkg/tool/db"
)

var _ tool.DB = (*Tool)(nil)

type Tool struct {
	tools sync.Map
}

func NewDB() *Tool {
	return &Tool{}
}

func (to *Tool) Add(ctx context.Context, tool tool.Tool) error {
	_, ok := to.tools.LoadOrStore(tool.Name(), tool)
	if ok {
		return db.ErrAlreadyExists
	}
	return nil
}

func (to *Tool) Get(ctx context.Context, name string) (tool.Tool, error) {
	if name == "" {
		return tool.Tool{}, db.ErrInvalidName
	}

	v, ok := to.tools.Load(name)
	if !ok {
		return tool.Tool{}, db.ErrNotFound
	}

	t, ok := v.(tool.Tool)
	if !ok {
		return tool.Tool{}, db.ErrInternal
	}
	return t, nil
}

func (to *Tool) All(ctx context.Context) ([]tool.Tool, error) {
	all := true
	var tt []tool.Tool
	to.tools.Range(func(_, value any) bool {
		t, ok := value.(tool.Tool)
		if !ok {
			all = false
			return false
		}

		tt = append(tt, t)
		return true
	})

	if !all {
		return nil, db.ErrInternal
	}
	return tt, nil
}

func (to *Tool) Clear(ctx context.Context) error {
	to.tools.Clear()
	return nil
}
