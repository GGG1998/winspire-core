package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/winspire/competition-host-stream/internal/config"
	httpx "github.com/winspire/competition-host-stream/internal/http"
	"github.com/winspire/competition-host-stream/internal/projections"
	ssebroker "github.com/winspire/competition-host-stream/internal/sse"
	"github.com/winspire/competition-host-stream/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	clients, err := store.NewClients(ctx, cfg)
	if err != nil {
		logger.Error("failed to init clients", "error", err)
		os.Exit(1)
	}
	defer clients.Close()

	broker := ssebroker.NewBroker(cfg.SSEPoolSize)
	defer broker.Close()
	eventRouter := ssebroker.NewEventRouter(broker)
	registry := ssebroker.NewRegistry(clients.PG, cfg.LeaseTTL)

	reader := projections.NewReader(clients.PG)
	cupProjector := projections.NewCupProjector(clients.PG)
	tournamentProjector := projections.NewTournamentProjector(clients.PG)
	attendanceProjector := projections.NewAttendanceProjector(clients.PG)
	matchProjector := projections.NewMatchProjector(clients.PG, eventRouter)
	_ = matchProjector // Reserved for event consumers that update match lobby views

	router := httpx.NewRouter(httpx.ServerDeps{
		Config:        cfg,
		Logger:        logger,
		HealthCheck:   clients.Health,
		Reader:        reader,
		CupProjector:  cupProjector,
		TourProjector: tournamentProjector,
		Attendance:    attendanceProjector,
		EventRouter:   eventRouter,
		Broker:        broker,
		Registry:      registry,
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.ServicePort),
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadTimeout,
	}

	go func() {
		logger.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownGrace)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	logger.Info("server stopped")
}
