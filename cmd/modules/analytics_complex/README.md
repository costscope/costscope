# Analytics Complex Module

## Overview

The Analytics Complex module provides enterprise-grade, type-safe analytics CLI with advanced ML capabilities for CostScope. This module represents the next evolution of cost analytics, combining type safety, machine learning, and performance optimization.

## Features

###  Machine Learning Capabilities
- **ML-powered forecasting** with multiple models (ARIMA, LSTM, Prophet)
- **Real-time anomaly detection** with configurable sensitivity
- **Advanced optimization algorithms** (Genetic, Particle Swarm)
- **Automated feature engineering** and model validation

###  Type-Safe Operations
- **Type-safe filter system** with automatic validation
- **Compile-time type checking** for all parameters
- **Automatic type conversion** with error handling
- **Schema validation** for complex configurations

###  Performance Optimization
- **Parallel processing** with configurable workers
- **Intelligent caching** with TTL management
- **Memory optimization** for large datasets
- **Batch processing** for streaming data

###  Advanced Transformations
- **Complex data aggregations** and pivots
- **Multi-dimensional normalization** and scaling
- **Custom transformation pipelines**
- **Real-time data streaming** support

###  Custom Analytics
- **User-defined SQL-like queries** for complex analysis
- **Custom transformation pipelines** with validation
- **Advanced scripting support** with multiple languages
- **Dynamic report generation** with flexible output formats

## CLI Commands

### Main Command
```bash
costscope analytics-complex
```

### Subcommands

#### 1. Analyze
Advanced analytics analysis with type-safe filtering and ML-powered insights.

```bash
# Basic analysis with type-safe filters
costscope analytics-complex analyze --service "ec2,rds" --region "us-east-1,us-west-2"

# Advanced analysis with ML features
costscope analytics-complex analyze --ml-enabled --anomaly-detection --forecast-periods 30

# Performance optimized analysis
costscope analytics-complex analyze --parallel --workers 8 --cache-enabled
```

**Key Flags:**
- `--service`: Service filter (comma-separated)
- `--region`: Region filter (comma-separated)
- `--ml-enabled`: Enable ML-powered analytics
- `--anomaly-detection`: Enable anomaly detection
- `--parallel`: Enable parallel processing
- `--workers`: Number of worker threads

#### 2. Forecast
ML-powered cost forecasting with confidence intervals.

```bash
# Basic forecasting with auto-selected model
costscope analytics-complex forecast --days 90

# Advanced forecasting with specific model
costscope analytics-complex forecast --model lstm --days 30 --confidence 95

# Ensemble forecasting with validation
costscope analytics-complex forecast --model ensemble --validation --uncertainty
```

**Supported Models:**
- `auto-arima`: Automatic ARIMA model selection
- `lstm`: Long Short-Term Memory neural networks
- `prophet`: Facebook Prophet time series forecasting
- `ensemble`: Combined model approach for best accuracy

**Key Flags:**
- `--model`: ML model selection
- `--days`: Forecast period in days
- `--confidence`: Confidence interval percentage
- `--validation`: Enable model validation

#### 3. Detect
Real-time anomaly detection with configurable sensitivity.

```bash
# Basic anomaly detection
costscope analytics-complex detect --sensitivity medium

# Real-time detection with alerts
costscope analytics-complex detect --real-time --alerts slack --threshold 0.95

# Advanced detection with custom parameters
costscope analytics-complex detect --method ensemble --window 7d --sensitivity high
```

**Detection Methods:**
- `isolation-forest`: Isolation Forest algorithm for outlier detection
- `one-class-svm`: One-Class Support Vector Machine
- `statistical`: Statistical-based anomaly detection
- `ensemble`: Combined approach for better accuracy

**Key Flags:**
- `--method`: Detection method
- `--sensitivity`: Detection sensitivity (low, medium, high)
- `--real-time`: Enable real-time detection
- `--alerts`: Alert channels (slack, email, webhook)

#### 4. Transform
Complex data transformations with type safety.

```bash
# Basic aggregation transformation
costscope analytics-complex transform --type aggregate --target cost --group-by service

# Advanced pivot transformation
costscope analytics-complex transform --type pivot --rows service --columns region --values cost

# Normalization with scaling
costscope analytics-complex transform --type normalize --method minmax --target cost
```

**Transformation Types:**
- `aggregate`: Data aggregation operations
- `pivot`: Data pivoting and reshaping
- `normalize`: Data normalization and scaling
- `filter`: Advanced filtering operations
- `join`: Data joining and merging

#### 5. Optimize
Advanced cost optimization with ML algorithms.

```bash
# Basic cost optimization
costscope analytics-complex optimize --algorithm genetic --target cost

# Multi-objective optimization
costscope analytics-complex optimize --algorithm particle-swarm --target both --multi-objective

# Custom optimization with constraints
costscope analytics-complex optimize --algorithm simulated-annealing --constraints constraints.json
```

