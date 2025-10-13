package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/costscope/costscope/internal/core/logging"
)

func newTestService() *BasicNotificationService {
	logger := logging.NewLogger(logging.LevelDebug)
	return NewBasicNotificationService(logger)
}

func mkNotif(channel string) *Notification {
	return &Notification{
		ID:        "n1",
		Type:      "test",
		Channel:   channel,
		Recipient: "r@test",
		Subject:   "subj",
		Message:   "msg",
		Severity:  SeverityInfo,
	}
}

func TestSendNotification_RoutesToRegisteredHandlers(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	// Only channels enabled by default should succeed
	tests := []struct {
		channel string
		wantErr bool
	}{
		{"email", false},
		{"slack", false},
		{"teams", false},
	}

	for _, tt := range tests {
		n := mkNotif(tt.channel)
		err := svc.SendNotification(ctx, n)
		if (err != nil) != tt.wantErr {
			t.Fatalf("channel %s: unexpected error state: %v", tt.channel, err)
		}
		if tt.wantErr {
			continue
		}
		if n.Status != StatusSent {
			t.Fatalf("channel %s: expected status %q, got %q", tt.channel, StatusSent, n.Status)
		}
		if n.SentAt == nil || n.SentAt.IsZero() {
			t.Fatalf("channel %s: SentAt not set", tt.channel)
		}
		if n.CreatedAt.IsZero() {
			t.Fatalf("channel %s: CreatedAt should be set by dispatcher", tt.channel)
		}
	}
}

func TestSendNotification_DisabledAndUnsupportedChannels(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	// Disabled by default
	if err := svc.SendNotification(ctx, mkNotif("discord")); err == nil {
		t.Fatalf("expected error for disabled channel 'discord'")
	}

	// Unsupported channel
	if err := svc.SendNotification(ctx, mkNotif("unknown")); err == nil {
		t.Fatalf("expected error for unsupported channel 'unknown'")
	}
}

func TestSendNotification_MissingHandler(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	// Use an enabled channel (teams), then remove its handler to simulate missing handler
	delete(svc.handlers, "teams")

	// Ensure channel config remains enabled
	cfg := svc.channelConfigurations["teams"]
	cfg.Enabled = true
	svc.channelConfigurations["teams"] = cfg

	n := mkNotif("teams")
	if err := svc.SendNotification(ctx, n); err == nil {
		t.Fatalf("expected error when handler missing for 'teams'")
	}
	if n.Status == StatusSent || (n.SentAt != nil && !n.SentAt.IsZero()) {
		t.Fatalf("notification should not be marked sent on handler error")
	}
}

func TestSendNotification_ContextCancellation(t *testing.T) {
	svc := newTestService()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	n := mkNotif("email")
	err := svc.SendNotification(ctx, n)
	if err == nil {
		t.Fatalf("expected context cancellation error")
	}
	// CreatedAt is set before delivery attempt
	if n.CreatedAt.IsZero() {
		t.Fatalf("CreatedAt should be set even on failure")
	}
	if n.SentAt != nil && !n.SentAt.IsZero() {
		t.Fatalf("SentAt should not be set on failure")
	}
	if n.Status == StatusSent {
		t.Fatalf("Status should not be %q on failure", StatusSent)
	}
}

func TestSendAlert_MixedChannelsAggregatesErrors(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	alert := &Alert{
		ID:          "a1",
		Type:        "threshold",
		Severity:    SeverityWarning,
		Source:      "unit-test",
		Component:   "comp",
		Title:       "Test Alert",
		Description: "desc",
		CreatedAt:   time.Now(),
	}

	// One success (email), one failure (disabled discord)
	err := svc.SendAlert(ctx, alert, []string{"email", "discord"})
	if err == nil {
		t.Fatalf("expected aggregated error when at least one channel fails")
	}
}
