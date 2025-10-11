package exporters

// goldenFixture contains the canonical JSON for TestJSONExporter_Golden.
// Kept inline to avoid external FS dependency during containerized runs (act).
var goldenFixture = []byte(`{
	"id": "golden-1",
	"title": "Golden Cost Analysis",
	"description": "Fixture",
	"generated_at": "2000-01-01T00:00:00Z",
	"date_range": {
		"start_date": "2000-01-01T00:00:00Z",
		"end_date": "2000-01-31T00:00:00Z"
	},
	"total_cost": 100.5,
	"currency": "USD",
	"cost_by_service": [],
	"cost_by_region": [],
	"cost_by_account": [],
	"cost_trends": [],
	"top_cost_drivers": [],
	"optimization_recommendations": [],
	"summary": {
		"total_items": 0,
		"total_cost": 100.5,
		"currency": "USD",
		"date_range": {
			"start_date": "2000-01-01T00:00:00Z",
			"end_date": "2000-01-31T00:00:00Z"
		}
	}
}`)
