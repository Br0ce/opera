package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/Br0ce/opera/pkg/agent/function"
	"github.com/Br0ce/opera/pkg/db"
	"github.com/Br0ce/opera/pkg/engine"
	"github.com/Br0ce/opera/pkg/monitor"
	"github.com/Br0ce/opera/pkg/reason/openai"
	"github.com/Br0ce/opera/pkg/tool"
	"github.com/Br0ce/opera/pkg/user"
)

const AgentID = "agentID"

type Agent struct {
	engine    engine.Engine
	db        db.Agent
	discovery tool.Discovery
	tr        trace.Tracer
	pr        propagation.TextMapPropagator
	log       *slog.Logger
}

func NewAgent(engine engine.Engine, db db.Agent, discovery tool.Discovery, log *slog.Logger) *Agent {
	return &Agent{
		engine:    engine,
		db:        db,
		discovery: discovery,
		tr:        monitor.Tracer("AgentHandler"),
		pr:        monitor.Propagator(),
		log:       log,
	}
}

func (ag *Agent) Create(w http.ResponseWriter, r *http.Request) {
	ctx := ag.pr.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	_, span := ag.tr.Start(ctx, "Create agent")
	defer span.End()
	ag.log.Info("create agent", "method", "Create", "traceID", monitor.TraceID(span))

	token := r.Header.Get("X-Api-Key")
	if token == "" {
		http.Error(w, "X-Api-Key empty", http.StatusBadRequest)
		return
	}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	model := r.FormValue("model")
	if model == "" {
		http.Error(w, "model is empty", http.StatusBadRequest)
		return
	}
	prompt := r.FormValue("system-prompt")

	ag.log.Debug("create agent", "method", "Create",
		"model", model,
		"prompt", prompt,
		"traceID", monitor.TraceID(span))

	reasoner := openai.NewReasoner(token, model, ag.log)
	a := function.NewAgent(prompt, ag.discovery, reasoner, ag.log)
	id, err := ag.db.Add(a)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := map[string]string{
		"object": "created",
		"id":     id,
	}
	bb, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if u, err := url.ParseRequestURI(r.RequestURI); err == nil {
		w.Header().Set("Location", path.Join(u.Path, id))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(bb)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (ag *Agent) Query(w http.ResponseWriter, r *http.Request) {
	ctx := ag.pr.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	ctx, span := ag.tr.Start(ctx, "Query agent")
	defer span.End()

	id := r.PathValue(AgentID)
	ag.log.Info("query agent", "method", "Query", "id", id, "traceID", monitor.TraceID(span))

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	text := r.FormValue("text")
	if text == "" {
		http.Error(w, "text is empty", http.StatusBadRequest)
		return
	}
	ag.log.Debug("query agent", "method", "Query",
		"agentID", id,
		"text", text,
		"traceID", monitor.TraceID(span))

	a, err := ag.db.Get(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("get agent %s: %s", id, err.Error()), http.StatusBadRequest)
		return
	}

	res, err := ag.engine.Query(ctx, user.Query{Text: text}, a)
	if err != nil {
		// TODO status
		http.Error(w, fmt.Sprintf("query: %s", err.Error()), http.StatusBadRequest)
		return
	}

	err = ag.db.Update(id, a)
	if err != nil {
		http.Error(w, fmt.Sprintf("update agent: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	ans := map[string]string{
		"object": "answer",
		"text":   res,
	}
	bb, err := json.Marshal(ans)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if u, err := url.ParseRequestURI(r.RequestURI); err == nil {
		w.Header().Set("Location", path.Join(u.Path, id))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bb)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (ag *Agent) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := ag.pr.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	_, span := ag.tr.Start(ctx, "delete agent")
	defer span.End()

	id := r.PathValue(AgentID)
	ag.log.Info("delete agent", "method", "Delete", "id", id, "traceID", monitor.TraceID(span))

	err := ag.db.Delete(id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			http.Error(w, fmt.Sprintf("delete agent: %s", err.Error()), http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("delete agent: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
