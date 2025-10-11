package api

import (
	"net/http"
	"testing"

	"local/costscope/internal/core/logging"
)

func TestWrapServerWithCasbinIfEnabled_DisabledNoop(t *testing.T) {
	prev := enterpriseCasbinEnabled
	enterpriseCasbinEnabled = false
	defer func() { enterpriseCasbinEnabled = prev }()

	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	logger := logging.NewLogger(logging.LevelInfo)
	// Should be no-op and not panic
	wrapServerWithCasbinIfEnabled(&h, logger)
}

func TestWrapServerWithCasbinIfEnabled_ModelPolicyMissing_Noop(t *testing.T) {
	prev := enterpriseCasbinEnabled
	prevModel := enterpriseCasbinModelPath
	prevPolicy := enterpriseCasbinPolicyPath
	enterpriseCasbinEnabled = true
	enterpriseCasbinModelPath = ""
	enterpriseCasbinPolicyPath = ""
	defer func() {
		enterpriseCasbinEnabled = prev
		enterpriseCasbinModelPath = prevModel
		enterpriseCasbinPolicyPath = prevPolicy
	}()

	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	logger := logging.NewLogger(logging.LevelInfo)
	wrapServerWithCasbinIfEnabled(&h, logger)
}
