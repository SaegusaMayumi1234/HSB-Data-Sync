package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/saegusamayumi1234/hsb-data-sync/internal/core"
	"github.com/saegusamayumi1234/hsb-data-sync/internal/job"
)

func main() {
	ctx := context.Background()

	app, err := core.NewApp(ctx)
	if err != nil {
		log.Fatalf("failed to bootstrap: %v", err)
	}

	app.Logger.Info("application started successfully")

	app.Manager.RegisterCron(job.NewSkyblockItemsUpdater())

	app.Manager.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan

	app.Logger.Info("received signal", "signal", sig)

	// 5. Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	if err := app.ShutdownApp(shutdownCtx); err != nil {
		app.Logger.Error("shutdown error", "error", err)
	}

	app.Logger.Info("application shutdown complete")
}
