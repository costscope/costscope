package integration

import "testing"

func TestDSLValidate_OK(t *testing.T) {
	dsl := &ActionDSL{
		Version: "v1",
		Actions: []DSLAction{
			{ID: "t.a", Category: "webhook", Use: "create", Short: "x", InputsSchema: map[string]any{"type": "object"}, Flags: []DSLFlag{{Name: "name", Type: "string", Usage: "n"}}},
			{ID: "t.grp", Category: "dashboard", Use: "config", Short: "g", Group: true},
		},
	}
	if err := dsl.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDSLValidate_Errors(t *testing.T) {
	dsl := &ActionDSL{
		Version: "",
		Actions: []DSLAction{
			{ID: "", Category: "", Use: "", Short: ""},                                                                              // missing id/category/use
			{ID: "dup", Category: "webhook", Use: "list", Short: "a"},                                                               // ok
			{ID: "dup", Category: "webhook", Use: "list", Short: "b"},                                                               // duplicate id
			{ID: "grpBad", Category: "dashboard", Use: "x", Short: "g", Group: true, Flags: []DSLFlag{{Name: "f", Type: "string"}}}, // group with flags
			{ID: "flagBad", Category: "webhook", Use: "x", Short: "s", Flags: []DSLFlag{{Name: "z", Type: "unsupported"}}},          // bad flag type
			{ID: "shortBad", Category: "webhook", Use: "x", Short: "s", Flags: []DSLFlag{{Name: "a", Type: "string", Shorthand: "ab"}}},
		},
	}
	if err := dsl.Validate(); err == nil {
		t.Fatalf("expected validation error, got nil")
	}
}
