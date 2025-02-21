package engine_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel"

	"github.com/Br0ce/opera/integration/assert"
	"github.com/Br0ce/opera/pkg/action"
	"github.com/Br0ce/opera/pkg/agent"
	"github.com/Br0ce/opera/pkg/db/inmem"
	"github.com/Br0ce/opera/pkg/engine"
	"github.com/Br0ce/opera/pkg/generate/openai"
	"github.com/Br0ce/opera/pkg/monitor"
	"github.com/Br0ce/opera/pkg/tool/discovery/docker"
	"github.com/Br0ce/opera/pkg/transport"
	"github.com/Br0ce/opera/pkg/user"
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
		query user.Query
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
	generator := openai.NewGenerator(token, "gpt-4o", otel.Tracer("Generator"), log)
	trans := transport.NewHTTP(time.Second*5, log)
	discovery, err := docker.NewDiscovery(inmem.NewToolDB(), trans, otel.Tracer("DockerDiscovery"), log)
	// discovery, err := config.NewDiscovery(ctx, "../../data/discovery/tools.json", db, otel.Tracer("ConfigDiscovery"), log)
	if err != nil {
		t.Fatalf("new discovery: %s", err.Error())
	}
	agent := agent.New("You are a  friendly assistent!", discovery, generator, otel.Tracer("Agent"), log)
	transporter := transport.NewHTTP(time.Second*30, log)
	actor := action.NewActor(discovery, transporter, otel.Tracer("Actor"), log)

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
				query: user.Query{Text: "Could you recommend surfing in Sydney at the moment?"},
			},
			want: "The weather in Sydney is 30 degrees and surfing is not recommended.",
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
			ok, err := assert.True(got, test.want, token)
			if err != nil {
				t.Fatalf("validate response: %s", err.Error())
			}

			if !ok {
				t.Errorf("Engine.Query() response invalid = got %s want %s", got, test.want)
			}
		})
	}
}
