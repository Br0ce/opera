package db

import (
	"iter"

	"github.com/Br0ce/opera/pkg/tool"
)

type Tool interface {
	Add(tool tool.Tool) error
	Get(name string) (tool.Tool, error)
	All() iter.Seq[tool.Tool]
	Clear()
}
