# Azure Billing → FOCUS Mapping

## Scope
- Cost Management / Usage Details (MCA, EA, CSP)
- PriceSheet and Meter metadata
- Reservations and Savings Plans (Benefits)
- Tags, ResourceLocation/Location

## Field Mapping (Usage Details → FOCUS)

- Billing
  - billing_account_id ← BillingAccountId | EnrollmentNumber | BillingProfileId
  - billing_account_name ← BillingAccountName | BillingProfileName
  - billing_currency ← upper(BillingCurrency | Currency)
  - billing_period_start/end ← UsageStart/UsageEnd | Date/UsageDate (+24h)

- Service/SKU
  - service_name ← ServiceName | Product
  - service_category ← ServiceFamily | MeterCategory
  - sku_id ← MeterId | SkuId
  - sku_price_id ← PartNumber | ProductOrderNumber

- Usage/Pricing
  - usage_quantity ← Quantity
  - usage_unit ← UnitOfMeasure | MeterUnit
  - list_unit_price ← RetailPrice | UnitPrice
  - list_cost ← RetailPrice × Quantity
  - pricing_category ← Standard | Spot | Reserved (from PricingModel/PricingModelName, or Benefits)
  - charge_class ← On-Demand | Commitment (Reserved when Benefits apply)

- Resource
  - resource_id ← ResourceId
  - resource_name ← ResourceName | InstanceName
  - resource_type ← ResourceType | ServiceTier
  - region ← normalize(Location | ResourceLocation) → lowercase (e.g., "eastus", "westus2")
  - tags ← parse JSON from Tags column

- Costs
  - effective_cost ← AmortizedCost | CostInBillingCurrency | Cost | CostInUSD | any *cost* column (first non-zero)
  - billed_cost ← CostInBillingCurrency | Cost (optional)

Notes
- provider_name = Microsoft Azure; publisher_name = Microsoft.
- charge_frequency = Daily; charge_period_* follows billing period derived above.

## Classification Rules

- ChargeCategory
  - Explicit from ChargeType/BillingType: Usage, Purchase, Credit, Tax, Adjustment
  - Otherwise infer: any negative effective/billed cost → Credit; default → Usage
- Benefits (Reservations/Savings Plans)
  - Detect via BenefitType/BenefitId/BenefitName or Reservation*/SavingsPlan* columns
  - Populate CommitmentDiscountType = Reservation|SavingsPlan
  - Populate CommitmentDiscountId/Name when present
  - When Benefits apply, set pricing_category=Reserved and charge_class=Commitment

## Normalization
- Region: lowercase of Location/ResourceLocation (input may be "East US" → "east us"; converter outputs lowercase without punctuation)
- Currency: always uppercased (usd → USD, eur → EUR)
- Tags: parsed from JSON payload in Tags column to FOCUS tags map

## Impact on Code
- internal/core/focus/conversion/azure/: mapping and pipeline (`row_mapper.goprocess_csv.goprocess_json.goreader.goheaders.go`).

```bash
,
```
- Note: root forwarders were removed 2025‑08‑19; use direct imports of the `conversion/azure` package.
- Unified adapter: enums/validators/time formats (opt‑in parity path)

## Tests and Fixtures
- Fixtures in tests/fixtures/azure/
  - usage.csv — standard usage row, checks currency/region normalization
  - reservation_credit.csv — negative cost with Reservation benefit; checks CommitmentDiscount*, ChargeCategory=Credit, Reserved pricing
  - tax_refund.csv — CostInBillingCurrency negative with ChargeType=Tax; classified as Tax with normalization
- Parity tests: azure_converter_unified_parity_test.go exercises fast-path vs unified adapter on all fixtures

## Notes on MCA/EA/CSP
- Columns vary slightly by agreement type. Fallback chains above (e.g., BillingAccountId/EnrollmentNumber/BillingProfileId; UsageStart/UsageEnd vs Date/UsageDate) ensure compatibility.
- PriceSheet and Meter metadata can be joined offline; current converter relies on columns present in Usage Details export.

## References
- Exports overview: https://learn.microsoft.com/azure/cost-management-billing/costs/export-classic
- Usage details (EA/MCA/CSP) overview: https://learn.microsoft.com/azure/cost-management-billing/automate/usage-details
- Cost Management exports (create/manage): https://learn.microsoft.com/azure/cost-management-billing/costs/export-templates

MCP: These rules align with the unified mapper adapter to keep outputs identical to the optimized path.
