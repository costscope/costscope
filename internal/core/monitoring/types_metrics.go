package monitoring

import "time"

// RealTimeMetrics contains real-time monitoring data
type RealTimeMetrics struct {
	Timestamp        time.Time          `json:"timestamp"`
	System           SystemMetrics      `json:"system"`
	Performance      PerformanceMetrics `json:"performance"`
	Resources        ResourceMetrics    `json:"resources"`
	Applications     ApplicationMetrics `json:"applications"`
	Business         BusinessMetrics    `json:"business"`
	Integrations     IntegrationMetrics `json:"integrations"`
	Providers        ProviderMetrics    `json:"providers"`
	ActiveAlerts     int                `json:"active_alerts"`
	HealthScore      int                `json:"health_score"`
	TrendIndicators  map[string]string  `json:"trend_indicators"`
	CollectionTimeMs int64              `json:"collection_time_ms"`
}

// SystemMetrics contains core system metrics
type SystemMetrics struct {
	Hostname          string        `json:"hostname"`
	Uptime            time.Duration `json:"uptime"`
	LoadAverage       []float64     `json:"load_average"`
	ProcessCount      int           `json:"process_count"`
	ThreadCount       int           `json:"thread_count"`
	FileDescriptors   int           `json:"file_descriptors"`
	SocketConnections int           `json:"socket_connections"`
	SystemCalls       int64         `json:"system_calls"`
	ContextSwitches   int64         `json:"context_switches"`
	Interrupts        int64         `json:"interrupts"`
	OSVersion         string        `json:"os_version"`
	KernelVersion     string        `json:"kernel_version"`
	Architecture      string        `json:"architecture"`
}

// ResourceMetrics contains resource utilization data
type ResourceMetrics struct {
	CPU                CPUMetrics         `json:"cpu"`
	Memory             MemoryMetrics      `json:"memory"`
	Disk               DiskMetrics        `json:"disk"`
	Network            NetworkMetrics     `json:"network"`
	GPU                []GPUMetrics       `json:"gpu"`
	Containers         []ContainerMetrics `json:"containers"`
	Services           []ServiceMetrics   `json:"services"`
	ResourceScore      int                `json:"resource_score"`
	UtilizationSummary string             `json:"utilization_summary"`
}

// ApplicationMetrics contains application-specific metrics
type ApplicationMetrics struct {
	RequestsPerSecond   float64                `json:"requests_per_second"`
	ResponseTime        LatencyMetrics         `json:"response_time"`
	ErrorRate           float64                `json:"error_rate"`
	ActiveSessions      int                    `json:"active_sessions"`
	ConcurrentUsers     int                    `json:"concurrent_users"`
	TransactionRate     float64                `json:"transaction_rate"`
	CacheHitRate        float64                `json:"cache_hit_rate"`
	DatabaseConnections int                    `json:"database_connections"`
	QueueDepth          int                    `json:"queue_depth"`
	TaskCompletionRate  float64                `json:"task_completion_rate"`
	FeatureUsage        map[string]int         `json:"feature_usage"`
	CustomMetrics       map[string]interface{} `json:"custom_metrics"`
}

// BusinessMetrics contains business-related metrics
type BusinessMetrics struct {
	CostSavings          float64               `json:"cost_savings"`
	ROI                  float64               `json:"roi"`
	Efficiency           float64               `json:"efficiency"`
	CustomerSatisfaction float64               `json:"customer_satisfaction"`
	SLA                  SLAMetrics            `json:"sla"`
	Availability         float64               `json:"availability"`
	MTTR                 float64               `json:"mttr"`
	MTBF                 float64               `json:"mtbf"`
	IncidentCount        int                   `json:"incident_count"`
	RevenueImpact        float64               `json:"revenue_impact"`
	UserEngagement       UserEngagementMetrics `json:"user_engagement"`
	BusinessKPIs         map[string]float64    `json:"business_kpis"`
}

// IntegrationMetrics contains metrics from integrations
type IntegrationMetrics struct {
	ConnectedSystems  int                    `json:"connected_systems"`
	HealthySystems    int                    `json:"healthy_systems"`
	FailedSystems     int                    `json:"failed_systems"`
	ActiveWorkflows   int                    `json:"active_workflows"`
	CompletedTasks    int                    `json:"completed_tasks"`
	FailedTasks       int                    `json:"failed_tasks"`
	DataThroughput    float64                `json:"data_throughput"`
	SyncLatency       float64                `json:"sync_latency"`
	IntegrationHealth float64                `json:"integration_health"`
	AlertCount        int                    `json:"alert_count"`
	SystemMappings    map[string]string      `json:"system_mappings"`
	ProcessingMetrics map[string]interface{} `json:"processing_metrics"`
}

// ProviderMetrics contains cloud provider metrics
type ProviderMetrics struct {
	ActiveProviders  []string               `json:"active_providers"`
	TotalResources   int                    `json:"total_resources"`
	CostData         map[string]float64     `json:"cost_data"`
	PerformanceData  map[string]float64     `json:"performance_data"`
	HealthStatus     map[string]string      `json:"health_status"`
	APICallRate      map[string]int         `json:"api_call_rate"`
	QuotaUtilization map[string]float64     `json:"quota_utilization"`
	RegionMetrics    map[string]interface{} `json:"region_metrics"`
	ServiceMetrics   map[string]interface{} `json:"service_metrics"`
	BillingMetrics   BillingMetrics         `json:"billing_metrics"`
	ProviderHealth   float64                `json:"provider_health"`
}

