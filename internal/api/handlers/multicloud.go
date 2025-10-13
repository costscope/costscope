package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/costscope/costscope/internal/api/response"
	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/core/multicloud"
	"github.com/costscope/costscope/internal/providers"
)

// MulticloudHandler provides HTTP endpoints for multicloud advanced operations
type MulticloudHandler struct {
	logger  *logging.Logger
	service *multicloud.MulticloudService
}

// NewMulticloudHandler constructs a new handler (lightweight stub data)
func NewMulticloudHandler(logger *logging.Logger, providerManager *providers.ProviderManager) *MulticloudHandler {
	return &MulticloudHandler{logger: logger, service: multicloud.NewMulticloudService(providerManager, logger)}
}

// NOTE: Route registration for multicloud endpoints is handled centrally:
//   * Basic (net/http) server: cmd/modules/api/routespec.go
//   * Enterprise (gin) server: cmd/modules/api/enterprise.go (GinRouteGroup under /multicloud)
// The former helper RegisterMulticloudRoutes has been removed to avoid duplicate route wiring.
// Tests now register handler methods directly when needed.

// Recommendations returns optimization recommendations (mock/demo)
func (h *MulticloudHandler) Recommendations(c *gin.Context) {
	var req struct {
		Providers          []string `json:"providers"`
		RiskTolerance      string   `json:"risk_tolerance"`
		SavingsThreshold   float64  `json:"savings_threshold"`
		MaxRecommendations int      `json:"max_recommendations"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AutoBadRequest(c, "invalid request")
		return
	}
	request := &multicloud.RecommendationRequest{Providers: req.Providers, SavingsThreshold: req.SavingsThreshold, MaxRecommendations: req.MaxRecommendations}
	res, err := h.service.GetRecommendations(context.Background(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response.AutoOK200(c, gin.H{"recommendations": res})
}

// Inventory returns unified inventory summary
func (h *MulticloudHandler) Inventory(c *gin.Context) {
	inv, err := h.service.GetUnifiedInventory(context.Background(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response.AutoOK200(c, gin.H{"inventory": inv})
}

// MigrationPlan generates a migration plan
func (h *MulticloudHandler) MigrationPlan(c *gin.Context) {
	var req struct {
		Source string `json:"source"`
		Target string `json:"target"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AutoBadRequest(c, "invalid request")
		return
	}
	mreq := &multicloud.MigrationRequest{SourceProvider: req.Source, TargetProvider: req.Target, MigrationTimeframe: 30 * 24 * time.Hour, SourceRegion: "us-east-1", TargetRegion: "us-east-1", IncludeDataTransfer: true, MigrationStrategy: multicloud.MigrationStrategyLiftAndShift}
	plan, err := h.service.GenerateMigrationPlan(context.Background(), mreq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response.AutoOK200(c, gin.H{"plan": plan})
}

// MigrationFeasibility analyzes migration feasibility
func (h *MulticloudHandler) MigrationFeasibility(c *gin.Context) {
	var req struct {
		Source string `json:"source"`
		Target string `json:"target"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AutoBadRequest(c, "invalid request")
		return
	}
	mreq := &multicloud.MigrationRequest{SourceProvider: req.Source, TargetProvider: req.Target, MigrationTimeframe: 30 * 24 * time.Hour, SourceRegion: "us-east-1", TargetRegion: "us-east-1", IncludeDataTransfer: true, MigrationStrategy: multicloud.MigrationStrategyLiftAndShift}
	feas, err := h.service.AnalyzeMigrationFeasibility(context.Background(), mreq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response.AutoOK200(c, gin.H{"feasibility": feas})
}
