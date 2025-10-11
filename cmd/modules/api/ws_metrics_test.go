package api

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestMetricsEndpoint(t *testing.T) {
	srv := newTestServer()
	rr := doReq(t, srv, "GET", "/metrics", nil)
	if rr.Code != 200 {
		t.Fatalf("metrics expected 200, got %d", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Fatalf("unexpected metrics content-type: %s", ct)
	}
	if !strings.Contains(rr.Body.String(), "# HELP") {
		t.Fatalf("metrics body looks wrong, got: %s", rr.Body.String()[:min(60, len(rr.Body.String()))])
	}
}

func TestWebSocketJobsUpgrade(t *testing.T) {
	// Start an HTTP test server for upgrade
	handler := newTestServer()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	wsURL := url.URL{Scheme: "ws", Host: u.Host, Path: "/ws/jobs/abc123"}
	conn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		t.Fatalf("websocket dial failed: %v", err)
	}
	defer func() { _ = conn.Close() }()

	// Expect a welcome message
	type welcome struct {
		Type  string `json:"type"`
		JobID string `json:"job_id"`
	}
	var msg welcome
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("read welcome failed: %v", err)
	}
	if msg.Type != "welcome" || msg.JobID != "abc123" {
		t.Fatalf("unexpected welcome payload: %+v", msg)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
