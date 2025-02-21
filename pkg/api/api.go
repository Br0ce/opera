package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Br0ce/opera/pkg/action"
	"github.com/Br0ce/opera/pkg/api/handler"
	"github.com/Br0ce/opera/pkg/db/inmem"
	"github.com/Br0ce/opera/pkg/engine"
	"github.com/Br0ce/opera/pkg/monitor"
	"github.com/Br0ce/opera/pkg/tool"
	"github.com/Br0ce/opera/pkg/tool/discovery/docker"
	"github.com/Br0ce/opera/pkg/transport"
)

type Api struct {
	mux http.Handler
	log *slog.Logger
}

func NewHTTP(ctx context.Context, log *slog.Logger) (*Api, context.CancelFunc, error) {
	mux := http.NewServeMux()
	transDisc := transport.NewHTTP(time.Second*5, log)
	discovery, err := docker.NewDiscovery(inmem.NewToolDB(), transDisc, log)
	if err != nil {
		return nil, nil, fmt.Errorf("new docker discovery: %w", err)
	}

	err = discovery.Refresh(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("refresh discovery: %w", err)
	}

	transEng := transport.NewHTTP(time.Second*30, log)
	eg := &engine.Engine{
		Actor:   action.NewActor(discovery, transEng, log),
		MaxIter: 10,
		Tr:      monitor.Tracer("Engine"),
		Log:     log.With("name", "Engine"),
	}

	agentHandler := handler.NewAgent(eg, inmem.NewAgentDB(), discovery, log)

	mux.HandleFunc("POST /v1/agents", agentHandler.Create)
	mux.HandleFunc(fmt.Sprintf("POST /v1/agents/{%s}", handler.AgentID), agentHandler.Query)

	api := &Api{
		mux: mux,
		log: log,
	}

	go api.refreshDiscovery(ctx, discovery, 30*time.Second)

	cancel := func() {
		log.Info("cancel api", "method", "NewHTTP")
	}

	return api, cancel, nil
}

func (a *Api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

func (a *Api) refreshDiscovery(ctx context.Context, discovery tool.Discovery, rate time.Duration) {
	tick := time.Tick(rate)
	for range tick {
		a.log.Debug("refresh tool discovery", "method", "refreshDiscovery")
		err := discovery.Refresh(ctx)
		if err != nil {
			a.log.Error("refresh tool discovery", "method", "refreshDiscovery", "error", err.Error())
		}
	}
}
