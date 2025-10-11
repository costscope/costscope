---
title: Minimal IAM Policies
description: Least privilege IAM examples for artifact export to S3/GCS.
---
## Objective
Provide baseline least‑privilege policies for reading billing exports and writing optional report artifacts.

## AWS (CUR in S3)
```json
{
	"Version": "2012-10-17",
	"Statement": [
		{ "Effect": "Allow", "Action": ["s3:GetObject", "s3:ListBucket"], "Resource": ["arn:aws:s3:::my-cur-bucket", "arn:aws:s3:::my-cur-bucket/*"] },
		{ "Effect": "Allow", "Action": ["athena:GetDataCatalog", "athena:StartQueryExecution", "athena:GetQueryExecution", "athena:GetQueryResults"], "Resource": "*" }
	]
}
```

## Azure (Cost Management Export in Storage Account)
Assign role  to the principal scoped to the export container; no write required.

```bash
Storage Blob Data Reader
```

## GCP (Billing Export in BigQuery)
Grant roles: `roles/bigquery.dataViewerroles/bigquery.jobUser` if running custom queries.

```bash
on dataset; optional
```

## Principles
- Read-only where possible
- Explicit bucket/dataset scoping (avoid wildcards)
- Separate principals for ingestion vs CI smoke tests

## Rotation / Key Management
Prefer short-lived credentials (STS, workload identity). If long-lived access keys exist, rotate quarterly and upon personnel changes.

Update if new provider integrations or write paths added.
