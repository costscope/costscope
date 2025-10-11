package monitoring

import "time"

// Alert and notification related types split from types.go for modularity

type Alert struct {
	ID                string            `json:"id"`
	Type              string            `json:"type"`
	Severity          string            `json:"severity"`
	Source            string            `json:"source"`
	Component         string            `json:"component"`
	Title             string            `json:"title"`
	Description       string            `json:"description"`
	Status            string            `json:"status"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	ResolvedAt        *time.Time        `json:"resolved_at,omitempty"`
	AcknowledgedAt    *time.Time        `json:"acknowledged_at,omitempty"`
	AcknowledgedBy    string            `json:"acknowledged_by,omitempty"`
	Value             float64           `json:"value"`
	Threshold         float64           `json:"threshold"`
	Tags              map[string]string `json:"tags"`
	Annotations       map[string]string `json:"annotations"`
	NotificationsSent int               `json:"notifications_sent"`
	EscalationLevel   int               `json:"escalation_level"`
}

type AlertDefinition struct {
	Name                 string            `json:"name"`
	Description          string            `json:"description"`
	MetricName           string            `json:"metric_name"`
	Condition            string            `json:"condition"`
	Threshold            float64           `json:"threshold"`
	Severity             string            `json:"severity"`
	NotificationChannels []string          `json:"notification_channels"`
	Tags                 map[string]string `json:"tags"`
	Enabled              bool              `json:"enabled"`
}

type AlertRule struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	Description          string            `json:"description"`
	Metric               string            `json:"metric"`
	Operator             string            `json:"operator"`
	Threshold            float64           `json:"threshold"`
	Duration             time.Duration     `json:"duration"`
	Severity             string            `json:"severity"`
	NotificationChannels []string          `json:"notification_channels"`
	Tags                 map[string]string `json:"tags"`
	Enabled              bool              `json:"enabled"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
}
