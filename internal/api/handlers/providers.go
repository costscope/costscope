package handlers

import (
	"github.com/gin-gonic/gin"

	"local/costscope/internal/api/response"
	"local/costscope/internal/core/logging"
)

// =====================================================================================
// Providers Handler - Cloud Provider Management
// =====================================================================================

// ProvidersHandler provides cloud provider management endpoints
type ProvidersHandler struct {
	logger *logging.Logger
}

// NewProvidersHandler creates a new providers handler
func NewProvidersHandler(logger *logging.Logger) *ProvidersHandler {
	return &ProvidersHandler{
		logger: logger,
	}
}

// ListProviders returns available cloud providers
func (h *ProvidersHandler) ListProviders(c *gin.Context) {
	providers := []gin.H{
		{
			"id":          "aws",
			"name":        "Amazon Web Services",
			"status":      "available",
			"features":    []string{"cur", "cost_explorer", "billing_alerts"},
			"auth_method": "iam_role",
		},
		{
			"id":          "azure",
			"name":        "Microsoft Azure",
			"status":      "available",
			"features":    []string{"cost_management", "billing_api"},
			"auth_method": "service_principal",
		},
		{
			"id":          "gcp",
			"name":        "Google Cloud Platform",
			"status":      "available",
			"features":    []string{"billing_export", "cloud_billing_api"},
			"auth_method": "service_account",
		},
	}

	// Unified response envelope (additive): existing shape nested under data
	type payload struct {
		Providers []gin.H `json:"providers"`
		Total     int     `json:"total"`
	}
	resp := payload{Providers: providers, Total: len(providers)}
	response.AutoOK200(c, resp)
}

// GetProvider returns specific provider details
func (h *ProvidersHandler) GetProvider(c *gin.Context) {
	provider := c.Param("provider")

	// Mock provider details
	providerData := gin.H{
		"id":         provider,
		"name":       "Provider " + provider,
		"status":     "available",
		"connection": "disconnected",
		"last_sync":  nil,
		"accounts":   []string{},
		"regions":    []string{},
		"services":   []string{},
	}

	response.AutoOK200(c, providerData)
}

// ConnectProvider establishes connection to a cloud provider
func (h *ProvidersHandler) ConnectProvider(c *gin.Context) {
	provider := c.Param("provider")

	var request struct {
		Credentials map[string]string      `json:"credentials" binding:"required"`
		Settings    map[string]interface{} `json:"settings"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		response.AutoBadRequest(c, "invalid request format")
		return
	}

	h.logger.InfoWithFields("Provider connection requested", map[string]interface{}{
		"provider": provider,
	})

	resp := struct {
		Provider string `json:"provider"`
		Status   string `json:"status"`
		Message  string `json:"message"`
	}{Provider: provider, Status: "connected", Message: "Provider connection established"}

	// 201 Created with request id injection via AutoCreated201.
	response.AutoCreated201(c, resp)
}

// TestConnection tests the connection to a cloud provider
func (h *ProvidersHandler) TestConnection(c *gin.Context) {
	provider := c.Param("provider")

	// Mock connection test
	resp := struct {
		Provider string `json:"provider"`
		Status   string `json:"status"`
		Latency  string `json:"latency"`
		Message  string `json:"message"`
	}{Provider: provider, Status: "success", Latency: "150ms", Message: "Connection test successful"}
	response.AutoOK200(c, resp)
}

// ListAccounts returns accounts for a provider
func (h *ProvidersHandler) ListAccounts(c *gin.Context) {
	provider := c.Param("provider")

	// Mock accounts
	accounts := []gin.H{
		{
			"id":     "123456789",
			"name":   "Production Account",
			"status": "active",
		},
		{
			"id":     "987654321",
			"name":   "Development Account",
			"status": "active",
		},
	}

	resp := struct {
		Provider string  `json:"provider"`
		Accounts []gin.H `json:"accounts"`
		Total    int     `json:"total"`
	}{Provider: provider, Accounts: accounts, Total: len(accounts)}
	response.AutoOK200(c, resp)
}

// ListRegions returns regions for a provider
func (h *ProvidersHandler) ListRegions(c *gin.Context) {
	provider := c.Param("provider")

	// Mock regions
	regions := []gin.H{
		{"id": "us-east-1", "name": "US East (N. Virginia)"},
		{"id": "us-west-2", "name": "US West (Oregon)"},
		{"id": "eu-west-1", "name": "Europe (Ireland)"},
	}

	resp := struct {
		Provider string  `json:"provider"`
		Regions  []gin.H `json:"regions"`
		Total    int     `json:"total"`
	}{Provider: provider, Regions: regions, Total: len(regions)}
	response.AutoOK200(c, resp)
}

// ListServices returns services for a provider
func (h *ProvidersHandler) ListServices(c *gin.Context) {
	provider := c.Param("provider")

	// Mock services
	services := []gin.H{
		{"id": "ec2", "name": "Elastic Compute Cloud", "category": "compute"},
		{"id": "s3", "name": "Simple Storage Service", "category": "storage"},
		{"id": "rds", "name": "Relational Database Service", "category": "database"},
	}

	resp := struct {
		Provider string  `json:"provider"`
		Services []gin.H `json:"services"`
		Total    int     `json:"total"`
	}{Provider: provider, Services: services, Total: len(services)}
	response.OK200(c, resp)
}
