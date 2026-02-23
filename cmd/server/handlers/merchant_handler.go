package handlers

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/yuno-payments/papaya-payout-engine/internal/merchant"
)

type MerchantHandler struct {
	merchantService *merchant.Service
}

func NewMerchantHandler(merchantService *merchant.Service) *MerchantHandler {
	return &MerchantHandler{merchantService: merchantService}
}

type CreateMerchantRequest struct {
	MerchantName         string  `json:"merchant_name"`
	Industry             string  `json:"industry"`
	Country              string  `json:"country"`
	TransactionVolume30d float64 `json:"transaction_volume_30d"`
	TransactionCount30d  int     `json:"transaction_count_30d"`
	ChargebackCount30d   int     `json:"chargeback_count_30d"`
	RefundRate           float64 `json:"refund_rate"`
	AccountAgeDays       int     `json:"account_age_days"`
	KYCVerified          bool    `json:"kyc_verified"`
}

type SeedRequest struct {
	Count int `json:"count"`
}

type SeedResponse struct {
	MerchantsCreated int            `json:"merchants_created"`
	Distribution     map[string]int `json:"distribution"`
}

func (h *MerchantHandler) Create(c echo.Context) error {
	var req CreateMerchantRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	m := &merchant.Merchant{
		MerchantName:        req.MerchantName,
		Industry:            req.Industry,
		Country:             req.Country,
		AccountAgeDays:      req.AccountAgeDays,
		KYCVerified:         req.KYCVerified,
	}

	created, err := h.merchantService.Create(c.Request().Context(), m)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"merchant_id":   created.ID,
		"merchant_name": created.MerchantName,
		"created_at":    created.CreatedAt,
	})
}

func (h *MerchantHandler) Get(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid merchant ID"})
	}

	m, err := h.merchantService.Get(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "merchant not found"})
	}

	return c.JSON(http.StatusOK, m)
}

func (h *MerchantHandler) List(c echo.Context) error {
	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	merchants, total, err := h.merchantService.List(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"merchants": merchants,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

func (h *MerchantHandler) Seed(c echo.Context) error {
	var req SeedRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if req.Count <= 0 {
		req.Count = 100
	}

	merchants, err := h.merchantService.Seed(c.Request().Context(), req.Count)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	distribution := make(map[string]int)
	for _, m := range merchants {
		chargebackRate := m.ChargebackRate.InexactFloat64()
		if chargebackRate < 0.5 && m.AccountAgeDays > 365 {
			distribution["LOW"]++
		} else if chargebackRate < 1.5 && m.AccountAgeDays > 90 {
			distribution["MEDIUM"]++
		} else {
			distribution["HIGH"]++
		}
	}

	return c.JSON(http.StatusCreated, SeedResponse{
		MerchantsCreated: len(merchants),
		Distribution:     distribution,
	})
}
