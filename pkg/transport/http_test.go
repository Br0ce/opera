package transport

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/Br0ce/opera/pkg/monitor"
)

func TestHTTPTransportPost(t *testing.T) {
	t.Parallel()

	type fields struct {
		timeout time.Duration
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

				transport := NewHTTP(test.fields.timeout)
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

	transport := NewHTTP(timeout)
	_, err := transport.Post(context.TODO(), srv.URL, nil, nil)
	if err == nil {
		t.Errorf("error, error != nil")
	}
}

func TestHTTPTransportPostTraceIDPropagation(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()
	shutdown, err := monitor.StartTestTracing(context.TODO(), false, "")
	if err != nil {
		t.Fatalf("start tracing: %s", err.Error())
	}
	defer func() {
		err := shutdown(ctx)
		if err != nil {
			fmt.Printf("shut down open telemetry")
		}
	}()

	tr := otel.Tracer("tester")
	ctx, _ = tr.Start(context.TODO(), "propagationTest")

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

	transport := NewHTTP(time.Second)
	_, err = transport.Post(ctx, srv.URL, nil, nil)
	if err != nil {
		t.Fatalf("transport: %s", err.Error())
	}
}

func TestHTTPTransportGet(t *testing.T) {
	t.Parallel()

	type fields struct {
		timeout time.Duration
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

				transport := NewHTTP(test.fields.timeout)
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

	transport := NewHTTP(timeout)
	_, err := transport.Get(context.TODO(), srv.URL, nil)
	if err == nil {
		t.Errorf("error, error != nil")
	}
}

func TestHTTPTransportGetTraceIDPropagation(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()
	shutdown, err := monitor.StartTestTracing(context.TODO(), false, "")
	if err != nil {
		t.Fatalf("start tracing: %s", err.Error())
	}
	defer func() {
		err := shutdown(ctx)
		if err != nil {
			fmt.Printf("shut down open telemetry")
		}
	}()

	tr := otel.Tracer("tester")
	ctx, _ = tr.Start(context.TODO(), "propagationTest")

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

	transport := NewHTTP(time.Second)
	_, err = transport.Get(ctx, srv.URL, nil)
	if err != nil {
		t.Fatalf("transport: %s", err.Error())
	}
}
