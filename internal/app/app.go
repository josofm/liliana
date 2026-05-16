package app

import (
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
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
	db, err := openDatabase(cfg)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - openDatabase: %w", err))
	}
	defer db.Close()

	userRepo := userRepo.NewPostgresRepo(db)
	deckRepo := deckRepo.NewPostgresRepo(db)

	// Passar a configuração para o router
	v1.NewRouter(handler, l, userRepo, deckRepo, cfg)

	httpServer := httpserver.New(handler)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: %s", s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}

func openDatabase(cfg *config.Config) (*sql.DB, error) {
	if cfg.DB.URL == "" {
		return nil, fmt.Errorf("database url is required")
	}

	db, err := sql.Open("pgx", cfg.DB.URL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
