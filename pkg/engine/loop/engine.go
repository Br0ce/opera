package loop

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/trace"

	"github.com/Br0ce/opera/pkg/action"
	"github.com/Br0ce/opera/pkg/agent"
	"github.com/Br0ce/opera/pkg/monitor"
	"github.com/Br0ce/opera/pkg/percept"
	"github.com/Br0ce/opera/pkg/user"
)

type Engine struct {
	actor   *action.Actor
	maxIter int
	tr      trace.Tracer
	log     *slog.Logger
}

func NewEngine(actor *action.Actor, maxIter int, log *slog.Logger) *Engine {
	return &Engine{
		actor:   actor,
		maxIter: maxIter,
		tr:      monitor.Tracer("Engine"),
		log:     log,
	}
}

func (eg *Engine) Query(ctx context.Context, query user.Query, agent agent.Agent) (string, error) {
	ctx, span := eg.tr.Start(ctx, "Query")
	defer span.End()

	percepts := []percept.Percept{percept.MakeUser(query)}
	for i := range eg.maxIter {
		eg.log.Debug("iterate agent", "method", "Query", "iterNum", i, "maxIter", eg.maxIter)

		next, err := agent.Action(ctx, percepts)
		if err != nil {
			return "", fmt.Errorf("agent actions: %w", err)
		}

		// If action is of type user, return the content.
		if content, ok := next.User(); ok {
			eg.log.Debug("found user action", "method", "Act", "content", content)
			return content, nil
		}

		if reason, ok := next.Reason(); ok {
			eg.log.Info(reason, "method", "Act")
		}

		percepts, err = eg.actor.Act(ctx, next)
		if err != nil {
			return "", fmt.Errorf("actor act: %w", err)
		}
	}

	return "", fmt.Errorf("reached max iterations %v", eg.maxIter)
}
