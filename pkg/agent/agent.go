package agent

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Br0ce/opera/pkg/action"
	"github.com/Br0ce/opera/pkg/message"
	"github.com/Br0ce/opera/pkg/percept"
	"github.com/Br0ce/opera/pkg/tool"
)

type Chatter interface {
	Chat(ctx context.Context, msgs []message.Message, tools []tool.Tool) (message.Message, error)
}

type Agent struct {
	client Chatter
	msgs   []message.Message
	tools  []tool.Tool
	log    *slog.Logger
}

func New(sysPrompt string, tools []tool.Tool, client Chatter, log *slog.Logger) *Agent {
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
		tools: tools,
		log:   log,
	}
}

// Action returns, based on the given perceptions and the history of prior perceptions an
// action which shoud be executed.
func (ag *Agent) Action(ctx context.Context, percepts []percept.Percept) (action.Action, error) {
	ag.msgs = append(ag.msgs, messages(percepts)...)

	answer, err := ag.client.Chat(ctx, ag.msgs, ag.tools)
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
