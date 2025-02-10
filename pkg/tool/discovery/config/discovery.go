package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"

	"github.com/Br0ce/opera/pkg/monitor"
	"github.com/Br0ce/opera/pkg/tool"
)

var _ tool.Discovery = (*Discovery)(nil)

type Discovery struct {
	db   tool.DB
	path string
	tr   trace.Tracer
	log  *slog.Logger
}

func NewDiscovery(ctx context.Context, path string, db tool.DB, tracer trace.Tracer, log *slog.Logger) (*Discovery, error) {
	di := &Discovery{
		db:   db,
		path: path,
		tr:   tracer,
		log:  log,
	}
	err := di.Refresh(ctx)
	if err != nil {
		return nil, fmt.Errorf("refresh db: %w", err)
	}

	return di, nil
}

func (di *Discovery) Get(ctx context.Context, name string) (tool.Tool, error) {
	_, span := di.tr.Start(ctx, "get tool")
	defer span.End()
	di.log.Debug("get tool", "method", "Get", "name", name, "traceID", monitor.TraceID(span))
	return di.db.Get(name)
}

func (di *Discovery) All(ctx context.Context) []tool.Tool {
	_, span := di.tr.Start(ctx, "get all tools")
	defer span.End()
	di.log.Debug("get all tools", "method", "All", "traceID", monitor.TraceID(span))

	var tt []tool.Tool
	for t := range di.db.All() {
		tt = append(tt, t)
	}
	return tt
}

func (di *Discovery) Refresh(ctx context.Context) error {
	_, span := di.tr.Start(ctx, "refresh all tools")
	defer span.End()
	di.log.Debug("refresh all tools", "method", "Refresh", "traceID", monitor.TraceID(span))

	items, err := readItems(di.path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	di.db.Clear()
	for _, item := range items {
		tool, err := item.Decode()
		if err != nil {
			return fmt.Errorf("decode tool %s: %w", item.Name, err)
		}
		err = di.db.Add(tool)
		if err != nil {
			return fmt.Errorf("add tool: %w", err)
		}
	}
	return nil
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
