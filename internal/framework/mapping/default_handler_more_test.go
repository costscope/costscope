package mapping

import (
	ftypes "local/costscope/internal/core/focus/types"
	"testing"
)

func TestUniversalDefaultHandler_ProviderBranchesAndPrecedence(t *testing.T) {
	// Azure
	azureCfg := &FieldMappingConfig{ProviderName: "azure", DefaultValues: map[string]interface{}{"ChargeClass": "Override"}}
	hAzure := NewUniversalDefaultHandler(azureCfg)
	if v := hAzure.GetDefaultValue("ChargeClass", FieldTypeString); v != "Override" {
		t.Fatalf("explicit default should win: %v", v)
	}
	if v := hAzure.GetDefaultValue("BillingCurrency", FieldTypeString); v != "USD" {
		t.Fatalf("azure provider default missing: %v", v)
	}

	// GCP
	gcpCfg := &FieldMappingConfig{ProviderName: "gcp", DefaultValues: map[string]interface{}{}}
	hGCP := NewUniversalDefaultHandler(gcpCfg)
	if v := hGCP.GetDefaultValue("ProviderName", FieldTypeString); v != ftypes.ProviderNames.GCP {
		t.Fatalf("gcp provider name default: %v", v)
	}

	// Generic / unknown
	genCfg := &FieldMappingConfig{ProviderName: "other"}
	hGen := NewUniversalDefaultHandler(genCfg)
	if v := hGen.GetDefaultValue("ChargeCategory", FieldTypeString); v == "" {
		t.Fatalf("generic defaults not applied")
	}
	// Type defaults fallback (field not present anywhere)
	if v := hGen.GetDefaultValue("NonExistingFloat", FieldTypeFloat64); v != 0.0 {
		t.Fatalf("float type default")
	}
	if v := hGen.GetDefaultValue("NonExistingInt", FieldTypeInt64); v != int64(0) {
		t.Fatalf("int type default")
	}
	if v := hGen.GetDefaultValue("NonExistingBool", FieldTypeBool); v != false {
		t.Fatalf("bool type default")
	}
	if v := hGen.GetDefaultValue("NonExistingTime", FieldTypeTime); v == nil {
		t.Fatalf("time type default should be zero time value")
	}
	if v := hGen.GetDefaultValue("NonExistingOptional", FieldTypeOptional); v != nil {
		t.Fatalf("optional default should be nil")
	}
}

func TestUniversalDefaultHandler_isZeroValueBranches(t *testing.T) {
	cfg := &FieldMappingConfig{ProviderName: "aws"}
	h := NewUniversalDefaultHandler(cfg)
	rec := &ftypes.FocusRecord{} // zero record triggers defaults
	if err := h.ApplyDefaults(rec); err != nil {
		t.Fatalf("apply defaults err: %v", err)
	}
	if rec.ProviderName == "" || rec.BillingCurrency == "" {
		t.Fatalf("expected provider defaults set")
	}
}
