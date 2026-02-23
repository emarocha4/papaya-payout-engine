package server

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/yuno-payments/papaya-payout-engine/cmd/server/handlers"
)

func setupRoutes(e *echo.Echo, h *Handlers) {
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/health-check", h.Health.HealthCheck)

	api := e.Group("/papaya-payout-engine/v1")

	api.POST("/merchants", h.Merchant.Create)
	api.GET("/merchants/:id", h.Merchant.Get)
	api.GET("/merchants", h.Merchant.List)
	api.POST("/merchants/seed", h.Merchant.Seed)

	api.POST("/risk/evaluate", h.Risk.Evaluate)
	api.POST("/risk/simulate", h.Risk.Simulate)
	api.GET("/risk/merchants/:id/profile", h.Risk.GetProfile)

	api.POST("/risk/batch-evaluate", h.Batch.BatchEvaluate)
}

type Handlers struct {
	Health   *handlers.HealthHandler
	Merchant *handlers.MerchantHandler
	Risk     *handlers.RiskHandler
	Batch    *handlers.BatchHandler
}
