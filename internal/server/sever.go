package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloud-mill/webrtc-signalling/internal/config"
	"github.com/cloud-mill/webrtc-signalling/internal/logger"
	"go.uber.org/zap"
)

func StartServer() {
	router := NewRouter(AuthMiddleware)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Config.Port),
		Handler: router,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Logger.Info("starting server", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Logger.Panic("server crashed", zap.Error(err))
		}
	}()

	<-stop
	logger.Logger.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Logger.Error("error during graceful shutdown", zap.Error(err))
	} else {
		logger.Logger.Info("server shutdown")
	}
}
