package inmem

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"go.opentelemetry.io/otel/trace"

	"github.com/Br0ce/opera/pkg/monitor"
	"github.com/Br0ce/opera/pkg/tool"
)

var _ tool.DB = (*Tool)(nil)

type Tool struct {
	tools map[string]tool.Tool
	mu    sync.RWMutex
	tr    trace.Tracer
	log   *slog.Logger
}

func NewDB(tracer trace.Tracer, log *slog.Logger) *Tool {
	return &Tool{
		tools: make(map[string]tool.Tool),
		tr:    tracer,
		log:   log,
	}
}

func (to *Tool) Add(ctx context.Context, tool tool.Tool) error {
	_, span := to.tr.Start(ctx, "add tool")
	defer span.End()
	to.log.Debug("add tool", "method", "Add", "name", tool.Name, "traceID", monitor.TraceID(span))

	to.mu.Lock()
	defer to.mu.Unlock()

	_, ok := to.tools[tool.Name()]
	if ok {
		return fmt.Errorf("tool already exists")
	}

	to.tools[tool.Name()] = tool

	return nil
}

func (to *Tool) Get(ctx context.Context, name string) (tool.Tool, error) {
	_, span := to.tr.Start(ctx, "get tool")
	defer span.End()
	to.log.Debug("get tool", "method", "Get", "name", name, "traceID", monitor.TraceID(span))

	to.mu.RLock()
	defer to.mu.RUnlock()

	if name == "" {
		return tool.Tool{}, fmt.Errorf("name %s: invalid", name)
	}

	t, ok := to.tools[name]
	if !ok {
		return tool.Tool{}, fmt.Errorf("%s: not found", name)
	}

	return t, nil
}

func (to *Tool) All(ctx context.Context) ([]tool.Tool, error) {
	_, span := to.tr.Start(ctx, "get all tools")
	defer span.End()
	to.log.Debug("get all tools", "method", "All", "traceID", monitor.TraceID(span))

	tt := make([]tool.Tool, 0, len(to.tools))
	for _, t := range to.tools {
		tt = append(tt, t)
	}

	return tt, nil
}
