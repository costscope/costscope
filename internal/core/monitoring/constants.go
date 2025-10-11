package monitoring

// Notification status constants
const (
	StatusSent    = "sent"
	StatusPending = "pending"
	StatusFailed  = "failed"
)

// Severity constants
const (
	SeverityCritical = "critical"
	SeverityWarning  = "warning"
	SeverityInfo     = "info"
)

// Health status constants
const (
	HealthyStatus  = "healthy"
	DegradedStatus = "degraded"
	CriticalStatus = "critical"
	UnknownStatus  = "unknown"
)

// Alert status constants
const (
	AlertStatusActive   = "active"
	AlertStatusResolved = "resolved"
)

// UI symbols
const (
	SymbolHealthy   = ""
	SymbolUnhealthy = ""
)

// Output format constants
const (
	FormatTable = "table"
	FormatJSON  = "json"
)
