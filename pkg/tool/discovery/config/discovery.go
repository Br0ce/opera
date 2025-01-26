package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/Br0ce/opera/pkg/monitor"
	"github.com/Br0ce/opera/pkg/tool"
)

var _ tool.Discovery = (*Discovery)(nil)

type Discovery struct {
	db  tool.DB
	tr  trace.Tracer
	log *slog.Logger
}

func NewDiscovery(ctx context.Context, path string, db tool.DB, log *slog.Logger) (*Discovery, error) {
	items, err := readItems(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	for _, item := range items {
		tool, err := item.Decode()
		if err != nil {
			return nil, fmt.Errorf("decode tool %s: %w", item.Name, err)
		}
		err = db.Add(ctx, tool)
		if err != nil {
			return nil, fmt.Errorf("add tool: %w", err)
		}
	}

	return &Discovery{
		db:  db,
		tr:  otel.Tracer("ConfigDiscovery"),
		log: log,
	}, nil
}

func (di *Discovery) Get(ctx context.Context, name string) (tool.Tool, error) {
	ctx, span := di.tr.Start(ctx, "get tool")
	defer span.End()
	di.log.Debug("get tool", "method", "Get", "name", name, "traceID", monitor.TraceID(span))

	return di.db.Get(ctx, name)
}

func (di *Discovery) All(ctx context.Context) ([]tool.Tool, error) {
	ctx, span := di.tr.Start(ctx, "get all tools")
	defer span.End()
	di.log.Debug("get all tools", "method", "All", "traceID", monitor.TraceID(span))

	return di.db.All(ctx)
}

func readItems(filename string) ([]Item, error) {
	bb, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", filename, err)
	}

	var items Items
	err = json.Unmarshal(bb, &items)
	if err != nil {
		return nil, fmt.Errorf("unmarshal items: %w", err)
	}

	return items.Tools, nil
}
