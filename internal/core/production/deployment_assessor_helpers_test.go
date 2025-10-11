package production

import (
	"testing"

	"local/costscope/internal/core/logging"
)

func TestAssessPerformanceTests_TargetsInfluence(t *testing.T) {
	ass := NewBasicDeploymentAssessor(logging.NewLogger(logging.LevelError))
	// When no special targets, expect base keys
	r := &DeploymentRequirements{Environment: "production", PerformanceTargets: map[string]float64{}}
	perf := ass.assessPerformanceTests(r)
	if _, ok := perf["spike_test"]; !ok {
		t.Fatalf("expected spike_test present")
	}

	// With max_latency key present should add latency_validation
	r2 := &DeploymentRequirements{Environment: "production", PerformanceTargets: map[string]float64{"max_latency": 200}}
	perf2 := ass.assessPerformanceTests(r2)
	if _, ok := perf2["latency_validation"]; !ok {
		t.Fatalf("expected latency_validation present when target set")
	}

	// With min_throughput should add throughput_validation
	r3 := &DeploymentRequirements{Environment: "production", PerformanceTargets: map[string]float64{"min_throughput": 1000}}
	perf3 := ass.assessPerformanceTests(r3)
	if _, ok := perf3["throughput_validation"]; !ok {
		t.Fatalf("expected throughput_validation present when target set")
	}
}

func TestAssessSecurityValidation_ComplianceStandardsAdded(t *testing.T) {
	ass := NewBasicDeploymentAssessor(logging.NewLogger(logging.LevelError))
	r := &DeploymentRequirements{SecurityRequirements: []string{"soc2", "pci_dss"}, ComplianceStandards: []string{"iso27001", "gdpr"}}
	sec := ass.assessSecurityValidation(r)
	// expect mapped keys for provided requirements
	if _, ok := sec["soc2_compliance"]; !ok {
		t.Fatalf("expected soc2_compliance present")
	}
	if _, ok := sec["pci_compliance"]; !ok {
		t.Fatalf("expected pci_compliance present")
	}
	if _, ok := sec["iso27001_validation"]; !ok {
		t.Fatalf("expected iso27001_validation present")
	}
	if _, ok := sec["gdpr_validation"]; !ok {
		t.Fatalf("expected gdpr_validation present")
	}
}

func TestGenerateBlockingIssues_NegativeBranches(t *testing.T) {
	ass := NewBasicDeploymentAssessor(logging.NewLogger(logging.LevelError))
	// Setup maps that cause each blocking issue to be triggered
	health := map[string]bool{"messaging_health": false}
	perf := map[string]bool{"spike_test": false}
	sec := map[string]bool{"vulnerability_scan": false, "compliance_check": false}
	env := map[string]bool{"dependencies": false}

	issues := ass.generateBlockingIssues(health, perf, sec, env)
	if len(issues) != 5 {
		t.Fatalf("expected 5 blocking issues, got %d", len(issues))
	}
}
