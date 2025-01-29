package transport

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel/propagation"

	"github.com/Br0ce/opera/pkg/monitor"
	"go.opentelemetry.io/otel"
)

func TestHTTPTransportPost(t *testing.T) {
	t.Parallel()

	log := monitor.NewTestLogger(false)

	type fields struct {
		timeout time.Duration
		log     *slog.Logger
	}

	tests := []struct {
		name         string
		fields       fields
		header       http.Header
		body         io.Reader
		want         []byte
		wantErr      bool
		mockUpstream http.HandlerFunc
	}{
		{
			name: "pass",
			fields: fields{
				timeout: time.Second,
				log:     log,
			},
			header: map[string][]string{
				"content-type": {"application/json"},
			},
			body:    strings.NewReader("payload"),
			want:    []byte("OK"),
			wantErr: false,
			mockUpstream: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Fatal("http method")
				}
				if !strings.Contains(r.Header.Get("content-type"), "application/json") {
					t.Fatal("header not found")
				}
				defer r.Body.Close()

				bb, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("read body, %s", err.Error())
				}

				if string(bb) != "payload" {
					t.Fatalf("body, got %s want %s", string(bb), "payload")
				}

				_, err = fmt.Fprintf(w, "OK")
				if err != nil {
					t.Errorf("could not write to responseWriter")
				}
			},
		},
		{
			name: "fail",
			fields: fields{
				timeout: time.Second,
				log:     log,
			},
			wantErr: true,
			mockUpstream: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "some error", http.StatusInternalServerError)
			},
		},
		{
			name: "timeout",
			fields: fields{
				timeout: time.Microsecond,
				log:     log,
			},
			wantErr: true,
			mockUpstream: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Second)
			},
		},
	}

	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				srv := httptest.NewServer(test.mockUpstream)
				defer srv.Close()

				transport := NewHTTP(test.fields.timeout, test.fields.log)
				got, err := transport.Post(context.TODO(), srv.URL, test.header, test.body)
				if (err != nil) != test.wantErr {
					t.Errorf("error, error = %s, wantErr %v", err.Error(), test.wantErr)
					return
				}

				if test.wantErr {
					return
				}

				if string(got) != string(test.want) {
					t.Errorf("post() got = %v, want %v", string(got), (test.want))
				}
			},
		)
	}
}

func TestHTTPTransportPostTimeout(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(time.Second)
		}))
	defer srv.Close()

	timeout := time.Microsecond

	transport := NewHTTP(timeout, monitor.NewLogger(false))
	_, err := transport.Post(context.TODO(), srv.URL, nil, nil)
	if err == nil {
		t.Errorf("error, error != nil")
	}
}

func TestHTTPTransportPostTraceIDPropagation(t *testing.T) {
	t.Parallel()

	monitor.StartTestTracing(context.TODO(), false, "")
	tr := otel.Tracer("tester")
	ctx, _ := tr.Start(context.TODO(), "propagationTest")

	want := monitor.TraceIDFromCtx(ctx)

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			propagator := otel.GetTextMapPropagator()
			ctx := propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))
			got := monitor.TraceIDFromCtx(ctx)
			if got != want {
				t.Fatalf("TraceIDPropagation: traceID got = %s want = %s", got, want)
			}

			w.WriteHeader(http.StatusOK)
		}))
	defer srv.Close()

	transport := NewHTTP(time.Second, monitor.NewTestLogger(false))
	_, err := transport.Post(ctx, srv.URL, nil, nil)
	if err != nil {
		t.Fatalf("transport: %s", err.Error())
	}
}

func TestHTTPTransportGet(t *testing.T) {
	t.Parallel()

	log := monitor.NewTestLogger(false)

	type fields struct {
		timeout time.Duration
		log     *slog.Logger
	}

	tests := []struct {
		name         string
		fields       fields
		header       http.Header
		want         []byte
		wantErr      bool
		mockUpstream http.HandlerFunc
	}{
		{
			name: "pass",
			fields: fields{
				timeout: time.Second,
				log:     log,
			},
			header: map[string][]string{
				"content-type": {"application/json"},
			},
			want:    []byte("OK"),
			wantErr: false,
			mockUpstream: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Fatal("http method")
				}
				if !strings.Contains(r.Header.Get("content-type"), "application/json") {
					t.Fatal("header not found")
				}

				_, err := fmt.Fprintf(w, "OK")
				if err != nil {
					t.Errorf("could not write to responseWriter")
				}
			},
		},
		{
			name: "fail",
			fields: fields{
				timeout: time.Second,
				log:     log,
			},
			wantErr: true,
			mockUpstream: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "some error", http.StatusInternalServerError)
			},
		},
		{
			name: "timeout",
			fields: fields{
				timeout: time.Microsecond,
				log:     log,
			},
			wantErr: true,
			mockUpstream: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Second)
			},
		},
	}

	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				srv := httptest.NewServer(test.mockUpstream)
				defer srv.Close()

				transport := NewHTTP(test.fields.timeout, test.fields.log)
				got, err := transport.Get(context.TODO(), srv.URL, test.header)
				if (err != nil) != test.wantErr {
					t.Errorf("error, error = %s, wantErr %v", err.Error(), test.wantErr)
					return
				}

				if test.wantErr {
					return
				}

				if string(got) != string(test.want) {
					t.Errorf("post() got = %v, want %v", string(got), (test.want))
				}
			},
		)
	}
}

func TestHTTPTransportGetTimeout(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(time.Second)
		}))
	defer srv.Close()

	timeout := time.Microsecond

	transport := NewHTTP(timeout, monitor.NewLogger(false))
	_, err := transport.Get(context.TODO(), srv.URL, nil)
	if err == nil {
		t.Errorf("error, error != nil")
	}
}

func TestHTTPTransportGetTraceIDPropagation(t *testing.T) {
	t.Parallel()

	monitor.StartTestTracing(context.TODO(), false, "")
	tr := otel.Tracer("tester")
	ctx, _ := tr.Start(context.TODO(), "propagationTest")

	want := monitor.TraceIDFromCtx(ctx)

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			propagator := otel.GetTextMapPropagator()
			ctx := propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))
			got := monitor.TraceIDFromCtx(ctx)
			if got != want {
				t.Fatalf("TraceIDPropagation: traceID got = %s want = %s", got, want)
			}

			w.WriteHeader(http.StatusOK)
		}))
	defer srv.Close()

	transport := NewHTTP(time.Second, monitor.NewTestLogger(false))
	_, err := transport.Get(ctx, srv.URL, nil)
	if err != nil {
		t.Fatalf("transport: %s", err.Error())
	}
}
