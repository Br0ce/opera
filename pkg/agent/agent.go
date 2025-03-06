package agent

import (
	"context"

	"github.com/Br0ce/opera/pkg/action"
	"github.com/Br0ce/opera/pkg/percept"
)

type Agent interface {
	Action(ctx context.Context, percepts []percept.Percept) (action.Action, error)
}
