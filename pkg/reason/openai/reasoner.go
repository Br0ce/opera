package openai

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"go.opentelemetry.io/otel/trace"

	"github.com/Br0ce/opera/pkg/action"
	"github.com/Br0ce/opera/pkg/history"
	"github.com/Br0ce/opera/pkg/monitor"
	"github.com/Br0ce/opera/pkg/tool"
)

type Reasoner struct {
	client *openai.Client
	model  string
	tr     trace.Tracer
	log    *slog.Logger
}

func NewReasoner(token string, modelName string, log *slog.Logger) *Reasoner {
	return &Reasoner{
		client: openai.NewClient(option.WithAPIKey(token)),
		model:  modelName,
		tr:     monitor.Tracer("Generator"),
		log:    log,
	}
}

func (re *Reasoner) Reason(ctx context.Context, hist history.History, tools []tool.Tool) (action.Action, error) {
	ctx, span := re.tr.Start(ctx, "reason about the history")
	defer span.End()
	re.log.Debug("execute chat request to openai",
		"method", "Reason",
		"traceID", monitor.TraceID(span))

	chat, err := re.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F(messages(hist)),
		Model:    openai.F(re.model),
		Tools:    openai.F(toolParams(tools)),
	})
	if err != nil {
		return action.Action{}, fmt.Errorf("openai: %w", err)
	}

	return decode(chat), nil
}

func toolParams(tools []tool.Tool) []openai.ChatCompletionToolParam {
	var params []openai.ChatCompletionToolParam
	for _, tool := range tools {
		params = append(params,
			openai.ChatCompletionToolParam{
				Type: openai.F(openai.ChatCompletionToolTypeFunction),
				Function: openai.F(openai.FunctionDefinitionParam{
					Name:        openai.String(tool.Name()),
					Description: openai.String(tool.Description()),
					Parameters: openai.F(openai.FunctionParameters{
						"type":       "object",
						"properties": tool.Parameters().Properties,
						"required":   tool.Parameters().Required,
					}),
				}),
			},
		)
	}
	return params
}

func messages(hist history.History) []openai.ChatCompletionMessageParamUnion {
	var mm []openai.ChatCompletionMessageParamUnion
	for _, event := range hist.All() {
		switch subject := event.(type) {
		case history.User:
			query := subject.Content
			if query.Text != "" {
				mm = append(mm, openai.UserMessageParts(openai.TextPart(query.Text)))
			}
			if query.Image != "" {
				mm = append(mm, openai.UserMessageParts(openai.ImagePart(query.Image)))
			}
		case history.Assistant:
			mm = append(mm, openai.AssistantMessage(subject.Content))
		case history.ToolCalls:
			toolCalls := make([]openai.ChatCompletionMessageToolCall, 0, len(subject.Content))
			for _, call := range subject.Content {
				toolCall := openai.ChatCompletionMessageToolCall{
					ID: call.ID,
					Function: openai.ChatCompletionMessageToolCallFunction{
						Arguments: call.Arguments,
						Name:      call.Name,
					},
					Type: openai.ChatCompletionMessageToolCallTypeFunction,
				}
				toolCalls = append(toolCalls, toolCall)
			}
			mm = append(mm, openai.ChatCompletionMessage{
				Role:      openai.ChatCompletionMessageRoleAssistant,
				ToolCalls: toolCalls,
			})
		case history.ToolResponse:
			response := subject.Content
			mm = append(mm, openai.ToolMessage(response.ID, response.Content))
		case history.System:
			mm = append(mm, openai.SystemMessage(subject.Content))
		}
	}

	return mm
}

func decode(chat *openai.ChatCompletion) action.Action {
	msg := chat.Choices[0].Message
	if len(msg.ToolCalls) == 0 {
		return action.MakeUser(msg.Content)
	}

	var cc []tool.Call
	for _, call := range msg.ToolCalls {
		c := tool.Call{
			ID:        call.ID,
			Name:      call.Function.Name,
			Arguments: call.Function.Arguments,
		}
		cc = append(cc, c)
	}

	return action.MakeTool(cc)
}
