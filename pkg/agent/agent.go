package agent

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/trace"

	"github.com/Br0ce/opera/pkg/action"
	"github.com/Br0ce/opera/pkg/history"
	"github.com/Br0ce/opera/pkg/monitor"
	"github.com/Br0ce/opera/pkg/percept"
	"github.com/Br0ce/opera/pkg/tool"
)

type Generator interface {
	Generate(ctx context.Context, hist history.History, tools []tool.Tool) (action.Action, error)
}

type Agent struct {
	gen       Generator
	history   history.History
	discovery tool.Discovery
	tr        trace.Tracer
	log       *slog.Logger
}

func New(sysPrompt string, discovery tool.Discovery, generate Generator, log *slog.Logger) *Agent {
	hist := history.History{}
	hist.AddSystem(sysPrompt)
	return &Agent{
		gen:       generate,
		history:   hist,
		discovery: discovery,
		tr:        monitor.Tracer("Agent"),
		log:       log,
	}
}

// Action returns, based on the given perceptions and the history of prior perceptions an
// action which can be executed.
func (ag *Agent) Action(ctx context.Context, percepts []percept.Percept) (action.Action, error) {
	ctx, span := ag.tr.Start(ctx, "Action")
	defer span.End()

	ag.history.AddPercepts(percepts)

	tools := ag.discovery.All(ctx)
	next, err := ag.gen.Generate(ctx, ag.history, tools)
	if err != nil {
		return action.Action{}, fmt.Errorf("chat: %w", err)
	}

	ag.history.AddAction(next)

	return next, nil
}
