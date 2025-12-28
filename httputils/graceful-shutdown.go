package httputils

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	defaultTerminationGracePeriod = 30 * time.Second
)

type ShutdownConfig struct {
	Server          *http.Server
	ShutdownTimeout time.Duration
	Signals         []os.Signal
	Logger          *slog.Logger
}

func RunWithGracefulShutdown(cfg ShutdownConfig) error {
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = defaultTerminationGracePeriod
	}

	if len(cfg.Signals) == 0 {
		cfg.Signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}

	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	listenErr := make(chan error, 1)

	go func() {
		cfg.Logger.Info("starting server", slog.String("addr", cfg.Server.Addr))
		if err := cfg.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			listenErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, cfg.Signals...)

	select {
	case err := <-listenErr:
		return err
	case sig := <-quit:
		cfg.Logger.Info("received shutdown signal", slog.Any("signal", sig))
	}

	cfg.Logger.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	return cfg.Server.Shutdown(ctx)
}
