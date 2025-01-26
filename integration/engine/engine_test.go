package engine_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"go.opentelemetry.io/otel"

	"github.com/Br0ce/opera/pkg/action"
	"github.com/Br0ce/opera/pkg/agent"
	"github.com/Br0ce/opera/pkg/engine"
	"github.com/Br0ce/opera/pkg/genai"
	"github.com/Br0ce/opera/pkg/monitor"
	"github.com/Br0ce/opera/pkg/tool/db/inmem"
	"github.com/Br0ce/opera/pkg/tool/discovery/docker"
	"github.com/Br0ce/opera/pkg/transport"
)

func TestEngine_Query(t *testing.T) {
	err := godotenv.Load("../../config/.env.test")
	if err != nil {
		t.Fatalf("load .env.test file: %s", err.Error())
	}
	token, ok := os.LookupEnv("OPENAI_TOKEN")
	if !ok {
		t.Fatal("OPENAI_TOKEN env not found")
	}
	type fields struct {
		Agent   *agent.Agent
		Actor   *action.Actor
		MaxIter int
		Log     *slog.Logger
	}
	type args struct {
		ctx   context.Context
		query string
	}

	ctx := context.TODO()
	tpShutdown, err := monitor.StartTestTracing(ctx, true, "tracing:4318")
	if err != nil {
		t.Fatalf("start tracing: %s", err.Error())
	}
	defer func() {
		err := tpShutdown(ctx)
		if err != nil {
			fmt.Printf("shut down open telemetry")
		}
	}()

	log := monitor.NewTestLogger(true)
	client := genai.NewClient(token, "gpt-4o", log)
	db := inmem.NewDB(log)
	trans := transport.NewHTTP(time.Second*5, log)
	discovery := docker.NewDiscovery(db, trans, log)
	// discovery, err := config.NewDiscovery(ctx, "../../data/discovery/tools.json", db, log)
	if err != nil {
		t.Fatalf("new discovery: %s", err.Error())
	}
	tools, err := discovery.All(ctx)
	if err != nil {
		t.Fatalf("find all tools: %s", err.Error())
	}
	agent := agent.New("You are a friendly assistent!", tools, client, log)
	transporter := transport.NewHTTP(time.Second*30, log)
	actor := action.NewActor(discovery, transporter, log)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "integraion",
			fields: fields{
				Agent:   agent,
				Actor:   actor,
				MaxIter: 5,
				Log:     log,
			},
			args: args{
				ctx:   ctx,
				query: "Could you recomend surfing in Sydney at the moment?",
			},
			want: "The weather in Sydney is 30 degrees and surfing is not recomented",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			eg := &engine.Engine{
				Agent:   test.fields.Agent,
				Actor:   test.fields.Actor,
				MaxIter: test.fields.MaxIter,
				Tr:      otel.Tracer("Engine"),
				Log:     test.fields.Log,
			}
			got, err := eg.Query(test.args.ctx, test.args.query)
			if (err != nil) != test.wantErr {
				t.Errorf("Engine.Query() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			ok, err := valid(got, test.want, token)
			if err != nil {
				t.Fatalf("validate response: %s", err.Error())
			}

			if !ok {
				t.Errorf("Engine.Query() response invalid = got %s want %s", got, test.want)
			}
		})
	}
}

func valid(test, groundTruth, token string) (bool, error) {
	var msgs []openai.ChatCompletionMessageParamUnion
	msg := fmt.Sprintf("Given this ground truth: %s\n Judge if this statement is valid: %s\n Just respond with true of false.", groundTruth, test)
	msgs = append(msgs, openai.SystemMessage("Your job is to make a fair decision."))
	msgs = append(msgs, openai.UserMessage(msg))

	cl := openai.NewClient(option.WithAPIKey(token))
	res, err := cl.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F(msgs),
		Model:    openai.F("gpt-4o"),
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
