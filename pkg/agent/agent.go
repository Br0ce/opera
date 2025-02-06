package agent

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/trace"

	"github.com/Br0ce/opera/pkg/action"
	"github.com/Br0ce/opera/pkg/message"
	"github.com/Br0ce/opera/pkg/percept"
	"github.com/Br0ce/opera/pkg/tool"
)

type Chatter interface {
	Chat(ctx context.Context, msgs []message.Message, tools []tool.Tool) (message.Message, error)
}

type Agent struct {
	client    Chatter
	msgs      []message.Message
	discovery tool.Discovery
	tr        trace.Tracer
	log       *slog.Logger
}

func New(sysPrompt string, discovery tool.Discovery, client Chatter, tracer trace.Tracer, log *slog.Logger) *Agent {
	return &Agent{
		client: client,
		msgs: []message.Message{{
			Role: message.SystemRole,
			User: []percept.User{
				{
					Type: "text",
					Text: sysPrompt,
				},
			},
		}},
		discovery: discovery,
		tr:        tracer,
		log:       log,
	}
}

// Action returns, based on the given perceptions and the history of prior perceptions an
// action which can be executed.
func (ag *Agent) Action(ctx context.Context, percepts []percept.Percept) (action.Action, error) {
	ctx, span := ag.tr.Start(ctx, "Action")
	defer span.End()

	err := ag.discovery.Refresh(ctx)
	if err != nil {
		return action.Action{}, fmt.Errorf("refresh discovery: %w", err)
	}
	tools, err := ag.discovery.All(ctx)
	if err != nil {
		return action.Action{}, fmt.Errorf("get all available tools: %w", err)
	}

	ag.msgs = append(ag.msgs, messages(percepts)...)
	answer, err := ag.client.Chat(ctx, ag.msgs, tools)
	if err != nil {
		return action.Action{}, fmt.Errorf("chat: %w", err)
	}
	ag.msgs = append(ag.msgs, answer)

	// The given answer requires a user action.
	if answer.ForUser() {
		return action.MakeUser(answer.Assistent), nil
	}

	return action.MakeTool(answer.Calls), nil
}

// messages returns the given perceptions as a slice of message.Messges.
func messages(percepts []percept.Percept) []message.Message {
	msgs := make([]message.Message, 0, len(percepts))
	for _, p := range percepts {
		if u, ok := p.User(); ok {
			msg := message.Message{
				Role: message.UserRole,
				User: u,
			}
			msgs = append(msgs, msg)
			continue
		}

		if t, ok := p.Tool(); ok {
			msg := message.Message{
				Role: message.ToolRole,
				Tool: t,
			}
			msgs = append(msgs, msg)
		}
	}

	return msgs
}
