package mapping

import "testing"

func TestUniversalTypeConverter_Table_StringAndEnums(t *testing.T) {
	utc := NewUniversalTypeConverter()
	strCases := []struct {
		in   string
		tr   string
		want string
	}{
		{"  xYz  ", "trim", "xYz"},
		{"MiXeD", "lowercase", "mixed"},
		{"  TWO  Words\t\t", "normalize_whitespace", "TWO Words"},
		{"/api", "remove_prefix_slash", "api"},
		{"api/", "remove_suffix_slash", "api"},
	}
	for _, tc := range strCases {
		out, err := utc.ConvertString(tc.in, FieldMapping{Transform: tc.tr})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if out.(string) != tc.want {
			t.Fatalf("transform %s: got %q want %q", tc.tr, out, tc.want)
		}
	}

	// Enums (case-insensitive mapping)
	emap := map[string]string{"discount": "Discount", "CREDIT": "Credit"}
	mapped, err := utc.ConvertEnum("DiScOuNt", emap)
	if err != nil || mapped != "Discount" {
		t.Fatalf("enum map 1 failed: %v %q", err, mapped)
	}
	mapped, err = utc.ConvertEnum("CREDIT", emap)
	if err != nil || mapped != "Credit" {
		t.Fatalf("enum map 2 failed: %v %q", err, mapped)
	}
}
