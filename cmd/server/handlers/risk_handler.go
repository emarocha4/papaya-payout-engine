package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/yuno-payments/papaya-payout-engine/internal/risk"
)

type RiskHandler struct {
	riskService *risk.Service
}

func NewRiskHandler(riskService *risk.Service) *RiskHandler {
	return &RiskHandler{riskService: riskService}
}

type EvaluateRequest struct {
	MerchantID string `json:"merchant_id"`
	Simulation bool   `json:"simulation"`
}

type SimulateRequest struct {
	MerchantID string                 `json:"merchant_id"`
	Overrides  map[string]interface{} `json:"overrides"`
}

func (h *RiskHandler) Evaluate(c echo.Context) error {
	var req EvaluateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	merchantID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid merchant ID"})
	}

	decision, err := h.riskService.EvaluateMerchant(c.Request().Context(), merchantID, req.Simulation)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, decision)
}

func (h *RiskHandler) Simulate(c echo.Context) error {
	var req SimulateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	merchantID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid merchant ID"})
	}

	decision, err := h.riskService.SimulateMerchant(c.Request().Context(), merchantID, req.Overrides)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, decision)
}

func (h *RiskHandler) GetProfile(c echo.Context) error {
	idStr := c.Param("id")
	merchantID, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid merchant ID"})
	}

	profile, err := h.riskService.GetMerchantProfile(c.Request().Context(), merchantID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, profile)
}
