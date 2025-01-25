package engine

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Br0ce/opera/pkg/action"
	"github.com/Br0ce/opera/pkg/agent"
	"github.com/Br0ce/opera/pkg/percept"
)

type Engine struct {
	Agent   *agent.Agent
	Actor   *action.Actor
	MaxIter int
	Log     *slog.Logger
}

func (eg *Engine) Query(ctx context.Context, query string) (string, error) {
	percepts := []percept.Percept{percept.MakeTextUser(query)}
	for i := range eg.MaxIter {
		eg.Log.Debug("iterate agent", "method", "Act", "iterNum", i, "maxIter", eg.MaxIter)

		action, err := eg.Agent.Action(ctx, percepts)
		if err != nil {
			return "", fmt.Errorf("agent actions: %w", err)
		}

		// If action is of type user, return the content.
		if content, ok := action.User(); ok {
			eg.Log.Debug("found user action", "method", "Act", "content", content)
			return content, nil
		}

		content, ok := action.Tool()
		if !ok {
			// Todo
			return "", fmt.Errorf("action is not of type tool")
		}

		percepts, err = eg.Actor.Act(ctx, content)
		if err != nil {
			return "", fmt.Errorf("actor act: %w", err)
		}
	}

	return "", nil
}
