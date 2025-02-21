package inmem

import (
	"iter"
	"sync"

	"github.com/Br0ce/opera/pkg/db"
	"github.com/Br0ce/opera/pkg/tool"
)

var _ db.Tool = (*Tool)(nil)

type Tool struct {
	tools sync.Map
}

func NewToolDB() *Tool {
	return &Tool{}
}

// Add stores the tool that can be retrieved with tool name.
// If a tool with the same name is already stored, a db.ErrAlreadyExists is returned.
func (to *Tool) Add(tool tool.Tool) error {
	_, ok := to.tools.LoadOrStore(tool.Name(), tool)
	if ok {
		return db.ErrAlreadyExists
	}
	return nil
}

// Get returns the tool stored for the given name.
// If no tool is found for the given name, a db.ErrNotFound is returned.
func (to *Tool) Get(name string) (tool.Tool, error) {
	if name == "" {
		return tool.Tool{}, db.ErrInvalidID
	}
	v, ok := to.tools.Load(name)
	if !ok {
		return tool.Tool{}, db.ErrNotFound
	}

	t, ok := v.(tool.Tool)
	if !ok {
		// This should not happen.
		return tool.Tool{}, db.ErrInternal
	}
	return t, nil
}

// All returns a tool.Tool iterator.
func (to *Tool) All() iter.Seq[tool.Tool] {
	return func(yield func(tool.Tool) bool) {
		to.tools.Range(func(_, value any) bool {
			if t, ok := value.(tool.Tool); ok {
				return yield(t)
			}
			return true
		})
	}
}

// Clear deletes all tool.Tool entries.
func (to *Tool) Clear() {
	to.tools.Clear()
}
