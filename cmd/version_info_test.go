package cmd

import (
	"encoding/json"
	"testing"
)

func TestBuildVersionInfoNonEmpty(t *testing.T) {
	v := buildVersionInfo()
	if v.Version == "" {
		t.Fatalf("Version empty")
	}
	if v.Commit == "" {
		t.Fatalf("Commit empty")
	}
	if v.BuildDate == "" {
		t.Fatalf("BuildDate empty")
	}
	if v.GoVersion == "" {
		t.Fatalf("GoVersion empty")
	}
}

func TestMarshalVersionJSON(t *testing.T) {
	v := buildVersionInfo()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, k := range []string{"version", "commit", "build_date", "go_version"} {
		if _, ok := m[k]; !ok {
			t.Errorf("missing key %s", k)
		}
	}
}
