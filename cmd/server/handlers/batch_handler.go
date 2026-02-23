package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/yuno-payments/papaya-payout-engine/internal/platform/constants"
	"github.com/yuno-payments/papaya-payout-engine/internal/risk"
)

type BatchHandler struct {
	riskService   *risk.Service
	merchantStore risk.MerchantRepository
}

func NewBatchHandler(riskService *risk.Service, merchantStore risk.MerchantRepository) *BatchHandler {
	return &BatchHandler{
		riskService:   riskService,
		merchantStore: merchantStore,
	}
}

type BatchEvaluateRequest struct {
	MerchantIDs []string `json:"merchant_ids"`
	Simulation  bool     `json:"simulation"`
}

type EvaluationError struct {
	MerchantID string `json:"merchant_id"`
	Error      string `json:"error"`
}

func (h *BatchHandler) BatchEvaluate(c echo.Context) error {
	var req BatchEvaluateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if len(req.MerchantIDs) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "merchant_ids cannot be empty"})
	}

	if len(req.MerchantIDs) > constants.MaxBatchSize {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("batch size exceeds maximum of %d merchants", constants.MaxBatchSize),
		})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), constants.BatchTimeout)
	defer cancel()

	batchID := uuid.New()
	log.Printf("[INFO] Starting batch evaluation %s: %d merchants (simulation=%v)",
		batchID, len(req.MerchantIDs), req.Simulation)
	startTime := time.Now()

	decisions := make([]*risk.RiskDecision, 0, len(req.MerchantIDs))
	failedEvaluations := make([]EvaluationError, 0)
	var mu sync.Mutex

	workers := constants.DefaultWorkers
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
				mu.Lock()
				failedEvaluations = append(failedEvaluations, EvaluationError{
					MerchantID: merchantIDStr,
					Error:      "invalid UUID format",
				})
				mu.Unlock()
				return
			}

			decision, err := h.riskService.EvaluateMerchant(ctx, merchantID, req.Simulation)
			if err != nil {
				mu.Lock()
				failedEvaluations = append(failedEvaluations, EvaluationError{
					MerchantID: merchantIDStr,
					Error:      err.Error(),
				})
				mu.Unlock()
				return
			}

			decision.BatchID = &batchID

			mu.Lock()
			decisions = append(decisions, decision)
			mu.Unlock()
		}(idStr)
	}

	wg.Wait()

	duration := time.Since(startTime)
	log.Printf("[INFO] Batch %s completed in %v: %d successful, %d failed",
		batchID, duration, len(decisions), len(failedEvaluations))

	if len(decisions) == 0 {
		log.Printf("[WARN] Batch %s: all evaluations failed", batchID)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"batch_id":        batchID,
			"total_merchants": len(req.MerchantIDs),
			"successful":      0,
			"failed":          len(failedEvaluations),
			"errors":          failedEvaluations,
			"message":         "no merchants could be evaluated successfully",
			"simulation":      req.Simulation,
		})
	}

	summary := h.generateSummary(c.Request().Context(), decisions)
	highRiskMerchants := h.identifyHighRiskMerchants(decisions)

	if len(highRiskMerchants) > 0 {
		log.Printf("[WARN] Batch %s: %d high-risk merchants detected",
			batchID, len(highRiskMerchants))
	}

	response := map[string]interface{}{
		"batch_id":            batchID,
		"total_merchants":     len(req.MerchantIDs),
		"successful":          len(decisions),
		"failed":              len(failedEvaluations),
		"evaluated_at":        decisions[0].EvaluatedAt,
		"summary":             summary,
		"high_risk_merchants": highRiskMerchants,
		"errors":              failedEvaluations,
		"simulation":          req.Simulation,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *BatchHandler) generateSummary(ctx context.Context, decisions []*risk.RiskDecision) map[string]interface{} {
	byHoldPeriod := make(map[string]int)
	byReserve := make(map[string]int)
	byRiskLevel := make(map[string]int)
	volumeByTier := make(map[string]float64)
	totalVolume := 0.0

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

		merchant, err := h.merchantStore.Get(ctx, d.MerchantID)
		if err == nil {
			vol := merchant.TransactionVolume30d.InexactFloat64()
			totalVolume += vol
			volumeByTier[string(d.PayoutHoldPeriod)] += vol
		}
	}

	return map[string]interface{}{
		"by_hold_period": byHoldPeriod,
		"by_reserve":     byReserve,
		"by_risk_level":  byRiskLevel,
		"total_volume":   totalVolume,
		"volume_by_tier": volumeByTier,
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
