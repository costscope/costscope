package mapping

import (
	"reflect"
	"testing"
	"time"
)

// sampleStruct used for struct extraction tests
type sampleStruct struct {
	ID        int    `json:"id"`
	Name      string `csv:"name"`
	Value     *float64
	Flag      bool
	Count     uint32
	CreatedAt time.Time
	Nested    struct{ Label string }
}

func TestUniversalValueExtractor_CSVRowExtraction(t *testing.T) {
	uve := NewUniversalValueExtractor()
	src := CSVRowSource{Header: []string{"Region", "b"}, Row: []string{"us-east-1", "valB"}}

	// Case-insensitive header match
	v, ok, err := uve.ExtractString(src, "region")
	if err != nil || !ok || v != "us-east-1" {
		t.Fatalf("expected region value, got %v %v %v", v, ok, err)
	}

	// Field not present
	if _, ok, err := uve.ExtractString(src, "missing"); err != nil || ok {
		t.Fatalf("expected missing field no error")
	}

	// Row shorter than header triggers error
	short := CSVRowSource{Header: []string{"A", "B", "C"}, Row: []string{"1", "2"}}
	if _, _, err := uve.ExtractString(short, "C"); err == nil {
		t.Fatalf("expected short row error")
	}
}

func TestUniversalValueExtractor_StructExtractionVariants(t *testing.T) {
	uve := NewUniversalValueExtractor()
	f := 42.5
	s := &sampleStruct{ID: 7, Name: "alpha", Value: &f, Flag: true, Count: 9, CreatedAt: time.Unix(100, 0)}
	s.Nested.Label = "inner"

	// Direct field name
	if v, ok, err := uve.ExtractString(s, "Flag"); err != nil || !ok || v != "true" {
		t.Fatalf("flag extract failed: %v %v %v", v, ok, err)
	}
	// json tag
	if v, ok, err := uve.ExtractString(s, "id"); err != nil || !ok || v != "7" {
		t.Fatalf("json tag extract failed: %v %v %v", v, ok, err)
	}
	// csv tag
	if v, ok, err := uve.ExtractString(s, "name"); err != nil || !ok || v != "alpha" {
		t.Fatalf("csv tag extract failed: %v %v %v", v, ok, err)
	}
	// case-insensitive search
	if v, ok, err := uve.ExtractString(s, "vAlUe"); err != nil || !ok || v != "42.5" {
		t.Fatalf("case-insensitive extract failed: %v %v %v", v, ok, err)
	}
	// pointer nil path exists=true returns empty string
	s.Value = nil
	if v, ok, err := uve.ExtractString(s, "value"); err != nil || !ok || v != "" {
		t.Fatalf("nil pointer extract failed: %v %v %v", v, ok, err)
	}
	// unknown field
	if v, ok, err := uve.ExtractString(s, "doesNotExist"); err != nil || ok || v != "" {
		t.Fatalf("expected not found: %v %v %v", v, ok, err)
	}
	// non-struct error
	if _, _, err := uve.ExtractString(5, "X"); err == nil {
		t.Fatalf("expected non-struct error")
	}

	// GetAvailableFields (struct & maps & CSV)
	fields := uve.GetAvailableFields(s)
	if len(fields) == 0 {
		t.Fatalf("expected struct fields")
	}
	if uve.GetAvailableFields(map[string]interface{}{"a": 1, "b": 2}) == nil {
		t.Fatalf("expected map keys slice")
	}
	if uve.GetAvailableFields(map[string]string{"x": "y"}) == nil {
		t.Fatalf("expected string map keys slice")
	}
	if uve.GetAvailableFields(CSVRowSource{Header: []string{"H1"}})[0] != "H1" {
		t.Fatalf("expected header field")
	}
}

func TestUniversalValueExtractor_convertFieldFallbacks(t *testing.T) {
	uve := NewUniversalValueExtractor()
	f := 1.23
	st := &sampleStruct{Value: &f}
	// convertFieldToString exercised indirectly above; here ensure numeric unsigned & nested struct default path
	st.Count = 17
	st.Nested.Label = "n" // nested struct should format using fmt (default branch)
	if v, ok, err := uve.ExtractString(st, "Count"); err != nil || !ok || v != "17" {
		t.Fatalf("uint extract failed")
	}
	if v, ok, err := uve.ExtractString(st, "Nested"); err != nil || !ok || v == "" {
		t.Fatalf("nested struct stringify failed: %v / %v / %v", v, ok, err)
	}
}

func Test_getMapKeys_GenericHelper(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	keys := getMapKeys(m)
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys")
	}
}

func Test_getStructFields_UnexportedIgnored(t *testing.T) {
	uve := NewUniversalValueExtractor()
	fields := uve.getStructFields(sampleStruct{})
	for _, f := range fields {
		if f == "unexported" {
			t.Fatalf("unexported field should be ignored")
		}
	}
	// non-struct returns nil
	if uve.getStructFields(5) != nil {
		t.Fatalf("expected nil for non-struct")
	}
}

// Guard against accidental changes to reflection logic: ensure exported count matches struct definition minus unexported.
func Test_getStructFields_Count(t *testing.T) {
	uve := NewUniversalValueExtractor()
	fields := uve.getStructFields(sampleStruct{})
	typ := reflect.TypeOf(sampleStruct{})
	total := 0
	for i := 0; i < typ.NumField(); i++ {
		if typ.Field(i).IsExported() {
			total++
		}
	}
	if len(fields) != total {
		t.Fatalf("field count mismatch got %d want %d", len(fields), total)
	}
}
