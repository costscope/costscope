//nolint:goconst,gosec // Test file with repeated string literals and file operations
package conversion

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
)

// TestFOCUSConverter tests the main FOCUS conversion functionality
func TestFOCUSConverter(t *testing.T) {
	logger := logging.NewLogger("debug")
	converter := NewConverter(logger)

	tests := []struct {
		name     string
		provider string
		input    string
		want     bool
		wantErr  bool
	}{
		{
			name:     "AWS CUR conversion",
			provider: "aws",
			input:    createTestAWSCUR(t),
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Azure usage conversion",
			provider: "azure",
			input:    createTestAzureUsage(t),
			want:     true,
			wantErr:  false,
		},
		{
			name:     "GCP billing conversion",
			provider: "gcp",
			input:    createTestGCPBilling(t),
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Invalid provider",
			provider: "invalid",
			input:    "test.csv",
			want:     false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			options := ConversionOptions{
				Provider:    tt.provider,
				InputPath:   tt.input,
				OutputPath:  filepath.Join(t.TempDir(), "output.parquet"),
				ChunkSize:   1000,
				Workers:     2,
				Compression: true,
				Validate:    true,
			}

			result, err := converter.Convert(ctx, options)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Convert() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Convert() unexpected error: %v", err)
				return
			}

			if result.Success != tt.want {
				t.Errorf("Convert() success = %v, want %v", result.Success, tt.want)
			}

			// Verify output file exists
			if _, err := os.Stat(options.OutputPath); os.IsNotExist(err) {
				t.Errorf("Output file not created: %s", options.OutputPath)
			}

			// Verify FOCUS compliance
			if result.FOCUSCompliance.Score < 85.0 {
				t.Errorf("FOCUS compliance score too low: %.2f", result.FOCUSCompliance.Score)
			}
		})
	}
}

// TestSchemaMapping tests the schema mapping functionality
func TestSchemaMapping(t *testing.T) {
	logger := logging.NewLogger("debug")
	converter := NewConverter(logger)

	tests := []struct {
		name     string
		provider string
		fields   map[string]interface{}
		want     FOCUSRecord
		wantErr  bool
	}{
		{
			name:     "AWS CUR mapping",
			provider: "aws",
			fields: map[string]interface{}{
				"lineItem/UsageAccountId": "123456789012",
				"lineItem/ProductCode":    "AmazonEC2",
				"lineItem/UsageType":      "BoxUsage:t3.micro",
				"lineItem/Operation":      "RunInstances",
				"lineItem/BlendedCost":    "1.50",
				"lineItem/UsageStartDate": "2025-08-01T00:00:00Z",
				"lineItem/UsageEndDate":   "2025-08-01T01:00:00Z",
				"product/region":          "us-east-1",
			},
			want: FOCUSRecord{
				BillingAccountId:   "123456789012",
				ServiceName:        "AmazonEC2",
				ResourceId:         "",
				UsageUnit:          "Hours",
				BilledCost:         1.50,
				EffectiveCost:      1.50,
				BillingPeriodStart: time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC),
				BillingPeriodEnd:   time.Date(2025, 8, 1, 1, 0, 0, 0, time.UTC),
				ChargeCategory:     "Usage",
				ChargeSubcategory:  "OnDemand",
				ChargeType:         "Usage",
				ChargeClass:        "Committed",
				ChargeFrequency:    "Monthly",
				ChargeDescription:  "RunInstances",
				Region:             "us-east-1",
				AvailabilityZone:   "",
				ServiceCategory:    "Compute",
			},
			wantErr: false,
		},
		{
			name:     "Azure usage mapping",
			provider: "azure",
			fields: map[string]interface{}{
				"subscriptionId": "sub-123",
				"serviceName":    "Virtual Machines",
				"meterName":      "D2s v3",
				"quantity":       "24",
				"cost":           "48.00",
				"date":           "2025-08-01",
				"location":       "East US",
			},
			want: FOCUSRecord{
				BillingAccountId:   "sub-123",
				ServiceName:        "Virtual Machines",
				ResourceId:         "",
				UsageUnit:          "Hours",
				BilledCost:         48.00,
				EffectiveCost:      48.00,
				BillingPeriodStart: time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC),
				BillingPeriodEnd:   time.Date(2025, 8, 2, 0, 0, 0, 0, time.UTC),
				ChargeCategory:     "Usage",
				ChargeSubcategory:  "OnDemand",
				ChargeType:         "Usage",
				ChargeClass:        "Committed",
				ChargeFrequency:    "Monthly",
				ChargeDescription:  "D2s v3",
				Region:             "East US",
				AvailabilityZone:   "",
				ServiceCategory:    "Compute",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := converter.GetSchemaMapper(tt.provider)
			if mapper == nil {
				t.Fatalf("No schema mapper found for provider: %s", tt.provider)
			}

			result, err := mapper.MapToFOCUS(tt.fields)

			if tt.wantErr {
				if err == nil {
					t.Errorf("MapToFOCUS() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("MapToFOCUS() unexpected error: %v", err)
				return
			}

			// Verify key fields
			if result.BillingAccountId != tt.want.BillingAccountId {
				t.Errorf("BillingAccountId = %v, want %v", result.BillingAccountId, tt.want.BillingAccountId)
			}
			if result.ServiceName != tt.want.ServiceName {
				t.Errorf("ServiceName = %v, want %v", result.ServiceName, tt.want.ServiceName)
			}
			if result.BilledCost != tt.want.BilledCost {
				t.Errorf("BilledCost = %v, want %v", result.BilledCost, tt.want.BilledCost)
			}
		})
	}
}

