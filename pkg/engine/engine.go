package engine

import (
	"context"

	"github.com/Br0ce/opera/pkg/agent"
	"github.com/Br0ce/opera/pkg/user"
)

type Engine interface {
	Query(ctx context.Context, query user.Query, agent agent.Agent) (string, error)
}
