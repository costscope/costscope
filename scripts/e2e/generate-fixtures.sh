#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -d "$SCRIPT_DIR/../ci/lib" ]]; then SCRIPTS_DIR="$SCRIPT_DIR/.."; else SCRIPTS_DIR="$SCRIPT_DIR"; fi
# shellcheck source=../ci/lib/common.sh
source "$SCRIPTS_DIR/ci/lib/common.sh"
WORKDIR="${ROOT_DIR}/tmp/e2e"
DATA_DIR="${WORKDIR}/data"
mkdir -p "${DATA_DIR}"
: "${E2E_ROWS:=1000}"
ci::log "[fixtures] Generating synthetic provider CSVs in ${DATA_DIR} rows=${E2E_ROWS}"

aws_file="${DATA_DIR}/aws-cur.csv"
printf 'bill/BillingAccountId,bill/BillingAccountName,bill/BillingCurrency,lineItem/UsageAccountId,lineItem/UnblendedCost,lineItem/UsageStartDate,lineItem/UsageEndDate,product/ProductName,lineItem/UsageAmount,lineItem/UsageType,lineItem/LineItemDescription,lineItem/LineItemType,pricing/PriceId,lineItem/ResourceId\n' > "${aws_file}"
awk -v rows="${E2E_ROWS}" 'BEGIN{srand(); for(i=1;i<=rows;i++){cost=rand()*0.05;usage=rand()*10;printf("ba-123,MainAcct,USD,123456789012,%.4f,2025-08-01 00:00:00,2025-08-01 01:00:00,AmazonEC2,%.4f,Hours,RunInstances usage,Usage,pri-%d,i-%d\n",cost,usage,i,i)}}' >> "${aws_file}"

azure_file="${DATA_DIR}/azure-cost.csv"
printf 'BillingAccountId,BillingAccountName,BillingCurrency,SubscriptionId,SubscriptionName,ServiceName,ServiceFamily,ResourceId,Quantity,UnitOfMeasure,AmortizedCost,RetailPrice,UsageStart,UsageEnd,Tags\n' > "${azure_file}"
awk -v rows="${E2E_ROWS}" 'BEGIN{srand(); for(i=1;i<=rows;i++){qty=rand()*5;cost=rand()*0.03;printf("ba-az-1,AzureMain,USD,sub-abc,ProdSub,Virtual Machines,Compute,/subscriptions/sub-abc/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm%d,%.3f,Hours,%.4f,%.4f,2025-08-01T00:00:00Z,2025-08-01T01:00:00Z,env=prod;team=finops\n",i,qty,cost,cost)}}' >> "${azure_file}"

gcp_file="${DATA_DIR}/gcp-billing.csv"
printf 'billing_account_id,billing_account_name,currency,project.id,project.name,service.description,service.id,sku.id,sku.description,usage_start_time,usage_end_time,usage.amount,usage.unit,cost,labels\n' > "${gcp_file}"
awk -v rows="${E2E_ROWS}" 'BEGIN{srand(); for(i=1;i<=rows;i++){cost=rand()*0.04;usage=rand()*8;printf("ba-gcp-1,GCPMain,USD,proj-1,Project One,Compute Engine,svc-1,sku-%d,Standard VM,2025-08-01T00:00:00Z,2025-08-01T01:00:00Z,%.4f,hour,%.5f,env=prod|team=finops\n",i,usage,cost,i)}}' >> "${gcp_file}"

ci::log "[fixtures] Generated: ${aws_file} ${azure_file} ${gcp_file}"
