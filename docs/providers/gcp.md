# GCP Billing → FOCUS Mapping

## Scope
- BigQuery Billing Export schema: invoice, cost, credits, sku, location, labels, ancestry
- Credit types: CUD, Sustained Use, Spot/Preemptible, Promotional, Free tier
- Time fields, timezones, partitioning
- Optional: Cloud Catalog/Price API for enrichment

## Field Mapping (BQ Export → FOCUS)
- billing_account_id/name → billing_account_id/name
- currency → billing_currency (uppercased ISO, unified path normalization)
- usage_start_time/usage_end_time → billing_period_start/end and charge_period_start/end
- provider_name → Google Cloud Platform; publisher_name → Google
- service.description → service_name; service.category → service_category
- sku.id → sku_id; pricing.price_id → sku_price_id
- cost → effective_cost; usage.amount → usage_quantity; usage.unit → usage_unit/pricing_unit
- location.region → region (lowercased); location.zone → availability_zone (lowercased)
- labels/system_labels/resource.labels → tags (object or array formats supported)
- credits[] → ChargeCategory=Credit when present or when cost < 0; CommitmentDiscount* populated when metadata exists

## Classification Rules
- ChargeCategory: Usage by default; Credit when cost < 0 or credits[] present
- Credits[] classification and enrichment:
  - Committed Use Discount (CUD): when credits[].type indicates commitment; populate CommitmentDiscountId/Type/Name from credits metadata
  - Sustained Use: classified as Credit
  - Spot/Preemptible: classified as Credit, and unified path sets pricing_category=Spot
  - Promotional/Free tier: classified as Credit
  - Unknown credit types: remain Credit if cost < 0 or credits present
  - Note: Detailed classification is based on credits[].type/name; see BigQuery export schema

## Normalization
- Currency: uppercased (e.g., usd → USD) in unified path
- Region/Zone: lowercased for consistency (e.g., us-central1-a)
- Labels: keys lowercased and spaces to underscores

## Impact on Code
- internal/core/focus/conversion/gcp/: provider-scoped mapper and pipeline (e.g., `process_csv.goprocess_json.gomapping_csv.gomappers.gohelpers.go`).

```bash
,
```
- Note: root forwarders/wrappers were removed 2025‑08‑19; use direct imports of the `conversion/gcp` package.
- Unified adapter: tolerant time parsing relies on shared helpers; remains opt‑in for parity and diagnostics

## Tests and Fixtures
- Added fixtures in tests/fixtures/gcp:
  - usage_minimal.csv
  - credit_cud.csv (with id/type/name metadata)
  - credit_spot.csv (type/name contains Spot)
  - credit_sustained_promo.csv (sustained + promotional)
- Tests:
  - gcp_converter_unified_parity_test.go: parity core fields; credit classification assertions

## Open Questions
- Any new credit types or schema changes in BQ export
- Partitioning/time format nuances

## References
- BigQuery Billing export tables overview and schema: https://cloud.google.com/billing/docs/how-to/export-data-bigquery-tables#schema
- Standard usage table: https://cloud.google.com/billing/docs/how-to/export-data-bigquery-tables/standard-usage
- Detailed usage table: https://cloud.google.com/billing/docs/how-to/export-data-bigquery-tables/detailed-usage
- Pricing export: https://cloud.google.com/billing/docs/how-to/export-data-bigquery-tables/pricing-data
- CUD metadata export: https://cloud.google.com/billing/docs/how-to/export-data-bigquery-tables/cud-export

Notes:
- Partitioning/timezones: timestamps are parsed as RFC3339 when available; internal Parquet encoding uses UTC epoch millis for FOCUS fields.