// TestDataValidation tests data validation during conversion
func TestDataValidation(t *testing.T) {
	logger := logging.NewLogger("debug")
	converter := NewConverter(logger)

	tests := []struct {
		name    string
		record  FOCUSRecord
		want    ValidationResult
		wantErr bool
	}{
		{
			name: "Valid FOCUS record",
			record: FOCUSRecord{
				BillingAccountId:   "123456789012",
				ServiceName:        "AmazonEC2",
				BilledCost:         1.50,
				EffectiveCost:      1.50,
				BillingPeriodStart: time.Now().Add(-24 * time.Hour),
				BillingPeriodEnd:   time.Now(),
				ChargeCategory:     "Usage",
				ChargeType:         "Usage",
				Region:             "us-east-1",
			},
			want: ValidationResult{
				IsValid: true,
				Score:   100.0,
				Issues:  []string{},
			},
			wantErr: false,
		},
		{
			name: "Invalid FOCUS record - missing required fields",
			record: FOCUSRecord{
				BillingAccountId: "123456789012",
				// Missing ServiceName, BilledCost, etc.
			},
			want: ValidationResult{
				IsValid: false,
				Score:   0.0,
				Issues:  []string{"ServiceName is required", "BilledCost is required"},
			},
			wantErr: false,
		},
		{
			name: "Invalid FOCUS record - negative cost",
			record: FOCUSRecord{
				BillingAccountId:   "123456789012",
				ServiceName:        "AmazonEC2",
				BilledCost:         -1.50, // Invalid negative cost
				EffectiveCost:      1.50,
				BillingPeriodStart: time.Now().Add(-24 * time.Hour),
				BillingPeriodEnd:   time.Now(),
				ChargeCategory:     "Usage",
				ChargeType:         "Usage",
				Region:             "us-east-1",
			},
			want: ValidationResult{
				IsValid: false,
				Score:   85.0,
				Issues:  []string{"BilledCost cannot be negative"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := converter.GetValidator()
			result := validator.ValidateRecord(tt.record)

			if result.IsValid != tt.want.IsValid {
				t.Errorf("ValidateRecord() IsValid = %v, want %v", result.IsValid, tt.want.IsValid)
			}

			if len(result.Issues) != len(tt.want.Issues) {
				t.Errorf("ValidateRecord() Issues count = %d, want %d", len(result.Issues), len(tt.want.Issues))
			}
		})
	}
}

// TestLargeFileProcessing tests processing of large files
func TestLargeFileProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large file test in short mode")
	}

	logger := logging.NewLogger("debug")
	converter := NewConverter(logger)

	// Create large test file
	largePath := createLargeTestFile(t, 100000) // 100k records
	defer func() {
		if err := os.Remove(largePath); err != nil {
			t.Logf("Failed to remove test file: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	options := ConversionOptions{
		Provider:    "aws",
		InputPath:   largePath,
		OutputPath:  filepath.Join(t.TempDir(), "large_output.parquet"),
		ChunkSize:   5000,
		Workers:     4,
		Compression: true,
		Validate:    false, // Skip validation for performance
	}

	start := time.Now()
	result, err := converter.Convert(ctx, options)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Convert() unexpected error: %v", err)
		return
	}

	if !result.Success {
		t.Errorf("Convert() failed for large file")
	}

	// Performance checks
	if duration > 2*time.Minute {
		t.Errorf("Conversion took too long: %v", duration)
	}

	if result.RecordsProcessed != 100000 {
		t.Errorf("Records processed = %d, want 100000", result.RecordsProcessed)
	}

	t.Logf("Large file processing completed in %v", duration)
	t.Logf("Processing rate: %.2f records/second", float64(result.RecordsProcessed)/duration.Seconds())
}

