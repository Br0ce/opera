package tool

import (
	"iter"
)

type DB interface {
	Add(tool Tool) error
	Get(name string) (Tool, error)
	All() iter.Seq[Tool]
	Clear()
}
