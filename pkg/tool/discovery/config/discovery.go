package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/Br0ce/opera/pkg/tool"
)

var _ tool.Discovery = (*Discovery)(nil)

type Discovery struct {
	db  tool.DB
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
		log: log,
	}, nil
}

func (di *Discovery) Get(ctx context.Context, name string) (tool.Tool, error) {
	return di.db.Get(ctx, name)
}

func (di *Discovery) All(ctx context.Context) ([]tool.Tool, error) {
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
