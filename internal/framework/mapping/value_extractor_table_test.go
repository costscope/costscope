package mapping

import "testing"

func TestUniversalValueExtractor_ParseErrors_Table(t *testing.T) {
	uve := NewUniversalValueExtractor()
	src := map[string]string{
		"badFloat": "not-a-float",
		"badInt":   "12.3",
		"badBool":  "maybe",
		"empty":    "",
		"okFloat":  "3.14",
		"okInt":    "42",
		"okBoolT":  "true",
		"okBoolF":  "no",
	}

	// Float
	if _, ok, err := uve.ExtractFloat(src, "empty"); ok || err != nil {
		t.Fatalf("empty float: ok=%v err=%v", ok, err)
	}
	if _, ok, err := uve.ExtractFloat(src, "badFloat"); !ok || err == nil {
		t.Fatalf("bad float should error; ok=%v err=%v", ok, err)
	}
	if v, ok, err := uve.ExtractFloat(src, "okFloat"); !ok || err != nil || v != 3.14 {
		t.Fatalf("ok float got v=%v ok=%v err=%v", v, ok, err)
	}

	// Int
	if _, ok, err := uve.ExtractInt(src, "empty"); ok || err != nil {
		t.Fatalf("empty int: ok=%v err=%v", ok, err)
	}
	if _, ok, err := uve.ExtractInt(src, "badInt"); !ok || err == nil {
		t.Fatalf("bad int should error; ok=%v err=%v", ok, err)
	}
	if v, ok, err := uve.ExtractInt(src, "okInt"); !ok || err != nil || v != 42 {
		t.Fatalf("ok int got v=%v ok=%v err=%v", v, ok, err)
	}

	// Bool
	if _, ok, err := uve.ExtractBool(src, "empty"); ok || err != nil {
		t.Fatalf("empty bool: ok=%v err=%v", ok, err)
	}
	if _, ok, err := uve.ExtractBool(src, "badBool"); !ok || err == nil {
		t.Fatalf("bad bool should error; ok=%v err=%v", ok, err)
	}
	if v, ok, err := uve.ExtractBool(src, "okBoolT"); !ok || err != nil || v != true {
		t.Fatalf("ok bool true got v=%v ok=%v err=%v", v, ok, err)
	}
	if v, ok, err := uve.ExtractBool(src, "okBoolF"); !ok || err != nil || v != false {
		t.Fatalf("ok bool false got v=%v ok=%v err=%v", v, ok, err)
	}
}