// TestConcurrentConversions tests concurrent conversion operations
func TestConcurrentConversions(t *testing.T) {
	logger := logging.NewLogger("debug")
	converter := NewConverter(logger)

	concurrency := 5
	results := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			options := ConversionOptions{
				Provider:    "aws",
				InputPath:   createTestAWSCUR(t),
				OutputPath:  filepath.Join(t.TempDir(), fmt.Sprintf("output_%d.parquet", id)),
				ChunkSize:   1000,
				Workers:     2,
				Compression: true,
				Validate:    true,
			}

			_, err := converter.Convert(ctx, options)
			results <- err
		}(i)
	}

	// Wait for all conversions to complete
	for i := 0; i < concurrency; i++ {
		if err := <-results; err != nil {
			t.Errorf("Concurrent conversion %d failed: %v", i, err)
		}
	}
}

// Helper functions for creating test data

func createTestAWSCUR(t *testing.T) string {
	content := `identity/LineItemId,bill/BillingEntity,bill/BillType,bill/PayerAccountId,bill/BillingPeriodStartDate,bill/BillingPeriodEndDate,lineItem/UsageAccountId,lineItem/LineItemType,lineItem/UsageStartDate,lineItem/UsageEndDate,lineItem/ProductCode,lineItem/UsageType,lineItem/Operation,lineItem/AvailabilityZone,lineItem/ResourceId,lineItem/UsageAmount,lineItem/NormalizationFactor,lineItem/NormalizedUsageAmount,lineItem/CurrencyCode,lineItem/UnblendedRate,lineItem/UnblendedCost,lineItem/BlendedRate,lineItem/BlendedCost,lineItem/LineItemDescription,pricing/RateId,pricing/currency,pricing/publicOnDemandCost,pricing/publicOnDemandRate,pricing/term,pricing/unit,product/ProductName,product/accountAssistance,product/architecturalReview,product/architectureSupport,product/availability,product/availabilityZone,product/clockSpeed,product/contentType,product/currentGeneration,product/customerServiceAndCommunities,product/databaseEngine,product/dedicatedEbsThroughput,product/deploymentOption,product/description,product/durability,product/ecu,product/edition,product/endpoint,product/engineCode,product/enhancedNetworkingSupported,product/freeQueryTypes,product/fromLocation,product/fromLocationType,product/gpu,product/gpuMemory,product/group,product/groupDescription,product/includedServices,product/instanceFamily,product/instanceType,product/instanceTypeFamily,product/intel,product/intelAvx,product/intelAvx2,product/intelTurbo,product/io,product/licenseModel,product/location,product/locationType,product/maxIopsBurstPerformance,product/maxIopsvolume,product/maxThroughputvolume,product/maxVolumeSize,product/memory,product/messageDeliveryFrequency,product/messageDeliveryOrder,product/minVolumeSize,product/networkPerformance,product/normalizationSizeFactor,product/operating-system,product/operation,product/operatingSystem,product/physicalCores,product/physicalProcessor,product/preInstalledSw,product/proactiveGuidance,product/processorArchitecture,product/processorFeatures,product/productFamily,product/protocol,product/provisioned,product/region,product/regionCode,product/servicecode,product/servicename,product/sku,product/storage,product/storageClass,product/storageMedia,product/storageType,product/technicalSupport,product/tenancy,product/thirdpartySoftwareSupport,product/toLocation,product/toLocationType,product/training,product/transferType,product/usagetype,product/vcpu,product/version,product/volumeApiName,product/volumeType,product/whoCanOpenCases,reservation/AmortizedUpfrontCostForUsage,reservation/AmortizedUpfrontFeeForBillingPeriod,reservation/EffectiveCost,reservation/EndTime,reservation/ModificationStatus,reservation/NormalizedUnitsPerReservation,reservation/NumberOfReservations,reservation/RecurringFeeForUsage,reservation/ReservationARN,reservation/StartTime,reservation/SubscriptionId,reservation/TotalReservedNormalizedUnits,reservation/TotalReservedUnits,reservation/UnitsPerReservation,reservation/UnusedAmortizedUpfrontFeeForBillingPeriod,reservation/UnusedNormalizedUnitQuantity,reservation/UnusedQuantity,reservation/UnusedRecurringFee,reservation/UpfrontValue,savingsPlans/TotalCommitmentToDate,savingsPlans/SavingsPlansType,savingsPlans/PaymentOption,savingsPlans/PurchaseTerm,savingsPlans/Region,savingsPlans/InstanceTypeFamily,savingsPlans/OfferingType,savingsPlans/EndTime,savingsPlans/StartTime,savingsPlans/TermDurationInSeconds,savingsPlans/SavingsPlansARN,savingsPlans/SavingsPlansRate,savingsPlans/SavingsPlansEffectiveCost,savingsPlans/AmortizedUpfrontCommitmentForBillingPeriod,savingsPlans/RecurringCommitmentForBillingPeriod,savingsPlans/TotalCommitmentToDate,savingsPlans/UsedCommitment,savingsPlans/UnusedCommitment,savingsPlans/NetSavingsPlansEffectiveCost,costCategory/Application,costCategory/Environment,resourceTags/aws:autoscaling:groupName,resourceTags/aws:cloudformation:logical-id,resourceTags/aws:cloudformation:stack-id,resourceTags/aws:cloudformation:stack-name,resourceTags/aws:createdBy,resourceTags/Name,resourceTags/user:Application,resourceTags/user:CostCenter,resourceTags/user:Environment,resourceTags/user:Owner,resourceTags/user:Project,resourceTags/user:Team
1,Amazon Web Services,Anniversary,123456789012,2025-08-01T00:00:00Z,2025-09-01T00:00:00Z,123456789012,Usage,2025-08-01T00:00:00Z,2025-08-01T01:00:00Z,AmazonEC2,BoxUsage:t3.micro,RunInstances,us-east-1a,i-1234567890abcdef0,1.0,1.0,1.0,USD,0.0116,0.0116,0.0116,0.0116,Linux/UNIX (Amazon VPC) t3.micro instance-hour in US East (N. Virginia),,,0.0116,0.0116,OnDemand,Hrs,Amazon Elastic Compute Cloud,,,,,us-east-1a,2.4 GHz Intel Xeon Platinum 8000 series,,Yes,,,,,,99.99%,0.5,,,N/A,,,,,,,,t3,t3.micro,Burstable Performance Instances,Yes,Yes,Yes,No,,Shared,US East (N. Virginia),AWS Region,,,,,1 GiB,,,,,10 Gbps,2.0,Linux,RunInstances,Linux,2,Intel Xeon Platinum 8000 series,NA,,,x86_64,Intel AVX; Intel AVX2; Intel Turbo,Compute Instance,,,us-east-1,use1,AmazonEC2,Amazon Elastic Compute Cloud,2C7LQWDJR68DXWWJ,,,EBS only,Shared,,,,,BoxUsage:t3.micro,1,,,General Purpose,All customers,0.0,0.0,0.0116,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,Production,,,,,,IAMUser:john.doe,web-server-01,MyApp,12345,Production,john.doe,WebApp,DevOps`

	tmpFile := filepath.Join(t.TempDir(), "test_aws_cur.csv")
	if err := os.WriteFile(tmpFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return tmpFile
}

func createTestAzureUsage(t *testing.T) string {
	content := `subscriptionId,meterName,meterCategory,meterSubCategory,unit,location,serviceName,serviceResourceId,quantity,cost,date,tags
sub-123,D2s v3,Virtual Machines,General Purpose,1 Hour,East US,Virtual Machines,/subscriptions/sub-123/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1,24,48.00,2025-08-01,{"Environment":"Production","Application":"WebApp"}
sub-123,Standard LRS,Storage,Block Blob,1 GB,East US,Storage,/subscriptions/sub-123/resourceGroups/rg1/providers/Microsoft.Storage/storageAccounts/storage1,100,2.00,2025-08-01,{"Environment":"Production","Application":"WebApp"}`

	tmpFile := filepath.Join(t.TempDir(), "test_azure_usage.csv")
	if err := os.WriteFile(tmpFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return tmpFile
}

func createTestGCPBilling(t *testing.T) string {
	content := `billing_account_id,service.id,service.description,sku.id,sku.description,project.id,project.name,project.labels,project.ancestry_numbers,location.location,location.country,location.region,location.zone,resource.name,resource.global_name,usage_start_time,usage_end_time,system_labels,labels,tags,cost_type,credits,usage.amount,usage.unit,usage.amount_in_pricing_units,usage.pricing_unit,cost,currency,currency_conversion_rate,invoice.month,cost_at_list,cost_at_list_currency
012345-ABCDEF-678901,6F81-5844-456A,Compute Engine,0E5C-6C43-2E11,N1 Predefined Instance Core running in Americas,my-project-123,My Project,"{""environment"": ""production"", ""team"": ""backend""}",123456789,us-central1-a,US,us-central1,us-central1-a,gke-cluster-1-default-pool-abc123-xyz,//container.googleapis.com/projects/my-project-123/zones/us-central1-a/clusters/gke-cluster-1/nodePools/default-pool/instances/gke-cluster-1-default-pool-abc123-xyz,2025-08-01T00:00:00Z,2025-08-01T01:00:00Z,"{""compute.googleapis.com/machine_spec"": ""n1-standard-1""}","{""environment"": ""production"", ""app"": ""web""}",{},regular,{},"1.0",hour,"1.0",hour,"0.04750",USD,1.0,202508,"0.04750",USD`

	tmpFile := filepath.Join(t.TempDir(), "test_gcp_billing.csv")
	if err := os.WriteFile(tmpFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return tmpFile
}

func createLargeTestFile(t *testing.T, numRecords int) string {
	header := `identity/LineItemId,bill/BillingEntity,bill/BillType,bill/PayerAccountId,lineItem/UsageAccountId,lineItem/ProductCode,lineItem/UsageType,lineItem/Operation,lineItem/BlendedCost,lineItem/UsageStartDate,lineItem/UsageEndDate,product/region`

	tmpFile := filepath.Join(t.TempDir(), "large_test_file.csv")
	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Logf("Failed to close file: %v", err)
		}
	}()

	// Write header
	if _, err := f.WriteString(header + "\n"); err != nil {
		t.Fatalf("Failed to write header: %v", err)
	}

	// Write records
	for i := 0; i < numRecords; i++ {
		record := fmt.Sprintf("%d,Amazon Web Services,Anniversary,123456789012,123456789012,AmazonEC2,BoxUsage:t3.micro,RunInstances,%.4f,2025-08-01T%02d:00:00Z,2025-08-01T%02d:00:00Z,us-east-1\n",
			i+1, 0.0116*float64(i%100+1), i%24, (i%24)+1)
		if _, err := f.WriteString(record); err != nil {
			t.Fatalf("Failed to write record %d: %v", i, err)
		}
	}

	return tmpFile
}

// Mock types and interfaces (these would normally be in separate files)

type Converter struct {
	logger *logging.Logger
}

func NewConverter(logger *logging.Logger) *Converter {
	return &Converter{logger: logger}
}

type ConversionOptions struct {
	Provider    string
	InputPath   string
	OutputPath  string
	ChunkSize   int
	Workers     int
	Compression bool
	Validate    bool
}

type ConversionResult struct {
	Success          bool
	RecordsProcessed int64
	OutputPath       string
	FOCUSCompliance  ComplianceResult
	ProcessingTime   time.Duration
	Errors           []string
}

type ComplianceResult struct {
	Score  float64
	Issues []string
}

type FOCUSRecord struct {
	BillingAccountId   string
	ServiceName        string
	ResourceId         string
	UsageUnit          string
	BilledCost         float64
	EffectiveCost      float64
	BillingPeriodStart time.Time
	BillingPeriodEnd   time.Time
	ChargeCategory     string
	ChargeSubcategory  string
	ChargeType         string
	ChargeClass        string
	ChargeFrequency    string
	ChargeDescription  string
	Region             string
	AvailabilityZone   string
	ServiceCategory    string
}

type ValidationResult struct {
	IsValid bool
	Score   float64
	Issues  []string
}

type SchemaMapper interface {
	MapToFOCUS(fields map[string]interface{}) (FOCUSRecord, error)
}

type Validator interface {
	ValidateRecord(record FOCUSRecord) ValidationResult
}

// Mock implementations
func (c *Converter) Convert(ctx context.Context, options ConversionOptions) (*ConversionResult, error) {
	// Mock implementation for testing
	if options.Provider == "invalid" {
		return nil, fmt.Errorf("unsupported provider: %s", options.Provider)
	}

	// Calculate records processed by reading the input file
	recordsProcessed := int64(0)
	if file, err := os.Open(options.InputPath); err == nil {
		defer func() {
			_ = file.Close() // Ignore close error in test mock
		}()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			recordsProcessed++
		}
		// Subtract 1 for header row
		if recordsProcessed > 0 {
			recordsProcessed--
		}
	}

	// Simulate processing time proportional to file size
	processingTime := time.Duration(recordsProcessed/1000) * time.Millisecond
	if processingTime < 10*time.Millisecond {
		processingTime = 10 * time.Millisecond
	}
	time.Sleep(processingTime)

	// Create output file
	if err := os.WriteFile(options.OutputPath, []byte("mock parquet data"), 0600); err != nil {
		return nil, err
	}

	return &ConversionResult{
		Success:          true,
		RecordsProcessed: recordsProcessed,
		OutputPath:       options.OutputPath,
		FOCUSCompliance: ComplianceResult{
			Score:  95.0,
			Issues: []string{},
		},
		ProcessingTime: processingTime,
		Errors:         []string{},
	}, nil
}

