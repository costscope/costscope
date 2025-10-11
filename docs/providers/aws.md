# AWS Billing → FOCUS Mapping (AWS CUR v1.0+ → FOCUS v1.2)

This note documents the current AWS CUR mapping in CostScope, required classifications (SP/RI/credits/tax/refund), normalization rules, manifest support, and minimal fixtures for parity tests with the unified adapter.

## Scope
- AWS Cost and Usage Reports (CUR): bill/*, lineItem/*, pricing/*, product/*
- Savings Plans (SP), Reserved Instances (RI), Private Pricing/EDP, Taxes, Refunds, Credits
- Manifest partitioning and compression compatibility (.csv, .csv.gz; manifest.json, manifest.json.gz)

## CUR → FOCUS core mapping (implemented)

Converter modules: optimized path under `internal/core/focus/conversion/aws/converter.gomapper.goreader.goheaders.goconversion/aws`. Unified adapter remains optional for parity.

```bash
(e.g.,
```

```bash
,
```

```bash
). Root forwarders were removed 2025‑08‑19; use direct package imports
```

- bill/BillingAccountId → billing_account_id
- bill/BillingAccountName → billing_account_name
- bill/BillingCurrency → billing_currency
- lineItem/UsageStartDate → billing_period_start, charge_period_start
- lineItem/UsageEndDate → billing_period_end, charge_period_end
- lineItem/LineItemDescription → charge_description
- lineItem/Operation → charge_subcategory
- lineItem/UnblendedCost → effective_cost, list_cost
- lineItem/UsageAmount → pricing_quantity, usage_quantity
- lineItem/UsageType → pricing_unit, usage_unit
- product/ProductFamily → service_category
- product/ProductName → service_name, resource_type
- pricing/PriceId → sku_price_id (when available)
- lineItem/LineItemType → sku_id (used also for classification)
- lineItem/ResourceId → resource_id, resource_name
- product/Region → region (optional)
- provider_name/publisher_name → “Amazon Web Services”

Additional commitment fields (optional when present)
- savingsPlan/SavingsPlanId or savingsPlan/SavingsPlanArn → commitment_discount_id/name; commitment_discount_type="SavingsPlan"
- reservation/SubscriptionId or reservation/ReservationARN → commitment_discount_id/name; commitment_discount_type="ReservedInstance"

Notes
- List unit price: computed as UnblendedCost / UsageAmount when UsageAmount > 0, else 0.
- Time format: "YYYY-MM-DD HH:MM:SS" (UTC by convention in most CUR exports). Internally persisted as TIMESTAMP_MILLIS in Parquet.
- Defaults (in unified default handler): ChargeCategory=Usage, ChargeClass=On-Demand, PricingCategory=Standard. These are overridden by classification rules below when implemented.

## Classification rules (converter + unified adapter)

CUR discriminator: `lineItem/LineItemType` and related namespaces (savingsPlan/*, reservation/*), plus cost sign for credits.

- On-Demand usage
  - When LineItemType ∈ {"Usage"} or none of SP/RI markers apply
  - ChargeCategory=Usage; ChargeClass=On-Demand; PricingCategory=Standard

- Spot usage
  - When UsageType/Operation indicates Spot (e.g., "SpotUsage" variants)
  - ChargeCategory=Usage; ChargeClass=On-Demand; PricingCategory=Spot

- RI covered usage and fees
  - Covered usage: LineItemType ∈ {"DiscountedUsage"}
    - ChargeCategory=Usage; ChargeClass=Commitment; PricingCategory=Reserved
    - CommitmentDiscountType="ReservedInstance"; Id/Name from reservation/* when available
  - RI recurring fee: LineItemType ∈ {"RIFee"}
    - ChargeCategory=Purchase; ChargeClass=Commitment; CommitmentDiscountType="ReservedInstance"

- Savings Plans covered usage and fees
  - Covered usage: LineItemType ∈ {"SavingsPlanCoveredUsage"}
    - ChargeCategory=Usage; ChargeClass=Commitment; PricingCategory=Standard
    - CommitmentDiscountType="SavingsPlan"; Id/Name from savingsPlan/* when available
  - SP recurring fee: LineItemType ∈ {"SavingsPlanRecurringFee"}
    - ChargeCategory=Purchase; ChargeClass=Commitment; CommitmentDiscountType="SavingsPlan"
  - Negation/Adjustment lines: LineItemType ∈ {"SavingsPlanNegation"}
    - ChargeCategory=Adjustment; ChargeClass=Commitment; CommitmentDiscountType="SavingsPlan"

- Credits/Refunds/Taxes/Fees
  - Credits: LineItemType ∈ {"Credit"} or negative cost designated as credit → ChargeCategory=Credit
    - When tied to SP/RI (e.g., negations), also set CommitmentDiscountType accordingly
  - Refunds: LineItemType ∈ {"Refund"} → ChargeCategory=Adjustment (FOCUS doesn’t have a distinct Refund category)
  - Taxes: LineItemType ∈ {"Tax"} → ChargeCategory=Tax
  - Support/fee lines (e.g., "Fee") → ChargeCategory=Adjustment

Commitment metadata
- CommitmentDiscountId/Name/Type are optional fields; populate when savingsPlan/* or reservation/* columns exist. Canonical values for Type: "SavingsPlan", "ReservedInstance", "PrivatePricing", "EDP" (proposal; see Enums section).

## Normalization rules (soft; defaults unchanged)

- Region
  - Current: pass-through `product/RegionlineItem/AvailabilityZone`.

```bash
and optional
```
  - Proposal: normalize common aliases (e.g., "EU (Ireland)" → "eu-west-1"). Keep original in tags if available.

- Currency
  - Use `bill/BillingCurrency` as-is. FOCUS requires consistent currency across record; multi-currency datasets should be partitioned.

- Service/Operation
  - ServiceName from `product/ProductNameproduct/ProductFamily`.

```bash
; ServiceCategory from
```
  - Operation from `lineItem/Operation` (pass-through). Consider mapping well-known operations to canonical forms if needed.

- PricingCategory and ChargeClass
  - See classification rules. Defaults remain Standard/On-Demand for non-commitment, Spot for spot usage, Reserved for RI covered usage.

## Manifest, partitioning, compression

- Supported inputs: .csv, .csv.gz, manifest.json, manifest.json.gz (both confirmed by tests)
- `CURManifestreportKeys` are resolved relative to the manifest directory and processed sequentially.

```bash
structure parsed from JSON;
```
- For manifest inputs, output files are suffixed per report file while preserving the requested output extension.

## Enrichment opportunities (outside CUR)

- AWS Price List API (SP/RI rate context, unit/term clarifications); fill `sku_price_id` when missing and resolve list price variations.
- Cost Explorer (effective savings rates, amortized commitment views) for validation and QA-only comparisons.
- Private pricing/EDP: record detection via custom tag or known accounts/ARNs; set CommitmentDiscountType where signal exists.

## Enums and constants

Currently in code (`internal/core/focus/types/focus_v1_2.go`):
- ChargeCategory: Usage, Purchase, Tax, Adjustment, Credit
- ChargeClass: On-Demand, Commitment, Correction
- PricingCategory: Standard, Spot, Reserved

Proposed canonical values for CommitmentDiscountType (string):
- "SavingsPlan", "ReservedInstance", "PrivatePricing", "EDP" (no code enum required; use string constants in mapper)

Status: no code enum change required; adopt canonical strings in mapping logic.

## Impact on code (planned deltas)

- aws/converter (optimized path)
  - Extend header index to include: `lineItem/LineItemTypesavingsPlan/*reservation/*` when present.
  - Add classification switch on LineItemType and set ChargeCategory/ChargeClass/PricingCategory and CommitmentDiscount* fields accordingly.
  - Preserve fast path; keep unified adapter parity.

- Unified adapter
  - Add normalization hooks for region and enums; override defaults when LineItemType indicates commitment/spot/credits.
  - Add basic detectors for Spot (UsageType/Operation) and credits/tax/refund via LineItemType.

- Tests
  - Add parity tests for commitment, spot, credit/tax/refund; verify CommitmentDiscount* population when source fields exist.

## Minimal fixtures (added)

Folder: `tests/fixtures/aws/`
- `cur_minimal_usage.csv` – On-Demand usage record
- `cur_savingsplan_covered_usage.csv` – SavingsPlan covered usage record
- `cur_tax_refund.csv` – Five records: tax, refund, credit, spot usage, RI fee

Each fixture includes at least the converter’s required headers: `lineItem/UsageAccountIdlineItem/UnblendedCostlineItem/UsageStartDatelineItem/UsageEndDateproduct/ProductName`. Optional helpful columns are included where relevant.

## Acceptance checklist
- This document updated; fixtures added; proposed enum/normalization rules listed. Further PRs will implement the classification logic and tests described above.

## References
- AWS CUR Data dictionary: https://docs.aws.amazon.com/cur/latest/userguide/data-dictionary.html
- CUR discount details (reservations/discounts/split items): https://docs.aws.amazon.com/cur/latest/userguide/discount-details.html
- Savings Plans details (columns and semantics): https://docs.aws.amazon.com/cur/latest/userguide/savingsplans-columns.html