**Optimization Algorithms:**
- `genetic`: Genetic algorithm optimization
- `particle-swarm`: Particle Swarm Optimization
- `simulated-annealing`: Simulated Annealing
- `gradient-descent`: Gradient Descent optimization

#### 6. Custom
Custom analytics with user-defined queries and transformations.

```bash
# Custom SQL-like query
costscope analytics-complex custom --query "SELECT service, SUM(cost) FROM costs GROUP BY service" --validate

# Custom transformation pipeline
costscope analytics-complex custom --pipeline config.yaml --model custom-model

# Advanced scripting with validation
costscope analytics-complex custom --script analytics.py --validate --dry-run

# Custom ML model integration
costscope analytics-complex custom --model ensemble --config custom.yaml --output json
```

**Custom Features:**
- `query`: SQL-like query support for complex analysis
- `pipeline`: Custom transformation pipeline configuration
- `script`: Advanced scripting support (Python, JavaScript, R)
- `model`: Custom ML model integration
- `validate`: Query and script validation before execution
- `dry-run`: Validation without actual execution

## Technical Implementation

### Architecture
```
cmd/modules/analytics_complex/
├── commands/
│   ├── command_definition.go    # Main command structure and types
│   ├── command_handlers.go      # Command implementation and handlers
│   └── command_test.go         # Comprehensive test suite
└── register.go                 # Integration with root CLI
```

### Key Components

#### TypeSafeFilterConfig
Provides type-safe configuration for analytics filters with automatic validation:

```go
type TypeSafeFilterConfig struct {
    ServiceFilter    *FilterValue[[]string]
    RegionFilter     *FilterValue[[]string]
    CostThreshold    *FilterValue[float64]
    DateRange        *FilterValue[DateRange]
    // ... more filters
}
```

#### FilterValue
Generic type-safe filter values with automatic conversion:

```go
type FilterValue[T any] struct {
    Value     T      `json:"value"`
    Type      string `json:"type"`
    Operator  string `json:"operator,omitempty"`
    Validated bool   `json:"validated"`
}
```

#### MLConfiguration
Comprehensive ML-specific settings:

```go
type MLConfiguration struct {
    EnableForecasting      bool
    ForecastModel         string
    EnableAnomalyDetection bool
    AnomalyMethod         string
    EnableOptimization    bool
    OptimizationAlgorithm string
    // ... more ML settings
}
```

### Integration with Core Systems

The module integrates with:
- `internal/core/analytics_advanced`: Advanced analytics service
- `internal/core/logging`: Structured logging
- `internal/core/focus/analysis`: FOCUS dataset analysis

### Performance Features

1. **Parallel Processing**: Configurable worker pools for concurrent operations
2. **Intelligent Caching**: TTL-based caching with configurable policies
3. **Memory Optimization**: Efficient memory usage for large datasets
4. **Batch Processing**: Optimized batch operations for streaming data

### Testing

Comprehensive test suite includes:
- Command structure validation
- Flag presence and default value testing
- Type-safe filter creation testing
- Subcommand registration testing

Run tests:
```bash
go test ./cmd/modules/analytics_complex/...
```

## Usage Examples

### Complete Analysis Workflow

```bash
# 1. Comprehensive analysis with ML
costscope analytics-complex analyze \
  --service "ec2,rds,s3" \
  --region "us-east-1,us-west-2" \
  --ml-enabled \
  --anomaly-detection \
  --forecast-enabled \
  --parallel \
  --workers 8

# 2. Detailed forecasting
costscope analytics-complex forecast \
  --model ensemble \
  --days 90 \
  --confidence 95 \
  --validation \
  --uncertainty

# 3. Real-time anomaly monitoring
costscope analytics-complex detect \
  --method ensemble \
  --sensitivity high \
  --real-time \
  --alerts slack

# 4. Data transformation pipeline
costscope analytics-complex transform \
  --type aggregate \
  --target cost \
  --group-by service,region \
  --optimize-memory

# 5. Cost optimization recommendations
costscope analytics-complex optimize \
  --algorithm genetic \
  --target both \
  --multi-objective \
  --iterations 2000
```

## Future Enhancements

- **Custom Model Training**: Support for training custom ML models
- **Streaming Analytics**: Real-time streaming data processing
- **Advanced Alerts**: Integration with more alerting systems
- **Custom Transformations**: User-defined transformation functions
- **Model Registry**: Management of trained ML models

## Contributing

When contributing to this module:

1. **Type Safety**: Ensure all new features maintain type safety
2. **Testing**: Add comprehensive tests for new functionality
3. **Documentation**: Update this README for new features
4. **Performance**: Consider performance implications of changes
5. **ML Best Practices**: Follow ML engineering best practices

## Related Modules

- `analytics`: Basic analytics functionality
- `analytics_advanced`: Advanced analytics with ML
- `focus`: FOCUS dataset operations
- `reports`: Report generation
- `production`: Production readiness features
