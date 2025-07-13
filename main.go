package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shanduur/docker-status/pkg/lister"
	"github.com/shanduur/docker-status/pkg/store"
	"github.com/shanduur/docker-status/pkg/web"
	"github.com/shanduur/docker-status/pkg/web/middleware"
)

var (
	addr string
)

func main() {
	flag.StringVar(&addr, "addr", "0.0.0.0:8080", "")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	mux := http.NewServeMux()

	st := &store.Store{}

	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		log.Fatal(err)
	}

	l := lister.New(st, cli)
	go l.Run(ctx)

	mux.Handle("/",
		middleware.Add(
			web.UiHandler,
			middleware.Logging,
			middleware.RateLimit,
		),
	)
	mux.Handle("/ws/stats",
		middleware.Add(
			web.NewStatsHandler(st),
			middleware.Logging,
			middleware.RateLimit,
		),
	)
	mux.Handle("/metrics",
		middleware.Add(
			promhttp.Handler(),
			middleware.Logging,
			middleware.RateLimit,
		),
	)

	srv := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	log.Printf("Server started at %q", addr)

	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
}
