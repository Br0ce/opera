package mock

import (
	"iter"
	"sync"

	"github.com/Br0ce/opera/pkg/tool"
)

var _ tool.DB = (*ToolDB)(nil)

type ToolDB struct {
	AddFn        func(tool tool.Tool) error
	AddInvoked   bool
	GetFn        func(name string) (tool.Tool, error)
	GetInvoked   bool
	AllFn        func() iter.Seq[tool.Tool]
	AllInvoked   bool
	ClearInvoked bool
	mu           sync.Mutex
}

func (t *ToolDB) Add(tool tool.Tool) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.AddInvoked = true
	return t.AddFn(tool)
}

func (t *ToolDB) Get(name string) (tool.Tool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.GetInvoked = true
	return t.GetFn(name)
}

func (t *ToolDB) All() iter.Seq[tool.Tool] {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.AllInvoked = true
	return t.AllFn()
}

func (t *ToolDB) Clear() {
	t.ClearInvoked = true
}
