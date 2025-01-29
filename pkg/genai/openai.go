package genai

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"go.opentelemetry.io/otel/trace"

	"github.com/Br0ce/opera/pkg/action"
	"github.com/Br0ce/opera/pkg/message"
	"github.com/Br0ce/opera/pkg/monitor"
	"github.com/Br0ce/opera/pkg/percept"
	"github.com/Br0ce/opera/pkg/tool"
)

type Client struct {
	client *openai.Client
	model  string
	tr     trace.Tracer
	log    *slog.Logger
}

func NewClient(token string, modelName string, tracer trace.Tracer, log *slog.Logger) *Client {
	return &Client{
		client: openai.NewClient(option.WithAPIKey(token)),
		model:  modelName,
		tr:     tracer,
		log:    log,
	}
}

func (c *Client) Chat(ctx context.Context, msgs []message.Message, tt []tool.Tool) (message.Message, error) {
	ctx, span := c.tr.Start(ctx, "chat request")
	defer span.End()
	c.log.Debug("execute chat request to openai",
		"method", "Chat",
		"lenMsgs", len(msgs),
		"traceID", monitor.TraceID(span))

	chat, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F(messages(msgs)),
		Model:    openai.F(c.model),
		Tools:    openai.F(tools(tt)),
	})
	if err != nil {
		return message.Message{}, fmt.Errorf("openai: %w", err)
	}

	return decode(chat), nil
}

func tools(tools []tool.Tool) []openai.ChatCompletionToolParam {
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

func messages(msgs []message.Message) []openai.ChatCompletionMessageParamUnion {
	var mm []openai.ChatCompletionMessageParamUnion
	for _, msg := range msgs {
		switch msg.Role {
		case message.UserRole:
			for _, part := range msg.User {
				switch part.Type {
				case percept.TextType:
					mm = append(mm, openai.UserMessageParts(openai.TextPart(part.Text)))
				case percept.ImageType:
					mm = append(mm, openai.UserMessageParts(openai.ImagePart(part.ImageUrl)))
				}
			}
		case message.AssistantRole:
			if msg.ForUser() {
				mm = append(mm, openai.AssistantMessage(msg.Assistent))
				continue
			}
			toolCalls := make([]openai.ChatCompletionMessageToolCall, 0, len(msg.Calls))
			for _, call := range msg.Calls {
				toolCall := openai.ChatCompletionMessageToolCall{
					ID: call.ID,
					Function: openai.ChatCompletionMessageToolCallFunction{
						Arguments: call.Arguments,
						Name:      call.Name,
					},
					Type: "function",
				}
				toolCalls = append(toolCalls, toolCall)
			}
			mm = append(mm, openai.ChatCompletionMessage{
				Role:      message.AssistantRole,
				ToolCalls: toolCalls,
			})
		case message.SystemRole:
			mm = append(mm, openai.SystemMessage(msg.System))
		case message.ToolRole:
			mm = append(mm, openai.ToolMessage(msg.Tool.ID, msg.Tool.Content))

		}
	}

	return mm
}

func decode(chat *openai.ChatCompletion) message.Message {
	msg := chat.Choices[0].Message
	if len(msg.ToolCalls) == 0 {
		return message.Message{
			Role:      message.AssistantRole,
			Assistent: msg.Content,
			Created:   time.Unix(chat.Created, 0).UTC(),
		}
	}

	var cc []action.Call
	for _, call := range msg.ToolCalls {
		c := action.Call{
			ID:        call.ID,
			Name:      call.Function.Name,
			Arguments: call.Function.Arguments,
		}
		cc = append(cc, c)
	}

	return message.Message{
		Role:    message.AssistantRole,
		Calls:   cc,
		Created: time.Unix(chat.Created, 0).UTC(),
	}
}
