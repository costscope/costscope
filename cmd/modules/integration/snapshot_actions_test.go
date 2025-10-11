package integration

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type snapFlag struct {
	Name     string
	Usage    string
	Default  string
	Required bool
}

type snapAction struct {
	Use   string
	Short string
	Long  string
	Flags []snapFlag
}

func buildSnapshot(a *cobra.Command) snapAction {
	af := snapAction{Use: a.Use, Short: a.Short, Long: normalize(a.Long)}
	a.Flags().VisitAll(func(f *pflag.Flag) {
		req := false
		if ann, ok := f.Annotations["cobra_annotation_required_flag"]; ok && len(ann) > 0 && ann[0] == "true" {
			req = true
		}
		af.Flags = append(af.Flags, snapFlag{Name: f.Name, Usage: f.Usage, Default: f.DefValue, Required: req})
	})
	return af
}

func normalize(s string) string {
	var out []rune
	prevSpace := false
	for _, r := range s {
		if r == '\n' || r == '\t' || r == ' ' {
			if !prevSpace {
				out = append(out, ' ')
			}
			prevSpace = true
		} else {
			out = append(out, r)
			prevSpace = false
		}
	}
	return string(out)
}

func expectedSnapshot() map[string]snapAction {
	return map[string]snapAction{
		"webhook.create":             {Use: "create", Short: "Create a new webhook with advanced features", Long: normalize("Create a new webhook with support for custom headers, security, retry policies, and event filtering."), Flags: []snapFlag{{Name: "name", Usage: "Webhook name", Required: true}, {Name: "url", Usage: "Webhook URL", Required: true}, {Name: "events", Usage: "Events to subscribe to", Default: "[*]"}, {Name: "headers", Usage: "Custom HTTP headers"}, {Name: "secret", Usage: "Secret for signature verification"}, {Name: "max-retries", Usage: "Maximum retry attempts", Default: "3"}, {Name: "timeout", Usage: "Request timeout", Default: (30 * time.Second).String()}}},
		"webhook.list":               {Use: "list", Short: "List all webhooks", Long: normalize("List all configured webhooks with their status and recent delivery statistics."), Flags: []snapFlag{{Name: "status", Usage: "Filter by status (active, error, disabled)"}, {Name: "verbose", Usage: "Show detailed information", Default: "false"}}},
		"webhook.test":               {Use: "test", Short: "Test webhook delivery", Long: normalize("Send a test event to a webhook to verify configuration and connectivity."), Flags: []snapFlag{{Name: "webhook-id", Usage: "Webhook ID to test", Required: true}, {Name: "event", Usage: "Event type to send", Default: "test.webhook"}, {Name: "payload", Usage: "Custom JSON payload"}}},
		"webhook.delete":             {Use: "delete", Short: "Delete a webhook", Long: normalize("Delete a webhook and its delivery history."), Flags: []snapFlag{{Name: "webhook-id", Usage: "Webhook ID to delete", Required: true}, {Name: "confirm", Usage: "Skip confirmation prompt", Default: "false"}}},
		"webhook.delivery.list":      {Use: "list", Short: "List webhook deliveries"},
		"webhook.delivery.retry":     {Use: "retry", Short: "Retry failed delivery"},
		"webhook.delivery.stats":     {Use: "stats", Short: "Show delivery statistics"},
		"webhook.event.list":         {Use: "list", Short: "List available webhook events"},
		"webhook.event.trigger":      {Use: "trigger", Short: "Trigger a test event"},
		"dashboard.start":            {Use: "start", Short: "Start enhanced dashboard server", Long: normalize("Start the enhanced dashboard server with customizable features and security options."), Flags: []snapFlag{{Name: "port", Usage: "Dashboard port", Default: "8080"}, {Name: "theme", Usage: "Dashboard theme (modern, dark, light, corporate)", Default: "modern"}, {Name: "auto-open", Usage: "Automatically open dashboard in browser", Default: "true"}, {Name: "features", Usage: "Enable specific features", Default: "[real-time,interactive,mobile]"}, {Name: "layout", Usage: "Dashboard layout (default, grid, compact, executive)", Default: "default"}, {Name: "auth", Usage: "Enable authentication", Default: "false"}, {Name: "allowed-ips", Usage: "Allowed IP addresses/ranges (CIDR notation)"}}},
		"dashboard.status":           {Use: "status", Short: "Show dashboard status and metrics", Long: normalize("Display dashboard status, performance metrics, and configuration details."), Flags: []snapFlag{{Name: "verbose", Usage: "Show detailed status information", Default: "false"}}},
		"dashboard.stop":             {Use: "stop", Short: "Stop dashboard server", Long: normalize("Gracefully stop the dashboard server and save current state.")},
		"dashboard.config.show":      {Use: "show", Short: "Show current configuration"},
		"dashboard.config.set":       {Use: "set", Short: "Set configuration value"},
		"dashboard.config.reset":     {Use: "reset", Short: "Reset configuration to defaults"},
		"dashboard.widget.add":       {Use: "add", Short: "Add new widget"},
		"dashboard.widget.list":      {Use: "list", Short: "List dashboard widgets"},
		"dashboard.widget.remove":    {Use: "remove", Short: "Remove widget"},
		"dashboard.widget.configure": {Use: "configure", Short: "Configure widget"},
		"dashboard.plugin.install":   {Use: "install", Short: "Install plugin"},
		"dashboard.plugin.list":      {Use: "list", Short: "List plugins"},
		"dashboard.plugin.enable":    {Use: "enable", Short: "Enable plugin"},
		"dashboard.plugin.disable":   {Use: "disable", Short: "Disable plugin"},
		"connections.connect":        {Use: "connect", Short: "Connect to third-party systems with advanced management", Long: normalize("Integrate with billing systems, ITSM tools, and BI platforms with health monitoring and retry."), Flags: []snapFlag{{Name: "system", Usage: "system to connect to"}, {Name: "list", Usage: "list available integrations", Default: "false"}, {Name: "category", Usage: "filter by integration category"}, {Name: "config", Usage: "connection configuration (key=value pairs)"}, {Name: "health-check", Usage: "enable health monitoring for the connection", Default: "true"}, {Name: "timeout", Usage: "connection timeout", Default: (30 * time.Second).String()}, {Name: "auto-retry", Usage: "enable automatic retry on connection failures", Default: "true"}}},
	}
}

