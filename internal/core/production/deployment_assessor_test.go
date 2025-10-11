package production

import (
	"context"
	"local/costscope/internal/core/logging"
	"testing"
	"time"
)

func newTestAssessor() *BasicDeploymentAssessor {
	return NewBasicDeploymentAssessor(logging.NewLogger(logging.LevelError))
}

func TestDeploymentAssessor_ValidateEnvironmentVariants(t *testing.T) {
	ass := newTestAssessor()
	envs := []string{"production", "staging", "development", "unknown"}
	for _, e := range envs {
		v, err := ass.ValidateEnvironment(context.Background(), e)
		if err != nil {
			t.Fatalf("validate %s err: %v", e, err)
		}
		if v.Environment != e {
			t.Fatalf("env mismatch")
		}
		// Unknown should produce <=0 score
		if e == "unknown" && v.ValidationScore != 0 {
			t.Fatalf("expected zero score for unknown")
		}
	}
}

func TestDeploymentAssessor_AssessReadiness_OKAndError(t *testing.T) {
	ass := newTestAssessor()
	// nil requirements error path
	if _, err := ass.AssessReadiness(context.Background(), nil); err == nil {
		t.Fatalf("expected error for nil requirements")
	}
	req := &DeploymentRequirements{Environment: "production", MinHealthScore: 80, RequiredChecks: []string{"api_health"}, MonitoringEnabled: true, BackupRequired: true}
	r, err := ass.AssessReadiness(context.Background(), req)
	if err != nil {
		t.Fatalf("assess readiness err: %v", err)
	}
	if r.ReadinessScore == 0 || r.ReadinessStatus == "" {
		t.Fatalf("expected populated readiness fields")
	}
	if len(r.HealthChecks) == 0 || len(r.SecurityValidation) == 0 {
		t.Fatalf("expected health & security maps")
	}
}

func TestDeploymentAssessor_GenerateDeploymentPlan_Strategies(t *testing.T) {
	ass := newTestAssessor()
	strategies := []string{"blue_green", "rolling", "canary", "custom"}
	req := &DeploymentRequirements{Environment: "production"}
	for _, s := range strategies {
		p, err := ass.GenerateDeploymentPlan(context.Background(), s, req)
		if err != nil {
			t.Fatalf("plan %s err: %v", s, err)
		}
		if p.Strategy != s {
			t.Fatalf("strategy mismatch")
		}
		if len(p.Steps) == 0 || p.EstimatedDuration <= 0 {
			t.Fatalf("expected steps and duration for %s", s)
		}
		// Approval rules
		if (s == "blue_green" || s == "canary") && !p.ApprovalRequired {
			t.Fatalf("expected approval for %s", s)
		}
	}
	// nil requirements error branch
	if _, err := ass.GenerateDeploymentPlan(context.Background(), "blue_green", nil); err == nil {
		t.Fatalf("expected error for nil reqs")
	}
}

func TestDeploymentAssessor_RunHealthChecks(t *testing.T) {
	ass := newTestAssessor()
	comps := []string{"database", "cache", "api", "storage", "messaging", "monitoring", "custom"}
	res, err := ass.RunHealthChecks(context.Background(), comps)
	if err != nil {
		t.Fatalf("health checks err: %v", err)
	}
	if res.OverallHealthScore == 0 {
		t.Fatalf("expected non-zero score")
	}
	if len(res.ComponentResults) != len(comps) {
		t.Fatalf("component count mismatch")
	}
}

// Light timing sanity for calculation rounding path (implicit via AssessReadiness + RunHealthChecks) – ensure timestamp not zero.
func TestDeploymentAssessor_Timestamps(t *testing.T) {
	ass := newTestAssessor()
	req := &DeploymentRequirements{Environment: "production", MonitoringEnabled: true}
	r, _ := ass.AssessReadiness(context.Background(), req)
	if time.Since(r.AssessmentTimestamp) > time.Minute {
		t.Fatalf("unexpected stale timestamp")
	}
	plan, _ := ass.GenerateDeploymentPlan(context.Background(), "custom", req)
	if time.Since(plan.PlanCreatedAt) > time.Minute {
		t.Fatalf("plan timestamp invalid")
	}
}
