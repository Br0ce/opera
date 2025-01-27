package action

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/Br0ce/opera/pkg/monitor"
	"github.com/Br0ce/opera/pkg/percept"
	"github.com/Br0ce/opera/pkg/tool"
)

type Transporter interface {
	Post(ctx context.Context, addr string, header map[string][]string, body io.Reader) ([]byte, error)
}

type Actor struct {
	discovery tool.Discovery
	transport Transporter
	tr        trace.Tracer
	log       *slog.Logger
}

func NewActor(discovery tool.Discovery, transport Transporter, log *slog.Logger) *Actor {
	return &Actor{
		discovery: discovery,
		transport: transport,
		tr:        otel.Tracer("Actor"),
		log:       log,
	}
}

// Act executes all given tool calls concurrently and returns their results as a slice
// of perceptions.
func (ac *Actor) Act(ctx context.Context, calls []Call) ([]percept.Percept, error) {
	ctx, span := ac.tr.Start(ctx, "Act on calls")
	defer span.End()

	callNum := len(calls)
	results := make([]percept.Percept, 0, callNum)
	resultChan := make(chan percept.Percept, callNum)
	errChan := make(chan error, callNum)
	defer func() {
		close(resultChan)
		close(errChan)
	}()

	for i, call := range calls {
		ac.log.Debug("iterate tool calls", "method", "Act", "num", i, "traceID", monitor.TraceID(span))

		go func(ctx context.Context, call Call) {
			percept, err := ac.act(ctx, call)
			if err != nil {
				errChan <- fmt.Errorf("act on tool call %s: %w", call.Name, err)
			}
			resultChan <- percept
		}(ctx, call)
	}

	// Collect all results or errors from the act goroutines.
	var err error
	for range callNum {
		select {
		case result := <-resultChan:
			results = append(results, result)
		case callErr := <-errChan:
			err = errors.Join(err, callErr)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("call tool services: %w", err)
	}

	return results, nil
}

// act execute the call to the tool service and returns the result as a perception.
func (ac *Actor) act(ctx context.Context, call Call) (percept.Percept, error) {
	ctx, span := ac.tr.Start(ctx, "execute call")
	defer span.End()
	ac.log.Debug("execute call to the tool service",
		"method", "act",
		"toolName", call.Name,
		"traceID", monitor.TraceID(span))

	tool, err := ac.discovery.Get(ctx, call.Name)
	if err != nil {
		return percept.Percept{}, fmt.Errorf("get tool: %w", err)
	}

	addr := tool.Addr()
	attr := attribute.String("tool.addr", addr.String())
	span.SetAttributes(attr)

	header := make(map[string][]string)
	header["content-type"] = []string{"application/json"}
	resp, err := ac.transport.Post(ctx, addr.String(), header, strings.NewReader(call.Arguments))
	if err != nil {
		return percept.Percept{}, fmt.Errorf("transport: %w", err)
	}

	return percept.MakeTool(call.ID, string(resp)), nil
}
