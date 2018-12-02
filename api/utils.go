package api

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	zap "go.uber.org/zap"
	"nimona.io/go/log"
)

func Wait() {
	var gracefulStop = make(chan os.Signal)

	// Get Notified for incoming signals
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	signal.Notify(gracefulStop, os.Interrupt)

	// Wait for signal
	sig := <-gracefulStop

	ctx := context.Background()

	log.Logger(ctx).Info("Terminating...", zap.Any("signal", sig))
}
