package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	integration "local/costscope/cmd/modules/integration"
	"local/costscope/internal/core/logging"

	// Generated API stubs (if present)
	apiconnections "local/costscope/internal/api/handlers/integration/connections"
	apidashboard "local/costscope/internal/api/handlers/integration/dashboard"
	apiwebhook "local/costscope/internal/api/handlers/integration/webhook"
)

// buildIntegrationActionRoutes generates Gin routes for Integration Action DSL specs.
// It prefers concrete generated stubs when available, otherwise falls back to a generic 501.
func buildIntegrationActionRoutes(logger *logging.Logger) []GinRoute {
	// Map known actions to their concrete stub handlers
	handlersMap := map[string]gin.HandlerFunc{}

	// Webhook category stubs
	if h := apiwebhook.NewCreateHandler(logger); h != nil {
		handlersMap["/webhook/create"] = h.HandleCreate
	}
	if h := apiwebhook.NewListHandler(logger); h != nil {
		handlersMap["/webhook/list"] = h.HandleList // non-group list
	}
	if h := apiwebhook.NewTestHandler(logger); h != nil {
		handlersMap["/webhook/test"] = h.HandleTest
	}
	if h := apiwebhook.NewDeleteHandler(logger); h != nil {
		handlersMap["/webhook/delete"] = h.HandleDelete
	}
	if h := apiwebhook.NewRetryHandler(logger); h != nil {
		handlersMap["/webhook/delivery/retry"] = h.HandleRetry
	}
	if h := apiwebhook.NewStatsHandler(logger); h != nil {
		handlersMap["/webhook/delivery/stats"] = h.HandleStats
	}
	if h := apiwebhook.NewTriggerHandler(logger); h != nil {
		handlersMap["/webhook/event/trigger"] = h.HandleTrigger
	}

	// Dashboard category stubs
	if h := apidashboard.NewStartHandler(logger); h != nil {
		handlersMap["/dashboard/start"] = h.HandleStart
	}
	if h := apidashboard.NewStatusHandler(logger); h != nil {
		handlersMap["/dashboard/status"] = h.HandleStatus
	}
	if h := apidashboard.NewStopHandler(logger); h != nil {
		handlersMap["/dashboard/stop"] = h.HandleStop
	}
	if h := apidashboard.NewShowHandler(logger); h != nil {
		handlersMap["/dashboard/config/show"] = h.HandleShow
	}
	if h := apidashboard.NewSetHandler(logger); h != nil {
		handlersMap["/dashboard/config/set"] = h.HandleSet
	}
	if h := apidashboard.NewResetHandler(logger); h != nil {
		handlersMap["/dashboard/config/reset"] = h.HandleReset
	}
	if h := apidashboard.NewAddHandler(logger); h != nil {
		handlersMap["/dashboard/widget/add"] = h.HandleAdd
	}
	if h := apidashboard.NewListHandler(logger); h != nil {
		handlersMap["/dashboard/widget/list"] = h.HandleList
	}
	if h := apidashboard.NewRemoveHandler(logger); h != nil {
		handlersMap["/dashboard/widget/remove"] = h.HandleRemove
	}
	if h := apidashboard.NewConfigureHandler(logger); h != nil {
		handlersMap["/dashboard/widget/configure"] = h.HandleConfigure
	}
	if h := apidashboard.NewInstallHandler(logger); h != nil {
		handlersMap["/dashboard/plugin/install"] = h.HandleInstall
	}
	if h := apidashboard.NewEnableHandler(logger); h != nil {
		handlersMap["/dashboard/plugin/enable"] = h.HandleEnable
	}
	if h := apidashboard.NewDisableHandler(logger); h != nil {
		handlersMap["/dashboard/plugin/disable"] = h.HandleDisable
	}

	// Connections category stubs
	if h := apiconnections.NewConnectHandler(logger); h != nil {
		handlersMap["/connections/connect"] = h.HandleConnect
	}

	// Fallback handler (501 Not Implemented) preserving action metadata
	notImplemented := func(actionID, category string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{
				"status":    "not_implemented",
				"action_id": actionID,
				"category":  category,
			})
		}
	}

	var routes []GinRoute
	specs := integration.BuildDefaultActionSpecs()
	for _, s := range specs {
		if s.Group {
			continue // groups don't expose endpoints
		}
		// Build path including parent group segments to avoid collisions, e.g. /webhook/delivery/list
		segments := append([]string{s.Category}, append(s.Parents, s.Use)...)
		path := "/" + strings.Join(segments, "/")
		if h, ok := handlersMap[path]; ok {
			routes = append(routes, GinRoute{Method: http.MethodPost, Path: path, Handler: h})
			continue
		}
		// default 501 stub
		routes = append(routes, GinRoute{Method: http.MethodPost, Path: path, Handler: notImplemented(s.ID, s.Category)})
	}
	return routes
}