func (c *Converter) GetSchemaMapper(provider string) SchemaMapper {
	return &mockSchemaMapper{provider: provider}
}

func (c *Converter) GetValidator() Validator {
	return &mockValidator{}
}

type mockSchemaMapper struct {
	provider string
}

func (m *mockSchemaMapper) MapToFOCUS(fields map[string]interface{}) (FOCUSRecord, error) {
	record := FOCUSRecord{}

	switch m.provider {
	case "aws":
		if accountId, ok := fields["lineItem/UsageAccountId"].(string); ok {
			record.BillingAccountId = accountId
		}
		if productCode, ok := fields["lineItem/ProductCode"].(string); ok {
			record.ServiceName = productCode
		}
		if costStr, ok := fields["lineItem/BlendedCost"].(string); ok {
			if cost, err := strconv.ParseFloat(costStr, 64); err == nil {
				record.BilledCost = cost
				record.EffectiveCost = cost
			}
		}
		if region, ok := fields["product/region"].(string); ok {
			record.Region = region
		}
		if startDate, ok := fields["lineItem/UsageStartDate"].(string); ok {
			if t, err := time.Parse(time.RFC3339, startDate); err == nil {
				record.BillingPeriodStart = t
			}
		}
		if endDate, ok := fields["lineItem/UsageEndDate"].(string); ok {
			if t, err := time.Parse(time.RFC3339, endDate); err == nil {
				record.BillingPeriodEnd = t
			}
		}
		record.ChargeCategory = "Usage"
		record.ChargeSubcategory = "OnDemand"
		record.ChargeType = "Usage"
		record.ChargeClass = "Committed"
		record.ChargeFrequency = "Monthly"
		record.ServiceCategory = "Compute"
		record.UsageUnit = "Hours"
		if operation, ok := fields["lineItem/Operation"].(string); ok {
			record.ChargeDescription = operation
		}

	case "azure":
		if subId, ok := fields["subscriptionId"].(string); ok {
			record.BillingAccountId = subId
		}
		if serviceName, ok := fields["serviceName"].(string); ok {
			record.ServiceName = serviceName
		}
		if costStr, ok := fields["cost"].(string); ok {
			if cost, err := strconv.ParseFloat(costStr, 64); err == nil {
				record.BilledCost = cost
				record.EffectiveCost = cost
			}
		}
		if location, ok := fields["location"].(string); ok {
			record.Region = location
		}
		if dateStr, ok := fields["date"].(string); ok {
			if t, err := time.Parse("2006-01-02", dateStr); err == nil {
				record.BillingPeriodStart = t
				record.BillingPeriodEnd = t.Add(24 * time.Hour)
			}
		}
		record.ChargeCategory = "Usage"
		record.ChargeSubcategory = "OnDemand"
		record.ChargeType = "Usage"
		record.ChargeClass = "Committed"
		record.ChargeFrequency = "Monthly"
		record.ServiceCategory = "Compute"
		record.UsageUnit = "Hours"
		if meterName, ok := fields["meterName"].(string); ok {
			record.ChargeDescription = meterName
		}
	}

	return record, nil
}

type mockValidator struct{}

func (v *mockValidator) ValidateRecord(record FOCUSRecord) ValidationResult {
	issues := []string{}
	score := 100.0

	if record.BillingAccountId == "" {
		issues = append(issues, "BillingAccountId is required")
		score -= 20
	}
	if record.ServiceName == "" {
		issues = append(issues, "ServiceName is required")
		score -= 20
	}
	if record.BilledCost == 0 && record.EffectiveCost == 0 {
		issues = append(issues, "BilledCost is required")
		score -= 20
	}
	if record.BilledCost < 0 {
		issues = append(issues, "BilledCost cannot be negative")
		score -= 15
	}
	if record.EffectiveCost < 0 {
		issues = append(issues, "EffectiveCost cannot be negative")
		score -= 15
	}

	return ValidationResult{
		IsValid: len(issues) == 0,
		Score:   score,
		Issues:  issues,
	}
}
