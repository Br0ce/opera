package transport

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// HTTPTransport is a centralized mean for downstream request.
// It takes care of request timeouts an traceID propagation.
type HTTPTransporter struct {
	client  http.Client
	timeout time.Duration
	log     *slog.Logger
}

func NewHTTP(timeout time.Duration, log *slog.Logger) *HTTPTransporter {
	cl := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	return &HTTPTransporter{
		client:  cl,
		timeout: timeout,
		log:     log,
	}
}

// Post exectutes an http post request.
func (tp *HTTPTransporter) Post(ctx context.Context, addr string, header map[string][]string, body io.Reader) ([]byte, error) {
	tp.log.Debug("post request", "method", "Post", "addr", addr)

	timeout, cancel := context.WithTimeout(ctx, tp.timeout)
	defer cancel()

	request, err := http.NewRequestWithContext(timeout, http.MethodPost, addr, body)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	return tp.do(header, request)
}

// Get exectutes an http post request.
func (tp *HTTPTransporter) Get(ctx context.Context, addr string, header map[string][]string) ([]byte, error) {
	tp.log.Debug("get request", "method", "Get", "addr", addr)

	timeout, cancel := context.WithTimeout(ctx, tp.timeout)
	defer cancel()

	request, err := http.NewRequestWithContext(timeout, http.MethodGet, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	return tp.do(header, request)
}

func (tp *HTTPTransporter) do(header http.Header, request *http.Request) ([]byte, error) {
	tp.log.Debug("execute request", "method", "do")

	request.Header = header

	response, err := tp.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}

	if response.StatusCode != 200 {
		//todo check err type and if not url err add to percept
		return nil, fmt.Errorf("status code: %s", response.Status)
	}

	bb, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	response.Body.Close()

	return bb, nil
}
