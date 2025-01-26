package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/Br0ce/opera/pkg/monitor"
	"github.com/Br0ce/opera/pkg/tool"
)

type Transporter interface {
	Get(ctx context.Context, addr string, header map[string][]string) ([]byte, error)
}

const (
	name = "com.github.Br0ce.opera.tool.name"
	host = "com.docker.compose.service"
	port = "com.github.Br0ce.opera.tool.port"
	path = "com.github.Br0ce.opera.tool.path"
)

var _ tool.Discovery = (*Discovery)(nil)

type Discovery struct {
	db        tool.DB
	transport Transporter
	tr        trace.Tracer
	log       *slog.Logger
}

func NewDiscovery(db tool.DB, transport Transporter, log *slog.Logger) *Discovery {
	return &Discovery{
		db:        db,
		transport: transport,
		tr:        otel.Tracer("DockerDiscovery"),
		log:       log,
	}
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

	err := di.collectTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("discover tools: %w", err)
	}

	tools, err := di.db.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all tool from tool db: %w", err)
	}

	return tools, nil
}

func (di *Discovery) collectTools(ctx context.Context) error {
	di.log.Debug("collect all tools", "method", "collectTools")

	containers, err := containers(ctx)
	if err != nil {
		return fmt.Errorf("list containers: %w", err)
	}

	errChan := make(chan error, len(containers))
	defer close(errChan)

	for _, container := range containers {
		di.log.Debug("loop container", "method", "FindAll", "container ID", container.ID[:12], "imageTage", container.Image)
		go di.addTool(ctx, container, errChan)
	}

	var addErr error
	for range len(containers) {
		err = <-errChan
		addErr = errors.Join(addErr, err)
	}

	return addErr
}

func (di *Discovery) addTool(ctx context.Context, container types.Container, errChan chan error) {
	di.log.Debug("check if container is a tool and add to db",
		"method", "addTool",
		"containerImage", container.Image)

	tName, okName := container.Labels[name]
	tHost, okHost := container.Labels[host]
	tPort, okPort := container.Labels[port]
	tPath, okPath := container.Labels[path]

	if !(okName && okHost && okPort && okPath) {
		errChan <- nil
		return
	}

	di.log.Debug("found tool container",
		"method", "addTool",
		"containerID", container.ID[:12],
		"imageTage", container.Image)

	addr := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%s", tHost, tPort),
		Path:   tPath,
	}

	cfg, err := di.config(ctx, addr)
	if err != nil {
		errChan <- fmt.Errorf("get description: %w", err)
		return
	}

	tool, err := tool.MakeTool(
		tool.WithName(tName),
		tool.WithAddr(addr),
		tool.WithDescription(cfg.Description),
		tool.WithParameters(cfg.Properties, cfg.Required),
	)
	if err != nil {
		errChan <- fmt.Errorf("make tool: %w", err)
		return
	}

	err = di.db.Add(ctx, tool)
	if err != nil {
		errChan <- fmt.Errorf("add tool: %w", err)
		return
	}

	errChan <- nil
}

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

func containers(ctx context.Context) ([]types.Container, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("new docker client: %w", err)
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: false})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}
	return containers, nil
}
