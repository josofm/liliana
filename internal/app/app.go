package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/josofm/liliana/config"
	v1 "github.com/josofm/liliana/internal/controller/http/v1"
	deckRepo "github.com/josofm/liliana/internal/repository/deck"
	userRepo "github.com/josofm/liliana/internal/repository/user"
	"github.com/josofm/liliana/pkg/httpserver"
	"github.com/josofm/liliana/pkg/logger"
)

func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	handler := gin.New()
	userRepo := userRepo.NewInMemoryRepo()
	deckRepo := deckRepo.NewInMemoryRepo()
	v1.NewRouter(handler, l, userRepo, deckRepo)
	httpServer := httpserver.New(handler)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	var err error
	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}
