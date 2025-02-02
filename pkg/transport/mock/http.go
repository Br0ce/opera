package mock

import (
	"context"
	"io"
	"sync"
)

type Transporter struct {
	PostFn      func(ctx context.Context, addr string, header map[string][]string, body io.Reader) ([]byte, error)
	PostInvoked bool
	GetFn       func(ctx context.Context, addr string, header map[string][]string) ([]byte, error)
	GetInvoked  bool
	mu          sync.Mutex
}

func (tp *Transporter) Post(ctx context.Context, addr string, header map[string][]string, body io.Reader) ([]byte, error) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	tp.PostInvoked = true
	return tp.PostFn(ctx, addr, header, body)
}

func (tp *Transporter) Get(ctx context.Context, addr string, header map[string][]string) ([]byte, error) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	tp.GetInvoked = true
	return tp.GetFn(ctx, addr, header)
}
