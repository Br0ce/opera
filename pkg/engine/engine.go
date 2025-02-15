package engine

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/trace"

	"github.com/Br0ce/opera/pkg/action"
	"github.com/Br0ce/opera/pkg/agent"
	"github.com/Br0ce/opera/pkg/percept"
	"github.com/Br0ce/opera/pkg/user"
)

type Engine struct {
	Agent   *agent.Agent
	Actor   *action.Actor
	MaxIter int
	Tr      trace.Tracer
	Log     *slog.Logger
}

func (eg *Engine) Query(ctx context.Context, query user.Query) (string, error) {
	ctx, span := eg.Tr.Start(ctx, "Query")
	defer span.End()

	percepts := []percept.Percept{percept.MakeUser(query)}
	for i := range eg.MaxIter {
		eg.Log.Debug("iterate agent", "method", "Act", "iterNum", i, "maxIter", eg.MaxIter)

		next, err := eg.Agent.Action(ctx, percepts)
		if err != nil {
			return "", fmt.Errorf("agent actions: %w", err)
		}

		// If action is of type user, return the content.
		if content, ok := next.User(); ok {
			eg.Log.Debug("found user action", "method", "Act", "content", content)
			return content, nil
		}

		percepts, err = eg.Actor.Act(ctx, next)
		if err != nil {
			return "", fmt.Errorf("actor act: %w", err)
		}
	}

	return "", fmt.Errorf("reached max iterations %v", eg.MaxIter)
}
