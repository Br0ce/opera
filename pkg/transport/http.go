package transport

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type HTTPTransporter struct {
	timeout time.Duration
	log     *slog.Logger
}

func NewHTTP(timeout time.Duration, log *slog.Logger) *HTTPTransporter {
	return &HTTPTransporter{timeout: timeout, log: log}
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

	cl := http.Client{}
	request.Header = header
	response, err := cl.Do(request)
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
