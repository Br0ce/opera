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
	"github.com/Br0ce/opera/pkg/engine/loop"
	"github.com/Br0ce/opera/pkg/tool"
	"github.com/Br0ce/opera/pkg/tool/discovery/docker"
	"github.com/Br0ce/opera/pkg/transport"
)

type API struct {
	mux http.Handler
	log *slog.Logger
}

func NewHTTP(ctx context.Context, log *slog.Logger) (*API, context.CancelFunc, error) {
	mux := http.NewServeMux()
	transDisc := transport.NewHTTP(time.Second * 5)
	discovery, err := docker.NewDiscovery(ctx, inmem.NewToolDB(), transDisc, log)
	if err != nil {
		return nil, nil, fmt.Errorf("new docker discovery: %w", err)
	}

	transEng := transport.NewHTTP(time.Second * 30)
	actor := action.NewActor(discovery, transEng, log.With("name", "Actor"))
	engine := loop.NewEngine(actor, 10, log.With("name", "Engine"))
	agentHandler := handler.NewAgent(engine, inmem.NewAgentDB(), discovery, log.With("name", "AgentHandler"))

	mux.HandleFunc("POST /v1/agents", agentHandler.Create)
	mux.HandleFunc(fmt.Sprintf("POST /v1/agents/{%s}", handler.AgentID), agentHandler.Query)
	mux.HandleFunc(fmt.Sprintf("DELETE /v1/agents/{%s}", handler.AgentID), agentHandler.Delete)

	api := &API{
		mux: mux,
		log: log,
	}

	go api.refresh(ctx, discovery, 30*time.Second)

	cancel := func() {
		log.Info("cancel api", "method", "NewHTTP")
	}

	return api, cancel, nil
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

func (a *API) refresh(ctx context.Context, discovery tool.Discovery, rate time.Duration) {
	tick := time.Tick(rate)
	for range tick {
		a.log.Debug("refresh tool discovery", "method", "refreshDiscovery")
		err := discovery.Refresh(ctx)
		if err != nil {
			a.log.Error("refresh tool discovery", "method", "refreshDiscovery", "error", err.Error())
		}
	}
}
