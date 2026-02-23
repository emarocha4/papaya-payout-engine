package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/yuno-payments/papaya-payout-engine/internal/merchant"
	"github.com/yuno-payments/papaya-payout-engine/internal/risk"
)

type mockRiskService struct {
	evaluateMerchant    func(ctx context.Context, merchantID uuid.UUID, simulation bool) (*risk.RiskDecision, error)
	simulateMerchant    func(ctx context.Context, merchantID uuid.UUID, overrides map[string]interface{}) (*risk.RiskDecision, error)
	getMerchantProfile  func(ctx context.Context, merchantID uuid.UUID) (*merchant.MerchantProfile, error)
}

func (m *mockRiskService) EvaluateMerchant(ctx context.Context, merchantID uuid.UUID, simulation bool) (*risk.RiskDecision, error) {
	if m.evaluateMerchant != nil {
		return m.evaluateMerchant(ctx, merchantID, simulation)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRiskService) SimulateMerchant(ctx context.Context, merchantID uuid.UUID, overrides map[string]interface{}) (*risk.RiskDecision, error) {
	if m.simulateMerchant != nil {
		return m.simulateMerchant(ctx, merchantID, overrides)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRiskService) GetMerchantProfile(ctx context.Context, merchantID uuid.UUID) (*merchant.MerchantProfile, error) {
	if m.getMerchantProfile != nil {
		return m.getMerchantProfile(ctx, merchantID)
	}
	return nil, errors.New("not implemented")
}

func TestRiskHandler_Evaluate(t *testing.T) {
	merchantID := uuid.New()

	t.Run("successful evaluation", func(t *testing.T) {
		service := &mockRiskService{
			evaluateMerchant: func(ctx context.Context, id uuid.UUID, simulation bool) (*risk.RiskDecision, error) {
				if id != merchantID {
					t.Errorf("expected merchant ID %v, got %v", merchantID, id)
				}
				if simulation {
					t.Error("expected simulation=false")
				}
				return &risk.RiskDecision{
					MerchantID:               id,
					RiskScore:                15,
					RiskLevel:                risk.RiskLevelLow,
					PayoutHoldPeriod:         risk.HoldPeriodImmediate,
					RollingReservePercentage: 0,
					EvaluatedAt:              time.Now(),
					Simulation:               false,
				}, nil
			},
		}

		handler := NewRiskHandler(service)
		e := echo.New()
		reqBody := `{"merchant_id":"` + merchantID.String() + `","simulation":false}`
		req := httptest.NewRequest(http.MethodPost, "/evaluate", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Evaluate(c)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		var response risk.RiskDecision
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if response.MerchantID != merchantID {
			t.Errorf("expected merchant ID %v, got %v", merchantID, response.MerchantID)
		}
	})

	t.Run("evaluation with simulation", func(t *testing.T) {
		service := &mockRiskService{
			evaluateMerchant: func(ctx context.Context, id uuid.UUID, simulation bool) (*risk.RiskDecision, error) {
				if !simulation {
					t.Error("expected simulation=true")
				}
				return &risk.RiskDecision{
					MerchantID: id,
					RiskScore:  15,
					Simulation: true,
				}, nil
			},
		}

		handler := NewRiskHandler(service)
		e := echo.New()
		reqBody := `{"merchant_id":"` + merchantID.String() + `","simulation":true}`
		req := httptest.NewRequest(http.MethodPost, "/evaluate", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Evaluate(c)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		handler := NewRiskHandler(&mockRiskService{})
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/evaluate", strings.NewReader("{invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Evaluate(c)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("invalid merchant ID", func(t *testing.T) {
		handler := NewRiskHandler(&mockRiskService{})
		e := echo.New()
		reqBody := `{"merchant_id":"not-a-uuid","simulation":false}`
		req := httptest.NewRequest(http.MethodPost, "/evaluate", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Evaluate(c)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("service error", func(t *testing.T) {
		service := &mockRiskService{
			evaluateMerchant: func(ctx context.Context, id uuid.UUID, simulation bool) (*risk.RiskDecision, error) {
				return nil, errors.New("service error")
			},
		}

		handler := NewRiskHandler(service)
		e := echo.New()
		reqBody := `{"merchant_id":"` + merchantID.String() + `","simulation":false}`
		req := httptest.NewRequest(http.MethodPost, "/evaluate", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Evaluate(c)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}
	})
}

func TestRiskHandler_Simulate(t *testing.T) {
	merchantID := uuid.New()

	t.Run("successful simulation", func(t *testing.T) {
		service := &mockRiskService{
			simulateMerchant: func(ctx context.Context, id uuid.UUID, overrides map[string]interface{}) (*risk.RiskDecision, error) {
				if id != merchantID {
					t.Errorf("expected merchant ID %v, got %v", merchantID, id)
				}
				if overrides["chargeback_rate"] != 2.5 {
					t.Errorf("expected chargeback_rate 2.5, got %v", overrides["chargeback_rate"])
				}
				return &risk.RiskDecision{
					MerchantID: id,
					RiskScore:  35,
					Simulation: true,
				}, nil
			},
		}

		handler := NewRiskHandler(service)
		e := echo.New()
		reqBody := `{"merchant_id":"` + merchantID.String() + `","overrides":{"chargeback_rate":2.5}}`
		req := httptest.NewRequest(http.MethodPost, "/simulate", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Simulate(c)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		handler := NewRiskHandler(&mockRiskService{})
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/simulate", strings.NewReader("{invalid"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Simulate(c)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("invalid merchant ID", func(t *testing.T) {
		handler := NewRiskHandler(&mockRiskService{})
		e := echo.New()
		reqBody := `{"merchant_id":"invalid","overrides":{}}`
		req := httptest.NewRequest(http.MethodPost, "/simulate", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Simulate(c)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})
}

func TestRiskHandler_GetProfile(t *testing.T) {
	merchantID := uuid.New()

	t.Run("successful profile retrieval", func(t *testing.T) {
		service := &mockRiskService{
			getMerchantProfile: func(ctx context.Context, id uuid.UUID) (*merchant.MerchantProfile, error) {
				if id != merchantID {
					t.Errorf("expected merchant ID %v, got %v", merchantID, id)
				}
				return &merchant.MerchantProfile{
					MerchantID:   id,
					MerchantName: "Test Merchant",
					Industry:     "RETAIL",
					Country:      "US",
					RiskMetrics: merchant.RiskMetrics{
						TransactionVolume30d: decimal.NewFromFloat(100000),
						ChargebackRate:       decimal.NewFromFloat(0.5),
					},
				}, nil
			},
		}

		handler := NewRiskHandler(service)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/profile/"+merchantID.String(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(merchantID.String())

		err := handler.GetProfile(c)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		var profile merchant.MerchantProfile
		if err := json.Unmarshal(rec.Body.Bytes(), &profile); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if profile.MerchantID != merchantID {
			t.Errorf("expected merchant ID %v, got %v", merchantID, profile.MerchantID)
		}
	})

	t.Run("invalid merchant ID", func(t *testing.T) {
		handler := NewRiskHandler(&mockRiskService{})
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/profile/invalid", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("invalid-uuid")

		err := handler.GetProfile(c)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("service error", func(t *testing.T) {
		service := &mockRiskService{
			getMerchantProfile: func(ctx context.Context, id uuid.UUID) (*merchant.MerchantProfile, error) {
				return nil, errors.New("service error")
			},
		}

		handler := NewRiskHandler(service)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/profile/"+merchantID.String(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(merchantID.String())

		err := handler.GetProfile(c)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}
	})
}
