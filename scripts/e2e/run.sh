#!/usr/bin/env bash
## Temporarily disable -e (and pipefail) for deep debug; keep nounset for safety.
# Original: set -euo pipefail
set -u
set -o functrace
set -E
[[ -n "${E2E_DEBUG:-}" ]] && set -x
trap 'echo "[E2E-ERR] line:$LINENO status:$? cmd:$BASH_COMMAND" >&2' ERR
# E2E ingestion -> conversion -> validation -> analytics -> export scenario
# Generates synthetic provider billing data (10-50K rows) for AWS, Azure, GCP and exercises the CLI.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BIN="${ROOT_DIR}/costscope"
WORKDIR="${ROOT_DIR}/tmp/e2e"
DATA_DIR="${WORKDIR}/data"
OUT_DIR="${WORKDIR}/out"
REPORT_DIR="${WORKDIR}/reports"
JUNIT_FILE="${WORKDIR}/junit-e2e.xml"
mkdir -p "$DATA_DIR" "$OUT_DIR" "$REPORT_DIR"
# Prune old outputs to avoid noise between runs
rm -f "$OUT_DIR"/* || true
rm -f "$REPORT_DIR"/* || true

: "${E2E_ROWS:=10000}"  # default rows per provider within 10-50K target band (optimized default)
: "${E2E_TIMEOUT_VALIDATE:=40}"   # seconds
: "${E2E_TIMEOUT_ANALYTICS:=60}"  # seconds
: "${E2E_TIMEOUT_REPORT:=45}"     # seconds

if [[ ! -x "$BIN" ]]; then
  echo "Binary not found at $BIN. Build first (make build)." >&2
  exit 2
fi

log() { echo "[$(date -u +%Y-%m-%dT%H:%M:%SZ)] $*"; }

failures=0
skips=0
cases=()
case_results=()

record_case() {
  local name=$1 status=$2 duration=$3 message=$4
  cases+=("$name")
  case_results+=("$status|$duration|$message")
  if [ "$status" = "failed" ]; then
    failures=$((failures+1))
  elif [ "$status" = "skipped" ]; then
    skips=$((skips+1))
  fi
  return 0
}

has_file() { [[ -f "$1" ]]; }

# Heartbeat helpers
start_heartbeat() {
  local tag=$1
  (
    while true; do
      echo "[E2E][hb][$tag] $(date -u +%H:%M:%S)" >&2
      sleep 5
    done
  ) &
  echo $!
}
stop_heartbeat() { kill "$1" 2>/dev/null || true; }

# Track conversion success flags per provider (portable without associative arrays)
CONVERT_OK_aws=0
CONVERT_OK_azure=0
CONVERT_OK_gcp=0

# Generate synthetic AWS CUR CSV with required minimal columns + a few optional
aws_file="$DATA_DIR/aws-cur.csv"
log "Generating synthetic AWS CUR: $aws_file rows=$E2E_ROWS"
printf 'bill/BillingAccountId,bill/BillingAccountName,bill/BillingCurrency,lineItem/UsageAccountId,lineItem/UnblendedCost,lineItem/UsageStartDate,lineItem/UsageEndDate,product/ProductName,lineItem/UsageAmount,lineItem/UsageType,lineItem/LineItemDescription,lineItem/LineItemType,pricing/PriceId,lineItem/ResourceId\n' > "$aws_file"
awk -v rows="$E2E_ROWS" 'BEGIN{for(i=1;i<=rows;i++){cost=rand()*0.05;usage=rand()*10;printf("ba-123,MainAcct,USD,123456789012,%.4f,2025-08-01 00:00:00,2025-08-01 01:00:00,AmazonEC2,%.4f,Hours,RunInstances usage,Usage,pri-%d,i-%d\n",cost,usage,i,i)}}' >> "$aws_file"

# Minimal Azure synthetic (headers approximated for mapper fields used). We'll supply generic usage/cost columns used by mapping rules.
azure_file="$DATA_DIR/azure-cost.csv"
log "Generating synthetic Azure export: $azure_file rows=$E2E_ROWS"
printf 'BillingAccountId,BillingAccountName,BillingCurrency,SubscriptionId,SubscriptionName,ServiceName,ServiceFamily,ResourceId,Quantity,UnitOfMeasure,AmortizedCost,RetailPrice,UsageStart,UsageEnd,Tags\n' > "$azure_file"
awk -v rows="$E2E_ROWS" 'BEGIN{for(i=1;i<=rows;i++){qty=rand()*5;cost=rand()*0.03;printf("ba-az-1,AzureMain,USD,sub-abc,ProdSub,Virtual Machines,Compute,/subscriptions/sub-abc/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm%d,%.3f,Hours,%.4f,%.4f,2025-08-01T00:00:00Z,2025-08-01T01:00:00Z,env=prod;team=finops\n",i,qty,cost,cost)}}' >> "$azure_file"

# Minimal GCP synthetic export
gcp_file="$DATA_DIR/gcp-billing.csv"
log "Generating synthetic GCP export: $gcp_file rows=$E2E_ROWS"
printf 'billing_account_id,billing_account_name,currency,project.id,project.name,service.description,service.id,sku.id,sku.description,usage_start_time,usage_end_time,usage.amount,usage.unit,cost,labels\n' > "$gcp_file"
awk -v rows="$E2E_ROWS" 'BEGIN{for(i=1;i<=rows;i++){cost=rand()*0.04;usage=rand()*8;printf("ba-gcp-1,GCPMain,USD,proj-1,Project One,Compute Engine,svc-1,sku-%d,Standard VM,2025-08-01T00:00:00Z,2025-08-01T01:00:00Z,%.4f,hour,%.5f,env=prod|team=finops\n",i,usage,cost,i)}}' >> "$gcp_file"

convert_provider() {
  local provider=$1 input=$2
  local name="convert-$provider"
  local output="$OUT_DIR/${provider}-focus.parquet"
  local start=$(perl -MTime::HiRes=time -e 'printf("%d", time()*1000)')
  if [[ -n "${E2E_DEBUG:-}" ]]; then
    ( $BIN convert --provider "$provider" --input "$input" --output "$output" --streaming --chunk-size 4096 ) &
  else
    ( $BIN convert --provider "$provider" --input "$input" --output "$output" --streaming --chunk-size 4096 >/dev/null 2>&1 ) &
  fi
  local conv_pid=$!
  echo "[E2E] waiting for convert pid=$conv_pid provider=$provider"
  wait $conv_pid
  rc=$?
  echo "[E2E] post-rc provider=$provider rc=$rc"
  if [ $rc -eq 0 ]; then
    local end=$(perl -MTime::HiRes=time -e 'printf("%d", time()*1000)')
    record_case "$name" passed "$((end-start))" ""
    eval CONVERT_OK_${provider}=1
  else
    local end=$(perl -MTime::HiRes=time -e 'printf("%d", time()*1000)')
    echo "[DEBUG] convert $provider failed rc=$rc" >&2
    record_case "$name" failed "$((end-start))" "conversion failed rc=$rc"
    eval CONVERT_OK_${provider}=0
  fi
  echo "[E2E] starting JSON variant for $provider"
  # JSON export variant (moved ahead of rotation debug)
  local name_json="convert-${provider}-json"
  local output_json="$OUT_DIR/${provider}-focus.json"
  local start_json=$(perl -MTime::HiRes=time -e 'printf("%d", time()*1000)')
  if [[ $(eval echo \$CONVERT_OK_${provider}) -eq 1 ]]; then
    if [[ -n "${E2E_DEBUG:-}" ]]; then
      ( $BIN convert --provider "$provider" --input "$input" --output "$output_json" --streaming --chunk-size 4096 ) &
    else
      ( $BIN convert --provider "$provider" --input "$input" --output "$output_json" --streaming --chunk-size 4096 >/dev/null 2>&1 ) &
    fi
    local json_pid=$!
    echo "[E2E] waiting for json convert pid=$json_pid provider=$provider"
    wait $json_pid
    rcj=$?
    echo "[E2E] post-rc-json provider=$provider rc=$rcj"
    if [ $rcj -eq 0 ]; then
      local end_json=$(perl -MTime::HiRes=time -e 'printf("%d", time()*1000)')
      record_case "$name_json" passed "$((end_json-start_json))" ""
    else
      local end_json=$(perl -MTime::HiRes=time -e 'printf("%d", time()*1000)')
      echo "[DEBUG] convert $provider (json) failed rc=$rcj" >&2
      record_case "$name_json" failed "$((end_json-start_json))" "conversion(json) failed rc=$rcj"
    fi
  else
    local end_json=$(perl -MTime::HiRes=time -e 'printf("%d", time()*1000)')
    record_case "$name_json" skipped "$((end_json-start_json))" "skipped due to primary conversion failure"
  fi
  # Normalize rotated parquet filename (if rotation produced timestamped files)
  if [[ ! -f "$output" ]]; then
    # Find newest rotated file matching pattern provider-focus-*.parquet and copy to canonical path
    local latest
    latest=$(ls -1t "$OUT_DIR/${provider}-focus-"*.parquet 2>/dev/null | head -1 || true)
    if [[ -n "$latest" && -f "$latest" ]]; then
      [[ -n "${E2E_DEBUG:-}" ]] && echo "[E2E] normalizing rotated file $latest -> $output"
      cp -f "$latest" "$output" 2>/dev/null || true
    fi
  fi
  [[ -n "${E2E_DEBUG:-}" ]] && ls -1 "$OUT_DIR" 2>/dev/null | sed 's/^/[E2E] outdir: /'
  echo "[E2E] after rotation copy for $provider"
  echo "[E2E] convert_provider completed for $provider (rc=$rc)"
  return 0
}

validate_output() {
  local provider=$1
  local input="$OUT_DIR/${provider}-focus.parquet"
  local name="validate-$provider"
  local start=$(perl -MTime::HiRes=time -e 'printf("%d", time()*1000)')
  echo "[E2E] START validate $provider"
  if [[ $(eval echo \$CONVERT_OK_${provider}) -ne 1 || ! -f "$input" ]]; then
    local end=$(perl -MTime::HiRes=time -e 'printf("%d", time()*1000)'); record_case "$name" skipped "$(($end-start))" "skipped (no conversion output)"; echo "[E2E] END   validate $provider skipped"; return
  fi
  local hb=$(start_heartbeat "validate-$provider")
  # Timeout enforced via perl alarm (no watchdog pkill to avoid killing unrelated processes)
  perl -e "alarm $E2E_TIMEOUT_VALIDATE; exec @ARGV" env COSTSCOPE_E2E_MODE=1 \
    "$BIN" validate "$input" --schema --quality --quiet \
    >"$OUT_DIR/${provider}-validate.out" 2>"$OUT_DIR/${provider}-validate.err" &
  local vpid=$!
  wait $vpid; local rc=$?
  local end=$(perl -MTime::HiRes=time -e 'printf("%d", time()*1000)')
  stop_heartbeat "$hb"
  if [ $rc -eq 0 ]; then
    record_case "$name" passed "$((end-start))" "no anomalies"
    echo "[E2E] END   validate $provider rc=0"
  else
    echo "[DEBUG] validate $provider failed rc=$rc" >&2
    tail -n 30 "$OUT_DIR/${provider}-validate.err" >&2 || true
    record_case "$name" failed "$((end-start))" "validation failed rc=$rc"
    echo "[E2E] END   validate $provider rc=$rc"
  fi
}

analytics_basic() {
  local provider=$1
  local input="$OUT_DIR/${provider}-focus.parquet"
  local output_json="$OUT_DIR/${provider}-analytics.json"
  local name="analytics-$provider"
  local start=$(perl -MTime::HiRes=time -e 'printf("%d", time()*1000)')
  if [[ $(eval echo \$CONVERT_OK_${provider}) -ne 1 || ! -f "$input" ]]; then
    local end=$(perl -MTime::HiRes=time -e 'printf("%d", time()*1000)'); record_case "$name" skipped "$(($end-start))" "skipped (no conversion output)"; return
  fi
  echo "[E2E] START analytics $provider"
  local hb=$(start_heartbeat "analytics-$provider")
  perl -e "alarm $E2E_TIMEOUT_ANALYTICS; exec @ARGV" \
    "$BIN" analyze-enhanced "$input" --output json --memory-limit 512MB \
    >"$output_json" 2>"$output_json.stderr" &
  local apid=$!
  wait $apid; local rc=$?
  local end=$(perl -MTime::HiRes=time -e 'printf("%d", time()*1000)')
  stop_heartbeat "$hb"
  if [ $rc -eq 0 ]; then
    record_case "$name" passed "$((end-start))" ""
    echo "[E2E] END   analytics $provider rc=0"
  else
    echo "[DEBUG] analytics $provider failed rc=$rc" >&2
    tail -n 40 "$output_json.stderr" >&2 || true
    record_case "$name" failed "$((end-start))" "analytics failed rc=$rc"
    echo "[E2E] END   analytics $provider rc=$rc"
  fi
}

export_report() {
  local provider=$1
  local input="$OUT_DIR/${provider}-focus.parquet"
  local name="report-$provider"
  local output_report="$REPORT_DIR/${provider}-exec-summary.json"
  local start=$(perl -MTime::HiRes=time -e 'printf("%d", time()*1000)')
  if [[ $(eval echo \$CONVERT_OK_${provider}) -ne 1 || ! -f "$input" ]]; then
    local end=$(perl -MTime::HiRes=time -e 'printf("%d", time()*1000)'); record_case "$name" skipped "$(($end-start))" "skipped (no conversion output)"; return
  fi
  echo "[E2E] START report $provider"
  local hb=$(start_heartbeat "report-$provider")
  if perl -e "alarm $E2E_TIMEOUT_REPORT; exec @ARGV" "$BIN" reports generate --type executive-summary --input "$input" --format json --output "$output_report" >"$REPORT_DIR/${provider}-report.out" 2>"$REPORT_DIR/${provider}-report.err"; then
    local end=$(perl -MTime::HiRes=time -e 'printf("%d", time()*1000)'); stop_heartbeat "$hb"; record_case "$name" passed "$(($end-start))" ""; echo "[E2E] END   report $provider rc=0"
  else
    local rc=$?; local end=$(perl -MTime::HiRes=time -e 'printf("%d", time()*1000)'); stop_heartbeat "$hb"; echo "[DEBUG] report $provider failed rc=$rc" >&2; tail -n 30 "$REPORT_DIR/${provider}-report.err" >&2 || true; record_case "$name" failed "$(($end-start))" "report failed rc=$rc"; echo "[E2E] END   report $provider rc=$rc"
  fi
}

# Execute for each provider with post-call echo
convert_provider aws "$aws_file"; echo "[E2E] returned from convert_provider aws"
convert_provider azure "$azure_file"; echo "[E2E] returned from convert_provider azure"
convert_provider gcp "$gcp_file"; echo "[E2E] returned from convert_provider gcp"

validate_output aws
validate_output azure
validate_output gcp

analytics_basic aws
analytics_basic azure
analytics_basic gcp

export_report aws
export_report azure
export_report gcp

# Simple export scenario (reuse analytics output for packaging) - placeholder step
# In a full implementation this could call a reports generate or export command.

# Produce JUnit XML
log "Writing JUnit report: $JUNIT_FILE"
{
  echo "<?xml version=\"1.0\" encoding=\"UTF-8\"?>"
  echo "<testsuite name=\"costscope-e2e\" tests=\"${#cases[@]}\" failures=\"$failures\" skipped=\"$skips\" time=\"0\">"
  for i in "${!cases[@]}"; do
    IFS='|' read -r status duration message <<<"${case_results[$i]}"
    name="${cases[$i]}"
    seconds=$(awk -v ms=$duration 'BEGIN{printf "%.3f", ms/1000}')
    case $status in
      failed)
        echo "  <testcase name=\"$name\" time=\"$seconds\"><failure message=\"$message\"/></testcase>";;
      skipped)
        echo "  <testcase name=\"$name\" time=\"$seconds\"><skipped message=\"$message\"/></testcase>";;
      passed)
        echo "  <testcase name=\"$name\" time=\"$seconds\"/>";;
    esac
  done
  echo "</testsuite>"
} > "$JUNIT_FILE"

log "E2E completed: failures=$failures"
exit $failures