// Supporting types for detailed metrics
type CPUMetrics struct {
	UsagePercent   float64 `json:"usage_percent"`
	Cores          int     `json:"cores"`
	Frequency      float64 `json:"frequency"`
	LoadAverage1m  float64 `json:"load_average_1m"`
	LoadAverage5m  float64 `json:"load_average_5m"`
	LoadAverage15m float64 `json:"load_average_15m"`
	UserTime       float64 `json:"user_time"`
	SystemTime     float64 `json:"system_time"`
	IdleTime       float64 `json:"idle_time"`
	IOWaitTime     float64 `json:"iowait_time"`
}

type MemoryMetrics struct {
	TotalGB      float64 `json:"total_gb"`
	UsedGB       float64 `json:"used_gb"`
	FreeGB       float64 `json:"free_gb"`
	UsagePercent float64 `json:"usage_percent"`
	BuffersGB    float64 `json:"buffers_gb"`
	CachedGB     float64 `json:"cached_gb"`
	SwapTotalGB  float64 `json:"swap_total_gb"`
	SwapUsedGB   float64 `json:"swap_used_gb"`
	SwapFreeGB   float64 `json:"swap_free_gb"`
	PageFaults   int64   `json:"page_faults"`
}

type DiskMetrics struct {
	TotalGB        float64 `json:"total_gb"`
	UsedGB         float64 `json:"used_gb"`
	FreeGB         float64 `json:"free_gb"`
	UsagePercent   float64 `json:"usage_percent"`
	ReadOpsPerSec  float64 `json:"read_ops_per_sec"`
	WriteOpsPerSec float64 `json:"write_ops_per_sec"`
	ReadMBPerSec   float64 `json:"read_mb_per_sec"`
	WriteMBPerSec  float64 `json:"write_mb_per_sec"`
	IOUtilPercent  float64 `json:"io_util_percent"`
	QueueDepth     float64 `json:"queue_depth"`
}

type NetworkMetrics struct {
	BytesReceivedPerSec   float64 `json:"bytes_received_per_sec"`
	BytesSentPerSec       float64 `json:"bytes_sent_per_sec"`
	PacketsReceivedPerSec float64 `json:"packets_received_per_sec"`
	PacketsSentPerSec     float64 `json:"packets_sent_per_sec"`
	ErrorsReceived        int64   `json:"errors_received"`
	ErrorsSent            int64   `json:"errors_sent"`
	DroppedPackets        int64   `json:"dropped_packets"`
	NetworkLatency        float64 `json:"network_latency"`
	Bandwidth             float64 `json:"bandwidth"`
	Connections           int     `json:"connections"`
}

// Extended metrics types used by resources/business/provider sections
type GPUMetrics struct {
	DeviceID           string  `json:"device_id"`
	Name               string  `json:"name"`
	UtilizationPercent float64 `json:"utilization_percent"`
	MemoryUsedMB       float64 `json:"memory_used_mb"`
	MemoryTotalMB      float64 `json:"memory_total_mb"`
	Temperature        float64 `json:"temperature"`
	PowerUsageWatts    float64 `json:"power_usage_watts"`
	FanSpeed           float64 `json:"fan_speed"`
}

type ContainerMetrics struct {
	ContainerID  string         `json:"container_id"`
	Name         string         `json:"name"`
	Image        string         `json:"image"`
	Status       string         `json:"status"`
	CPUUsage     float64        `json:"cpu_usage"`
	MemoryUsage  float64        `json:"memory_usage"`
	NetworkIO    NetworkMetrics `json:"network_io"`
	DiskIO       DiskMetrics    `json:"disk_io"`
	RestartCount int            `json:"restart_count"`
}

type ServiceMetrics struct {
	ServiceName  string        `json:"service_name"`
	Status       string        `json:"status"`
	Health       string        `json:"health"`
	Version      string        `json:"version"`
	Uptime       time.Duration `json:"uptime"`
	ResponseTime float64       `json:"response_time"`
	ErrorRate    float64       `json:"error_rate"`
	Dependencies []string      `json:"dependencies"`
}

type SLAMetrics struct {
	TargetUptime       float64 `json:"target_uptime"`
	ActualUptime       float64 `json:"actual_uptime"`
	SLABreach          bool    `json:"sla_breach"`
	ResponseTimeSLA    float64 `json:"response_time_sla"`
	ActualResponseTime float64 `json:"actual_response_time"`
	ErrorRateSLA       float64 `json:"error_rate_sla"`
	ActualErrorRate    float64 `json:"actual_error_rate"`
	ComplianceScore    float64 `json:"compliance_score"`
}

type UserEngagementMetrics struct {
	ActiveUsers        int     `json:"active_users"`
	DailyActiveUsers   int     `json:"daily_active_users"`
	MonthlyActiveUsers int     `json:"monthly_active_users"`
	SessionDuration    float64 `json:"session_duration"`
	BounceRate         float64 `json:"bounce_rate"`
	ConversionRate     float64 `json:"conversion_rate"`
	UserSatisfaction   float64 `json:"user_satisfaction"`
}

type BillingMetrics struct {
	TotalCost         float64            `json:"total_cost"`
	DailyCost         float64            `json:"daily_cost"`
	MonthlyCost       float64            `json:"monthly_cost"`
	CostTrend         string             `json:"cost_trend"`
	BudgetUtilization float64            `json:"budget_utilization"`
	CostByService     map[string]float64 `json:"cost_by_service"`
	CostByRegion      map[string]float64 `json:"cost_by_region"`
	CostForecast      float64            `json:"cost_forecast"`
}
