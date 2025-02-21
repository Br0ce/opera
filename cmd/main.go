package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/Br0ce/opera/pkg/api"
	"github.com/Br0ce/opera/pkg/monitor"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := start(ctx); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("start opera: %s\n", err.Error())
			os.Exit(-1)
		}
	}
}

// start inits tracing and logging for the service and starts serving opera.
func start(ctx context.Context) error {
	err := godotenv.Load("config/.env")
	if err != nil {
		fmt.Printf("load ./config/.env file: %s\n", err.Error())
	}

	addr, ok := os.LookupEnv("ADDR")
	if !ok {
		addr = ":8080"
		fmt.Printf("cannot read environment variable ADDR, use %s\n", addr)
	}
	traceAddr, ok := os.LookupEnv("TRACE_ADDR")
	if !ok {
		traceAddr = "tracing:4318"
		fmt.Printf("cannot read environment variable TRACE_ADDR, use %s\n", traceAddr)
	}
	debugFlag, ok := os.LookupEnv("DEBUG_LOGGER")
	if !ok {
		fmt.Printf("cannot read environment variable DEBUG_LOGGER, using prod logger\n")
	}
	var debug bool
	if debugFlag == "true" {
		debug = true
	}

	readTimeout, ok := os.LookupEnv("READ_TIMEOUT")
	if !ok {
		fmt.Printf("cannot read environment variable READ_TIMEOUT, using prod logger\n")
	}
	readTmt, err := time.ParseDuration(readTimeout)
	if err != nil {
		return fmt.Errorf("parse read timeout %s: %s", readTimeout, err.Error())
	}

	writeTimeout, ok := os.LookupEnv("WRITE_TIMEOUT")
	if !ok {
		fmt.Printf("cannot read environment variable DEBUG_LOGGER, using prod logger\n")
	}
	writeTmt, err := time.ParseDuration(writeTimeout)
	if err != nil {
		return fmt.Errorf("parse write timeout %s: %s", writeTimeout, err.Error())
	}

	log := monitor.NewLogger(debug)
	tpShutdown, err := monitor.StartTracing(ctx, traceAddr)
	if err != nil {
		return fmt.Errorf("init tracing: %w", err)
	}
	defer func() {
		log.Info("shutdown open telemetry stack", "method", "start")
		err := tpShutdown(context.Background())
		if err != nil {
			fmt.Printf("shut down open telemetry: %s\n", err)
		}
	}()

	api, apiShutdown, err := api.NewHTTP(ctx, log)
	if err != nil {
		return fmt.Errorf("new http api: %w", err)
	}
	defer apiShutdown()

	srv := &http.Server{
		Addr:         addr,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  readTmt,
		WriteTimeout: writeTmt,
		Handler:      otelhttp.NewHandler(api, "/"),
	}

	srvErr := make(chan error, 1)
	log.Info("start api server",
		"method", "start",
		"addr", addr,
		"debugLogger", debug,
		"readTimeoutSec", readTmt.Seconds(),
		"writeTimeoutSec", writeTmt.Seconds())

	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	select {
	case err = <-srvErr:
		return err
	case <-ctx.Done():
		log.Info("context done, shut down server", "method", "start")
		return srv.Shutdown(context.Background())
	}
}
