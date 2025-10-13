package azure

import "github.com/costscope/costscope/internal/core/focus/types"

// getAzureMappingRules returns Azure to FOCUS mapping rules (minimal starter).
func getAzureMappingRules() *types.MappingRules {
	return &types.MappingRules{
		Provider: "azure",
		Version:  "1.0",
		FieldMaps: map[string]types.FieldMapping{
			"billing_account_id": {SourceField: "BillingAccountId", TargetField: "billing_account_id", Required: false},
			"sub_account_id":     {SourceField: "SubscriptionId", TargetField: "sub_account_id", Required: true},
			"service_name":       {SourceField: "ServiceName", TargetField: "service_name", Required: true},
			"usage_quantity":     {SourceField: "Quantity", TargetField: "usage_quantity", Required: true},
			"usage_unit":         {SourceField: "UnitOfMeasure", TargetField: "usage_unit", Required: true},
			"effective_cost":     {SourceField: "AmortizedCost", TargetField: "effective_cost", Required: false},
		},
	}
}
