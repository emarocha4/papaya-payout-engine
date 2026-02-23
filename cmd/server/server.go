package server

import (
	"context"
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/yuno-payments/papaya-payout-engine/cmd/server/handlers"
	"github.com/yuno-payments/papaya-payout-engine/internal/health"
	"github.com/yuno-payments/papaya-payout-engine/internal/merchant"
	"github.com/yuno-payments/papaya-payout-engine/internal/platform/config"
	"github.com/yuno-payments/papaya-payout-engine/internal/platform/database"
	"github.com/yuno-payments/papaya-payout-engine/internal/risk"
	"github.com/yuno-payments/papaya-payout-engine/internal/store"
	"gorm.io/gorm"
)

type Server struct {
	config   *config.Config
	db       *gorm.DB
	echo     *echo.Echo
	handlers *Handlers
}

func NewServer() (*Server, error) {
	cfg := config.Load()

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	merchantStore := store.NewMerchantStore(db)
	decisionStore := store.NewDecisionStore(db)

	merchantService := merchant.NewService(merchantStore)
	riskService := risk.NewService(merchantStore, decisionStore)
	healthService := health.NewService(db)

	h := &Handlers{
		Health:   handlers.NewHealthHandler(healthService),
		Merchant: handlers.NewMerchantHandler(merchantService),
		Risk:     handlers.NewRiskHandler(riskService),
		Batch:    handlers.NewBatchHandler(riskService, merchantStore),
	}

	e := echo.New()
	setupRoutes(e, h)

	return &Server{
		config:   cfg,
		db:       db,
		echo:     e,
		handlers: h,
	}, nil
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.config.Port)
	log.Printf("Starting server on %s", addr)
	return s.echo.Start(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down HTTP server...")
	if err := s.echo.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	log.Println("Closing database connections...")
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	return nil
}
