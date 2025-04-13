package engine_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"github.com/Br0ce/opera/integration/assert"
	"github.com/Br0ce/opera/pkg/action"
	"github.com/Br0ce/opera/pkg/agent"
	"github.com/Br0ce/opera/pkg/agent/function"
	"github.com/Br0ce/opera/pkg/db/inmem"
	"github.com/Br0ce/opera/pkg/engine/loop"
	"github.com/Br0ce/opera/pkg/monitor"
	"github.com/Br0ce/opera/pkg/reason/openai"
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
		Actor   *action.Actor
		MaxIter int
		Log     *slog.Logger
	}
	type args struct {
		ctx   context.Context
		query user.Query
		Agent agent.Agent
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
	trans := transport.NewHTTP(time.Second * 5)
	discovery, err := docker.NewDiscovery(ctx, inmem.NewToolDB(), trans, log)
	// discovery, err := config.NewDiscovery(ctx, "../../data/discovery/tools.json", db, otel.Tracer("ConfigDiscovery"), log)
	if err != nil {
		t.Fatalf("new discovery: %s", err.Error())
	}

	reasoner := openai.NewReasoner(token, "gpt-4o", log)
	sysPrompt := "You are an intelligent assistant that always explains your thought process before taking action."
	agent := function.NewAgent(sysPrompt, discovery, reasoner, log)
	transporter := transport.NewHTTP(time.Second * 30)
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
				Actor:   actor,
				MaxIter: 5,
				Log:     log,
			},
			args: args{
				ctx:   ctx,
				query: user.Query{Text: "Could you recommend surfing in Sydney at the moment?"},
				Agent: agent,
			},
			want: "The weather in Sydney is 30 degrees and surfing is not recommended.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			eg := loop.NewEngine(test.fields.Actor, test.fields.MaxIter, test.fields.Log)
			got, err := eg.Query(test.args.ctx, test.args.query, test.args.Agent)
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
