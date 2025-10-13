package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awscreds "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/costscope/costscope/internal/core/logging"
	"github.com/costscope/costscope/internal/providers/types"
)

// stsAPI abstracts STS for testability.
type stsAPI interface {
	GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

// AWSProvider implements the CloudProvider interface for AWS
type AWSProvider struct {
	name       string
	config     aws.Config
	region     string
	accessKey  string
	secretKey  string
	logger     *logging.Logger
	stsClient  stsAPI
	stsFactory func(aws.Config) stsAPI
}

// Name returns provider instance name (implements registry.Provider)
func (p *AWSProvider) Name() string { return p.name }

// NewAWSProvider creates a new AWS provider instance
func NewAWSProvider(providerConfig *types.ProviderConfig) (*AWSProvider, error) {
	logger := logging.NewLogger(logging.LevelInfo)

	accessKey, ok := providerConfig.Credentials["access_key"]
	if !ok || accessKey == "" {
		return nil, fmt.Errorf("access_key is required in credentials")
	}

	secretKey, ok := providerConfig.Credentials["secret_key"]
	if !ok || secretKey == "" {
		return nil, fmt.Errorf("secret_key is required in credentials")
	}

	region := "us-east-1" // default region
	if r, ok := providerConfig.Settings["region"]; ok {
		if regionStr, ok := r.(string); ok && regionStr != "" {
			region = regionStr
		}
	}

	// Create AWS config with credentials
	awsConfig, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(awscreds.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"", // session token (optional)
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	provider := &AWSProvider{
		name:       providerConfig.Name,
		config:     awsConfig,
		region:     region,
		accessKey:  accessKey,
		secretKey:  secretKey,
		logger:     logger,
		stsFactory: func(c aws.Config) stsAPI { return sts.NewFromConfig(c) },
	}
	provider.stsClient = provider.stsFactory(awsConfig)

	logger.Info("AWS provider initialized successfully")
	return provider, nil
}

// ValidateCredentials validates AWS credentials by calling STS GetCallerIdentity
func (p *AWSProvider) ValidateCredentials(ctx context.Context, credentials map[string]string) error {
	p.logger.Info("Validating AWS credentials")

	accessKey, ok := credentials["access_key"]
	if !ok || accessKey == "" {
		return fmt.Errorf("access_key is required")
	}

	secretKey, ok := credentials["secret_key"]
	if !ok || secretKey == "" {
		return fmt.Errorf("secret_key is required")
	}

	// Create static credentials provider
	credProvider := awscreds.NewStaticCredentialsProvider(accessKey, secretKey, "")

	// Create temporary config for validation
	tempConfig, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(p.region),
		config.WithCredentialsProvider(credProvider),
	)
	if err != nil {
		return fmt.Errorf("failed to create AWS config for validation: %w", err)
	}

	stsClient := p.stsFactory(tempConfig)
	if _, err = stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{}); err != nil {
		p.logger.Error(fmt.Sprintf("AWS credential validation failed: %v", err))
		return fmt.Errorf("invalid AWS credentials: %w", err)
	}

	p.logger.Info("AWS credentials validated successfully")
	return nil
}

// GetProviderInfo returns AWS provider information
func (p *AWSProvider) GetProviderInfo(ctx context.Context) (types.ProviderInfo, error) {
	p.logger.Info("Getting AWS provider information")
	identity, err := p.stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to get AWS caller identity: %v", err))
		return types.ProviderInfo{}, fmt.Errorf("failed to get AWS account info: %w", err)
	}
	accountID := ""
	userArn := ""
	if identity.Account != nil {
		accountID = *identity.Account
	}
	if identity.Arn != nil {
		userArn = *identity.Arn
	}
	info := types.ProviderInfo{
		Name:             "AWS",
		Type:             types.ProviderTypeAWS,
		Version:          "1.0.0",
		SupportedRegions: p.getSupportedRegions(),
		Capabilities:     []string{"cost-data", "resource-data", "account-info"},
		Metadata: map[string]string{
			"region":      p.region,
			"account_id":  accountID,
			"user_arn":    userArn,
			"sdk_version": "2.0",
		},
	}
	p.logger.Info("AWS provider information retrieved successfully")
	return info, nil
}

