package main

// Synthetic AWS CUR-like dataset generator for perf benchmarks.
// Generates deterministic rows with varying costs, usage, AZ, product.

import (
	"compress/gzip"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

const header = "bill/BillingAccountId,bill/BillingAccountName,bill/BillingCurrency,lineItem/UnblendedCost,lineItem/UsageAmount,lineItem/UsageStartDate,lineItem/UsageEndDate,lineItem/LineItemDescription,lineItem/Operation,lineItem/ResourceId,lineItem/UsageAccountId,lineItem/UsageType,lineItem/AvailabilityZone,product/ProductName,product/ProductFamily,product/Region,pricing/PriceId\n"

func main() {
	var (
		rows = flag.Int("rows", 20000, "Number of synthetic rows")
		out  = flag.String("out", "tests/perf/aws-cur-synth.csv.gz", "Output .csv or .csv.gz path")
	)
	flag.Parse()

	if err := os.MkdirAll("tests/perf", 0o750); err != nil { // restrictive perms (gosec G306)
		panic(err)
	}

	// Determine gzip
	useGz := strings.HasSuffix(*out, ".gz")
	// #nosec G304 path is user supplied for synthetic test data generation
	f, err := os.Create(*out)
	if err != nil {
		panic(err)
	}
	defer func() { _ = f.Close() }()

	var w *gzip.Writer
	var writeFn func(string)
	if useGz {
		w = gzip.NewWriter(f)
		defer func() { _ = w.Close() }()
		writeFn = func(s string) { _, _ = w.Write([]byte(s)) }
	} else {
		writeFn = func(s string) { _, _ = f.Write([]byte(s)) }
	}

	writeFn(header)
	// RNG seed retained for potential future randomness use; currently rows are deterministic without RNG.
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < *rows; i++ {
		tsStart := start.Add(time.Duration(i) * time.Minute)
		tsEnd := tsStart.Add(time.Minute)
		az := fmt.Sprintf("us-east-1%c", 'a'+rune(i%3))
		svcIdx := i % 5
		var svc, family, usageType, op, desc string
		switch svcIdx {
		case 0:
			svc = "Amazon Elastic Compute Cloud"
			family = "Compute Instance"
			usageType = "BoxUsage:t3.micro"
			op = "RunInstances"
			desc = "EC2 Instance Hours"
		case 1:
			svc = "Amazon Simple Storage Service"
			family = "Storage"
			usageType = "TimedStorage-ByteHrs"
			op = "PutObject"
			desc = "S3 Storage"
		case 2:
			svc = "Amazon Relational Database Service"
			family = "Database Instance"
			usageType = "InstanceUsage:db.t3.micro"
			op = "CreateDBInstance"
			desc = "RDS Database Hours"
		case 3:
			svc = "AWS Lambda"
			family = "Serverless Computing"
			usageType = "Request-x86"
			op = "Invoke"
			desc = "Lambda Function Invocations"
		default:
			svc = "Amazon CloudWatch"
			family = "Management Tools"
			usageType = "MetricMonitorUsage"
			op = "PutMetricData"
			desc = "CloudWatch Metrics"
		}
		cost := float64((i%100)+1) * 0.07
		usage := float64((i%500)+1) * 1.3
		resourceId := fmt.Sprintf("res-%06d", i)
		priceId := fmt.Sprintf("P%05d", i%1000)
		line := fmt.Sprintf("123456789012,PerfSynthetic,USD,%.2f,%.2f,%s,%s,%s,%s,%s,123456789012,%s,%s,%s,%s,us-east-1,%s\n",
			cost, usage,
			tsStart.Format("2006-01-02 15:04:05"),
			tsEnd.Format("2006-01-02 15:04:05"),
			desc, op, resourceId, usageType, az, svc, family, priceId)
		writeFn(line)
	}
	fmt.Printf("Synthetic dataset generated: %s (%d rows)\n", *out, *rows)
}
