package assert

import (
	"context"
	"fmt"
	"strconv"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func True(test, groundTruth, token string) (bool, error) {
	var msgs []openai.ChatCompletionMessageParamUnion
	msg := fmt.Sprintf("Given this ground truth: %s\n Judge if this statement is valid: %s\n Just respond with true of false.", groundTruth, test)
	msgs = append(msgs, openai.SystemMessage("Your job is to make a fair decision."))
	msgs = append(msgs, openai.UserMessage(msg))

	cl := openai.NewClient(option.WithAPIKey(token))
	res, err := cl.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F(msgs),
		Model:    openai.F("gpt-4"),
	})
	if err != nil {
		return false, fmt.Errorf("completion request: %w", err)
	}
	judgment, err := strconv.ParseBool(res.Choices[0].Message.Content)
	if err != nil {
		return false, fmt.Errorf("parse response: %w", err)
	}

	return judgment, nil
}
