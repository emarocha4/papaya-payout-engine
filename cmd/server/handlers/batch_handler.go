package handlers

import (
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/yuno-payments/papaya-payout-engine/internal/risk"
)

type BatchHandler struct {
	riskService *risk.Service
}

func NewBatchHandler(riskService *risk.Service) *BatchHandler {
	return &BatchHandler{riskService: riskService}
}

type BatchEvaluateRequest struct {
	MerchantIDs []string `json:"merchant_ids"`
	Simulation  bool     `json:"simulation"`
}

func (h *BatchHandler) BatchEvaluate(c echo.Context) error {
	var req BatchEvaluateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	batchID := uuid.New()
	decisions := make([]*risk.RiskDecision, 0, len(req.MerchantIDs))
	var mu sync.Mutex

	workers := 10
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup

	for _, idStr := range req.MerchantIDs {
		wg.Add(1)
		go func(merchantIDStr string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			merchantID, err := uuid.Parse(merchantIDStr)
			if err != nil {
				return
			}

			decision, err := h.riskService.EvaluateMerchant(c.Request().Context(), merchantID, req.Simulation)
			if err != nil {
				return
			}

			decision.BatchID = &batchID

			mu.Lock()
			decisions = append(decisions, decision)
			mu.Unlock()
		}(idStr)
	}

	wg.Wait()

	summary := h.generateSummary(decisions)
	highRiskMerchants := h.identifyHighRiskMerchants(decisions)

	response := map[string]interface{}{
		"batch_id":            batchID,
		"total_merchants":     len(decisions),
		"evaluated_at":        decisions[0].EvaluatedAt,
		"summary":             summary,
		"high_risk_merchants": highRiskMerchants,
		"simulation":          req.Simulation,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *BatchHandler) generateSummary(decisions []*risk.RiskDecision) map[string]interface{} {
	byHoldPeriod := make(map[string]int)
	byReserve := make(map[string]int)
	byRiskLevel := make(map[string]int)

	for _, d := range decisions {
		byHoldPeriod[string(d.PayoutHoldPeriod)]++
		reserveKey := "0_PERCENT"
		if d.RollingReservePercentage == 10 {
			reserveKey = "10_PERCENT"
		} else if d.RollingReservePercentage == 20 {
			reserveKey = "20_PERCENT"
		}
		byReserve[reserveKey]++
		byRiskLevel[string(d.RiskLevel)]++
	}

	return map[string]interface{}{
		"by_hold_period": byHoldPeriod,
		"by_reserve":     byReserve,
		"by_risk_level":  byRiskLevel,
	}
}

func (h *BatchHandler) identifyHighRiskMerchants(decisions []*risk.RiskDecision) []map[string]interface{} {
	highRisk := make([]map[string]interface{}, 0)

	for _, d := range decisions {
		if d.RiskScore > 60 {
			concerns := make([]string, 0)
			for _, factor := range d.Reasoning.PrimaryFactors {
				if factor.Impact == "NEGATIVE" || factor.Impact == "CRITICAL" {
					concerns = append(concerns, factor.Contribution)
				}
			}

			action := "Manual review required"
			if d.RiskScore > 80 {
				action = "Immediate manual approval required"
			}

			highRisk = append(highRisk, map[string]interface{}{
				"merchant_id":         d.MerchantID,
				"risk_score":          d.RiskScore,
				"risk_level":          d.RiskLevel,
				"primary_concerns":    concerns,
				"recommended_action":  action,
			})
		}
	}

	return highRisk
}
