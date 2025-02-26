package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"go.opentelemetry.io/otel/trace"

	"github.com/Br0ce/opera/pkg/db"
	"github.com/Br0ce/opera/pkg/monitor"
	"github.com/Br0ce/opera/pkg/tool"
)

const (
	name = "com.github.Br0ce.opera.tool.name"
	host = "com.docker.compose.service"
	port = "com.github.Br0ce.opera.tool.port"
	path = "com.github.Br0ce.opera.tool.path"
)

var _ tool.Discovery = (*Discovery)(nil)

var (
	errNotTool = errors.New("not a tool")
)

type Transporter interface {
	Get(ctx context.Context, addr string, header map[string][]string) ([]byte, error)
}

type Discovery struct {
	db        db.Tool
	client    client.APIClient
	transport Transporter
	mu        sync.RWMutex
	tr        trace.Tracer
	log       *slog.Logger
}

// NewDiscovery returns a pointer to a refreshed Discovery.
func NewDiscovery(ctx context.Context, db db.Tool, transport Transporter, log *slog.Logger) (*Discovery, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("new docker client: %w", err)
	}
	di := &Discovery{
		db:        db,
		client:    cli,
		transport: transport,
		tr:        monitor.Tracer("DockerDiscovery"),
		log:       log,
	}
	err = di.Refresh(ctx)
	if err != nil {
		return nil, fmt.Errorf("refresh: %w", err)
	}
	return di, nil
}

// Get returns the Tool for the given name from the database.
func (di *Discovery) Get(ctx context.Context, name string) (tool.Tool, error) {
	_, span := di.tr.Start(ctx, "get tool")
	defer span.End()
	di.log.Debug("get tool", "method", "Get", "name", name, "traceID", monitor.TraceID(span))

	di.mu.RLock()
	defer di.mu.RUnlock()

	to, err := di.db.Get(name)
	if err != nil {
		return tool.Tool{}, fmt.Errorf("get tool %s: %w", name, err)
	}
	return to, nil
}

// All returns all Tools from the database.
func (di *Discovery) All(ctx context.Context) []tool.Tool {
	_, span := di.tr.Start(ctx, "get all tools")
	defer span.End()
	di.log.Debug("get all tools", "method", "All", "traceID", monitor.TraceID(span))

	di.mu.RLock()
	defer di.mu.RUnlock()

	return slices.Collect(di.db.All())
}

// Refresh clears the database and adds any Tools found on the docker socket to the database.
// To set a timeout use the ctx.
func (di *Discovery) Refresh(ctx context.Context) error {
	ctx, span := di.tr.Start(ctx, "refresh tools")
	defer span.End()
	di.log.Debug("refresh all tools", "method", "refresh", "traceID", monitor.TraceID(span))

	conts, err := di.containers(ctx)
	if err != nil {
		return fmt.Errorf("list containers: %w", err)
	}

	errChan := make(chan error, len(conts))
	defer close(errChan)

	// Hold lock so that the db is not deleted while it is being read.
	di.mu.Lock()

	di.db.Clear()
	for _, cont := range conts {
		di.log.Debug("loop containers", "method", "Refresh", "imageTag", cont.Image)
		go func(ctx context.Context, cont types.Container) {
			tool, err := di.toTool(ctx, cont)
			if err != nil {
				if errors.Is(err, errNotTool) {
					errChan <- nil
					return
				}
				errChan <- fmt.Errorf("create tool from container %s: %w", cont.Image, err)
				return
			}
			err = di.db.Add(tool)
			if err != nil {
				errChan <- fmt.Errorf("add tool: %w", err)
				return
			}

			errChan <- nil
		}(ctx, cont)
	}

	var addErr error
	for range len(conts) {
		addErr = errors.Join(addErr, <-errChan)
	}
	di.mu.Unlock()

	return addErr
}

// toTool returns a Tool for the given container. If the container is not a Tool an ErrNotTool is returned.
func (di *Discovery) toTool(ctx context.Context, container types.Container) (tool.Tool, error) {
	di.log.Debug("get tool for container", "method", "toTool", "containerImage", container.Image)

	tName, okName := container.Labels[name]
	tHost, okHost := container.Labels[host]
	tPort, okPort := container.Labels[port]
	tPath, okPath := container.Labels[path]
	if !(okName && okHost && okPort && okPath) {
		return tool.Tool{}, errNotTool
	}

	di.log.Debug("found tool container", "method", "addTool")
	addr := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%s", tHost, tPort),
		Path:   tPath,
	}

	cfg, err := di.config(ctx, addr)
	if err != nil {
		return tool.Tool{}, fmt.Errorf("get config: %w", err)
	}

	result, err := tool.MakeTool(
		tool.WithName(tName),
		tool.WithAddr(addr),
		tool.WithDescription(cfg.Description),
		tool.WithParameters(cfg.Properties, cfg.Required),
	)
	if err != nil {
		return tool.Tool{}, fmt.Errorf("make tool: %w", err)
	}
	return result, nil
}

// config performs a get request to the config endpoint of the given addr and returns the response as a config.
func (di *Discovery) config(ctx context.Context, addr url.URL) (config, error) {
	ctx, span := di.tr.Start(ctx, "get config")
	defer span.End()
	di.log.Debug("get tool config", "method", "config", "addr", addr.String(), "traceID", monitor.TraceID(span))

	url := addr.JoinPath("config")
	bb, err := di.transport.Get(ctx, url.String(), nil)
	if err != nil {
		return config{}, fmt.Errorf("get request: %w", err)
	}
	var cfg config
	err = json.Unmarshal(bb, &cfg)
	if err != nil {
		return config{}, fmt.Errorf("unmarshal config: %w", err)
	}
	return cfg, nil
}

// containers returns all containers found on the docker socket.
func (di *Discovery) containers(ctx context.Context) ([]types.Container, error) {
	defer di.client.Close()
	containers, err := di.client.ContainerList(ctx, container.ListOptions{All: false})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}
	return containers, nil
}