// GetCostData retrieves cost data from AWS (placeholder implementation)
func (p *AWSProvider) GetCostData(ctx context.Context, params types.CostDataParams) ([]types.CostRecord, error) {
	p.logger.Info("Getting AWS cost data")
	records := p.createSampleCostRecords()
	p.logDataRetrieval("cost", len(records))
	return records, nil
}

// GetResourceData retrieves resource data from AWS (placeholder implementation)
func (p *AWSProvider) GetResourceData(ctx context.Context, params types.ResourceDataParams) ([]types.ResourceRecord, error) {
	p.logger.Info("Getting AWS resource data")

	// This is a placeholder implementation
	// In the real implementation, this would query AWS APIs for resources
	resources := p.createSampleResourceRecords()

	p.logDataRetrieval("resource", len(resources))
	return resources, nil
}

// Helper function to log data retrieval
func (p *AWSProvider) logDataRetrieval(dataType string, count int) {
	p.logger.Info(fmt.Sprintf("Retrieved %d %s records from AWS", count, dataType))
}

// Helper function to create sample cost records
func (p *AWSProvider) createSampleCostRecords() []types.CostRecord {
	return []types.CostRecord{
		{
			Date:       time.Now().Add(-24 * time.Hour),
			Amount:     123.45,
			Currency:   "USD",
			Service:    "EC2-Instance",
			Region:     p.region,
			ResourceID: "i-1234567890abcdef0",
			Tags:       map[string]string{"Environment": "production", "Team": "backend"},
			Metadata: map[string]interface{}{
				"instance_type": "t3.medium",
				"usage_type":    "BoxUsage:t3.medium",
			},
		},
		{
			Date:       time.Now().Add(-24 * time.Hour),
			Amount:     67.89,
			Currency:   "USD",
			Service:    "S3",
			Region:     p.region,
			ResourceID: "bucket-example-data",
			Tags:       map[string]string{"Environment": "production", "Team": "data"},
			Metadata: map[string]interface{}{
				"storage_class": "STANDARD",
				"usage_type":    "TimedStorage-ByteHrs",
			},
		},
	}
}

// Helper function to create sample resource records
func (p *AWSProvider) createSampleResourceRecords() []types.ResourceRecord {
	return []types.ResourceRecord{
		{
			ID:       "i-1234567890abcdef0",
			Name:     "web-server-prod",
			Type:     "EC2.Instance",
			Region:   p.region,
			Status:   "running",
			Cost:     123.45,
			Currency: "USD",
			Tags:     map[string]string{"Environment": "production", "Team": "backend"},
			Properties: map[string]interface{}{
				"instance_type":     "t3.medium",
				"availability_zone": p.region + "a",
				"vpc_id":            "vpc-12345678",
				"subnet_id":         "subnet-12345678",
			},
			CreatedAt: time.Now().Add(-72 * time.Hour),
			UpdatedAt: time.Now(),
		},
		{
			ID:       "bucket-example-data",
			Name:     "example-data-bucket",
			Type:     "S3.Bucket",
			Region:   p.region,
			Status:   "active",
			Cost:     67.89,
			Currency: "USD",
			Tags:     map[string]string{"Environment": "production", "Team": "data"},
			Properties: map[string]interface{}{
				"storage_class": "STANDARD",
				"versioning":    "Enabled",
				"encryption":    "AES256",
				"public_access": false,
			},
			CreatedAt: time.Now().Add(-168 * time.Hour),
			UpdatedAt: time.Now(),
		},
	}
}

// GetName returns the provider name
func (p *AWSProvider) GetName() string {
	return p.name
}

// GetType returns the provider type
func (p *AWSProvider) GetType() types.ProviderType {
	return types.ProviderTypeAWS
}

// GetSupportedRegions returns list of supported AWS regions
func (p *AWSProvider) GetSupportedRegions() []string {
	return p.getSupportedRegions()
}

// getSupportedRegions returns a list of commonly used AWS regions
func (p *AWSProvider) getSupportedRegions() []string {
	return []string{
		"us-east-1",      // N. Virginia
		"us-east-2",      // Ohio
		"us-west-1",      // N. California
		"us-west-2",      // Oregon
		"eu-west-1",      // Ireland
		"eu-west-2",      // London
		"eu-central-1",   // Frankfurt
		"ap-southeast-1", // Singapore
		"ap-southeast-2", // Sydney
		"ap-northeast-1", // Tokyo
		"ap-south-1",     // Mumbai
		"sa-east-1",      // São Paulo
	}
}
