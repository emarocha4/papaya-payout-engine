package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yuno-payments/papaya-payout-engine/internal/health"
)

type HealthHandler struct {
	healthService *health.Service
}

func NewHealthHandler(healthService *health.Service) *HealthHandler {
	return &HealthHandler{healthService: healthService}
}

func (h *HealthHandler) HealthCheck(c echo.Context) error {
	response, err := h.healthService.Check()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response)
	}
	return c.JSON(http.StatusOK, response)
}
