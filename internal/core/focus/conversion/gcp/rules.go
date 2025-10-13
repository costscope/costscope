package gcp

import "github.com/costscope/costscope/internal/core/focus/types"

// GetMappingRules returns GCP field mapping rules for the unified mapper.
func GetMappingRules() *types.MappingRules {
	return &types.MappingRules{
		Provider: "gcp",
		Version:  "1.0",
		FieldMaps: map[string]types.FieldMapping{
			"billing_account_id": {SourceField: "billing_account_id", TargetField: "billing_account_id", Required: false},
			"effective_cost":     {SourceField: "cost", TargetField: "effective_cost", Required: true},
			"usage_quantity":     {SourceField: "usage.amount", TargetField: "usage_quantity", Required: false},
			"billing_currency":   {SourceField: "currency", TargetField: "billing_currency", Required: false},
			"usage_unit":         {SourceField: "usage.unit", TargetField: "usage_unit", Required: false},
			"charge_description": {SourceField: "sku.description", TargetField: "charge_description", Required: false},
			"service_name":       {SourceField: "service.description", TargetField: "service_name", Required: false},
			"charge_subcategory": {SourceField: "service.description", TargetField: "charge_subcategory", Required: false},
			"sku_id":             {SourceField: "sku.id", TargetField: "sku_id", Required: false},
			"sku_price_id":       {SourceField: "pricing.price_id", TargetField: "sku_price_id", Required: false},
			"sub_account_id":     {SourceField: "project.id", TargetField: "sub_account_id", Required: false},
			"sub_account_name":   {SourceField: "project.name", TargetField: "sub_account_name", Required: false},
			"resource_id":        {SourceField: "resource.name", TargetField: "resource_id", Required: false},
			"resource_name":      {SourceField: "resource.name", TargetField: "resource_name", Required: false},
		},
	}
}
