package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func main() {
	addr, ok := os.LookupEnv("ADDR")
	if !ok {
		a := ":8080"
		log.Printf("no environment variable ADDR, default=%s\n", a)
		addr = a
	}
	log.Printf("ADDR is %v\n", addr)

	st := os.Getenv("SLEEP_TIME")
	sleepTime, err := time.ParseDuration(st)
	if err != nil {
		log.Println("no environment variable SLEEP_TIME, default=10s")
		sleepTime = time.Second * 10
	}
	log.Printf("SLEEP_TIME is %v sec\n", sleepTime)

	rf := os.Getenv("RANDOM_FAIL")
	var fail bool
	if rf == "true" {
		fail = true
	}
	log.Printf("RANDOM_FAIL is %v\n", fail)

	trAddr, ok := os.LookupEnv("TRACING_ADDR")
	if !ok {
		trAddr = "tracing:4318"
	}

	ctx := context.Background()
	shutdown, err := StartTracing(ctx, trAddr)
	if err != nil {
		fmt.Printf("ERROR start tracing: %s", err.Error())
	}
	defer func() {
		err := shutdown(ctx)
		if err != nil {
			fmt.Printf("shut down open telemetry")
		}
	}()
	tracer := otel.Tracer("SharkTool")
	propagator := otel.GetTextMapPropagator()

	m := http.NewServeMux()

	m.HandleFunc("GET /config", func(w http.ResponseWriter, r *http.Request) {
		ctx := propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))
		_, span := tracer.Start(ctx, "get config")
		defer span.End()

		config := struct {
			Name     string         `json:"name"`
			Desc     string         `json:"description"`
			Props    map[string]any `json:"properties"`
			Required []string
		}{
			Name: "get_shark_warning",
			Desc: "Get current shark warning level for the location",
			Props: map[string]any{
				"location": map[string]any{
					"type": "string",
				},
			},
			Required: []string{"location"},
		}

		bb, err := json.Marshal(config)
		if err != nil {
			http.Error(w, "marshal response", http.StatusInternalServerError)
			return
		}

		_, err = w.Write(bb)
		if err != nil {
			log.Println("Error: write response")
		}
	})

	m.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		ctx := propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))
		_, span := tracer.Start(ctx, "get shark warning")
		defer span.End()

		log.Printf("get shark warning request: traceID %s\n", span.SpanContext().TraceID().String())

		if r.Body == nil || r.ContentLength == 0 {
			http.Error(w, "body empty", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		bb, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}

		location := struct {
			Loc string `json:"location"`
		}{}

		err = json.Unmarshal(bb, &location)
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}

		log.Printf("received request for location=%s", location.Loc)

		time.Sleep(sleepTime)

		fmt.Fprintf(w, "many sharks and hight danger for %s", location.Loc)
		log.Println("finished processing")
	})

	log.Printf("start mock shark warning %s ...\n", addr)
	err = http.ListenAndServe(addr, m)
	log.Fatal(err)
}

// StartTracing bootstraps OpenTelemetry by setting a propagator and a trace provider with
// the given trace provider addr.
// Use the returned shutdown func to gracefully shutdown OpenTelemetry in case of no error.
func StartTracing(ctx context.Context, tpAddr string) (func(context.Context) error, error) {
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)

	tp, err := newTraceProvider(ctx, tpAddr)
	if err != nil {
		err = errors.Join(err, tp.Shutdown(ctx))
		return nil, fmt.Errorf("set trace provider: %w", err)
	}
	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}

// StartTestTracing returns a noop trace provider for testing and tpAddr is not considered.
// If integration is set to true, tpAddr is forwarded to StartTracing.
func StartTestTracing(ctx context.Context, integration bool, tpAddr string) (func(context.Context) error, error) {
	if integration {
		return StartTracing(ctx, tpAddr)
	}

	exporter := tracetest.NewNoopExporter()
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}

// newTraceProvider creates a trace provider with the given addr.
func newTraceProvider(ctx context.Context, addr string) (*trace.TracerProvider, error) {
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(addr),
		// TODO: Enable TLS
		otlptracehttp.WithInsecure(),
	)

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("sharkService"),
		)),
	)

	return tp, nil
}