func TestActionSpecsSnapshot(t *testing.T) {
	root := CreateIntegrationCommands()
	exp := expectedSnapshot()
	lookups := map[string][]string{
		"webhook.create":             {"webhook", "create"},
		"webhook.list":               {"webhook", "list"},
		"webhook.test":               {"webhook", "test"},
		"webhook.delete":             {"webhook", "delete"},
		"webhook.delivery.list":      {"webhook", "delivery", "list"},
		"webhook.delivery.retry":     {"webhook", "delivery", "retry"},
		"webhook.delivery.stats":     {"webhook", "delivery", "stats"},
		"webhook.event.list":         {"webhook", "event", "list"},
		"webhook.event.trigger":      {"webhook", "event", "trigger"},
		"dashboard.start":            {"dashboard", "start"},
		"dashboard.status":           {"dashboard", "status"},
		"dashboard.stop":             {"dashboard", "stop"},
		"dashboard.config.show":      {"dashboard", "config", "show"},
		"dashboard.config.set":       {"dashboard", "config", "set"},
		"dashboard.config.reset":     {"dashboard", "config", "reset"},
		"dashboard.widget.add":       {"dashboard", "widget", "add"},
		"dashboard.widget.list":      {"dashboard", "widget", "list"},
		"dashboard.widget.remove":    {"dashboard", "widget", "remove"},
		"dashboard.widget.configure": {"dashboard", "widget", "configure"},
		"dashboard.plugin.install":   {"dashboard", "plugin", "install"},
		"dashboard.plugin.list":      {"dashboard", "plugin", "list"},
		"dashboard.plugin.enable":    {"dashboard", "plugin", "enable"},
		"dashboard.plugin.disable":   {"dashboard", "plugin", "disable"},
		"connections.connect":        {"connections", "connect"},
	}
	find := func(path ...string) *cobra.Command {
		cur := root
		for _, p := range path {
			var next *cobra.Command
			for _, c := range cur.Commands() {
				if c.Use == p {
					next = c
					break
				}
			}
			if next == nil {
				t.Fatalf("missing command segment %s", p)
			}
			cur = next
		}
		return cur
	}
	for id, chain := range lookups {
		cmd := find(chain...)
		snap := buildSnapshot(cmd)
		want := exp[id]
		if snap.Use != want.Use || snap.Short != want.Short || normalize(snap.Long) != want.Long {
			t.Fatalf("snapshot mismatch %s metadata changed", id)
		}
		gotFlags := map[string]snapFlag{}
		for _, f := range snap.Flags {
			gotFlags[f.Name] = f
		}
		if len(gotFlags) != len(want.Flags) {
			t.Fatalf("snapshot %s flag count changed got=%d want=%d", id, len(gotFlags), len(want.Flags))
		}
		for _, wf := range want.Flags {
			gf, ok := gotFlags[wf.Name]
			if !ok {
				t.Fatalf("snapshot %s missing flag %s", id, wf.Name)
			}
			if gf.Usage != wf.Usage {
				t.Fatalf("snapshot %s flag %s usage changed got=%q want=%q", id, wf.Name, gf.Usage, wf.Usage)
			}
		}
	}
}
