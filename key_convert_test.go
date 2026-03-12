package stanza

import (
	"math"
	"testing"
	"time"
)

func testKey(value string) *Key {
	return newKey(nil, "test", value)
}

// ---------------------------------------------------------------------------
// parseBool
// ---------------------------------------------------------------------------

func TestParseBool(t *testing.T) {
	truthy := []string{"1", "t", "true", "yes", "y", "on", "TRUE", "Yes", " true "}
	for _, s := range truthy {
		v, err := parseBool(s)
		if err != nil || !v {
			t.Errorf("parseBool(%q) = %v, %v; want true, nil", s, v, err)
		}
	}
	falsy := []string{"0", "f", "false", "no", "n", "off", "FALSE", "No", " off "}
	for _, s := range falsy {
		v, err := parseBool(s)
		if err != nil || v {
			t.Errorf("parseBool(%q) = %v, %v; want false, nil", s, v, err)
		}
	}
	_, err := parseBool("maybe")
	if err == nil {
		t.Error("expected error for invalid bool")
	}
}

// ---------------------------------------------------------------------------
// Basic conversions
// ---------------------------------------------------------------------------

func TestBool(t *testing.T) {
	v, err := testKey("true").Bool()
	if err != nil || !v {
		t.Errorf("Bool() = %v, %v", v, err)
	}
	v, err = testKey("false").Bool()
	if err != nil || v {
		t.Errorf("Bool(false) = %v, %v", v, err)
	}
	_, err = testKey("nope").Bool()
	if err == nil {
		t.Error("expected error")
	}
}

func TestInt(t *testing.T) {
	v, err := testKey("42").Int()
	if err != nil || v != 42 {
		t.Errorf("Int() = %d, %v", v, err)
	}
	v, err = testKey("-7").Int()
	if err != nil || v != -7 {
		t.Errorf("Int(-7) = %d, %v", v, err)
	}
	_, err = testKey("abc").Int()
	if err == nil {
		t.Error("expected error")
	}
}

func TestInt64(t *testing.T) {
	v, err := testKey("9223372036854775807").Int64()
	if err != nil || v != math.MaxInt64 {
		t.Errorf("Int64() = %d, %v", v, err)
	}
}

func TestUint(t *testing.T) {
	v, err := testKey("100").Uint()
	if err != nil || v != 100 {
		t.Errorf("Uint() = %d, %v", v, err)
	}
	_, err = testKey("-1").Uint()
	if err == nil {
		t.Error("expected error for negative uint")
	}
}

func TestUint64(t *testing.T) {
	v, err := testKey("18446744073709551615").Uint64()
	if err != nil || v != math.MaxUint64 {
		t.Errorf("Uint64() = %d, %v", v, err)
	}
}

func TestFloat64(t *testing.T) {
	v, err := testKey("3.14").Float64()
	if err != nil || v != 3.14 {
		t.Errorf("Float64() = %f, %v", v, err)
	}
	_, err = testKey("not_a_float").Float64()
	if err == nil {
		t.Error("expected error")
	}
}

func TestDuration(t *testing.T) {
	v, err := testKey("5s").Duration()
	if err != nil || v != 5*time.Second {
		t.Errorf("Duration() = %v, %v", v, err)
	}
	v, err = testKey("1h30m").Duration()
	if err != nil || v != 90*time.Minute {
		t.Errorf("Duration(1h30m) = %v, %v", v, err)
	}
	_, err = testKey("nope").Duration()
	if err == nil {
		t.Error("expected error")
	}
}

func TestTime(t *testing.T) {
	v, err := testKey("2026-01-15T10:30:00Z").Time()
	if err != nil {
		t.Fatal(err)
	}
	if v.Year() != 2026 || v.Month() != 1 || v.Day() != 15 {
		t.Errorf("Time() = %v", v)
	}
	_, err = testKey("not-a-date").Time()
	if err == nil {
		t.Error("expected error")
	}
}

func TestTimeFormat(t *testing.T) {
	v, err := testKey("15/01/2026").TimeFormat("02/01/2006")
	if err != nil {
		t.Fatal(err)
	}
	if v.Day() != 15 || v.Month() != 1 || v.Year() != 2026 {
		t.Errorf("TimeFormat() = %v", v)
	}
}

// ---------------------------------------------------------------------------
// Must methods
// ---------------------------------------------------------------------------

func TestMustString(t *testing.T) {
	if v := testKey("hello").MustString("def"); v != "hello" {
		t.Errorf("got %q", v)
	}
	if v := testKey("").MustString("def"); v != "def" {
		t.Errorf("got %q, want def", v)
	}
}

func TestMustBool(t *testing.T) {
	if v := testKey("true").MustBool(); !v {
		t.Error("expected true")
	}
	if v := testKey("bad").MustBool(); v {
		t.Error("expected false default")
	}
	if v := testKey("bad").MustBool(true); !v {
		t.Error("expected true default")
	}
}

func TestMustInt(t *testing.T) {
	if v := testKey("42").MustInt(); v != 42 {
		t.Errorf("got %d", v)
	}
	if v := testKey("bad").MustInt(); v != 0 {
		t.Errorf("got %d, want 0", v)
	}
	if v := testKey("bad").MustInt(99); v != 99 {
		t.Errorf("got %d, want 99", v)
	}
}

func TestMustInt64(t *testing.T) {
	if v := testKey("100").MustInt64(); v != 100 {
		t.Errorf("got %d", v)
	}
	if v := testKey("bad").MustInt64(int64(7)); v != 7 {
		t.Errorf("got %d", v)
	}
}

func TestMustUint(t *testing.T) {
	if v := testKey("50").MustUint(); v != 50 {
		t.Errorf("got %d", v)
	}
	if v := testKey("bad").MustUint(uint(3)); v != 3 {
		t.Errorf("got %d", v)
	}
}

func TestMustUint64(t *testing.T) {
	if v := testKey("200").MustUint64(); v != 200 {
		t.Errorf("got %d", v)
	}
	if v := testKey("bad").MustUint64(uint64(5)); v != 5 {
		t.Errorf("got %d", v)
	}
}

func TestMustFloat64(t *testing.T) {
	if v := testKey("2.5").MustFloat64(); v != 2.5 {
		t.Errorf("got %f", v)
	}
	if v := testKey("bad").MustFloat64(1.1); v != 1.1 {
		t.Errorf("got %f", v)
	}
}

func TestMustDuration(t *testing.T) {
	if v := testKey("10s").MustDuration(); v != 10*time.Second {
		t.Errorf("got %v", v)
	}
	if v := testKey("bad").MustDuration(3 * time.Second); v != 3*time.Second {
		t.Errorf("got %v", v)
	}
}

func TestMustTime(t *testing.T) {
	v := testKey("2026-01-15T10:30:00Z").MustTime()
	if v.Year() != 2026 {
		t.Errorf("got %v", v)
	}
	def := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	if v := testKey("bad").MustTime(def); !v.Equal(def) {
		t.Errorf("got %v, want %v", v, def)
	}
	if v := testKey("bad").MustTime(); !v.IsZero() {
		t.Errorf("expected zero time, got %v", v)
	}
}

func TestMustTimeFormat(t *testing.T) {
	v := testKey("15/01/2026").MustTimeFormat("02/01/2006")
	if v.Day() != 15 {
		t.Errorf("got %v", v)
	}
	def := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	if v := testKey("bad").MustTimeFormat("02/01/2006", def); !v.Equal(def) {
		t.Errorf("got %v", v)
	}
}

// ---------------------------------------------------------------------------
// In methods
// ---------------------------------------------------------------------------

func TestIn(t *testing.T) {
	k := testKey("b")
	if v := k.In("def", []string{"a", "b", "c"}); v != "b" {
		t.Errorf("got %q", v)
	}
	k = testKey("z")
	if v := k.In("def", []string{"a", "b", "c"}); v != "def" {
		t.Errorf("got %q, want def", v)
	}
}

func TestInInt(t *testing.T) {
	if v := testKey("2").InInt(0, []int{1, 2, 3}); v != 2 {
		t.Errorf("got %d", v)
	}
	if v := testKey("9").InInt(0, []int{1, 2, 3}); v != 0 {
		t.Errorf("got %d", v)
	}
}

func TestInInt64(t *testing.T) {
	if v := testKey("2").InInt64(0, []int64{1, 2, 3}); v != 2 {
		t.Errorf("got %d", v)
	}
	if v := testKey("9").InInt64(0, []int64{1, 2, 3}); v != 0 {
		t.Errorf("got %d", v)
	}
}

func TestInUint(t *testing.T) {
	if v := testKey("2").InUint(0, []uint{1, 2, 3}); v != 2 {
		t.Errorf("got %d", v)
	}
	if v := testKey("9").InUint(0, []uint{1, 2, 3}); v != 0 {
		t.Errorf("got %d", v)
	}
}

func TestInUint64(t *testing.T) {
	if v := testKey("2").InUint64(0, []uint64{1, 2, 3}); v != 2 {
		t.Errorf("got %d", v)
	}
}

func TestInFloat64(t *testing.T) {
	if v := testKey("2.5").InFloat64(0, []float64{1.5, 2.5, 3.5}); v != 2.5 {
		t.Errorf("got %f", v)
	}
	if v := testKey("9.9").InFloat64(0, []float64{1.5, 2.5}); v != 0 {
		t.Errorf("got %f", v)
	}
}

func TestInTime(t *testing.T) {
	t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	def := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	k := testKey("2026-01-01T00:00:00Z")
	if v := k.InTime(def, []time.Time{t1, t2}); !v.Equal(t1) {
		t.Errorf("got %v", v)
	}
	k = testKey("2099-01-01T00:00:00Z")
	if v := k.InTime(def, []time.Time{t1, t2}); !v.Equal(def) {
		t.Errorf("got %v, want default", v)
	}
}

func TestInTimeFormat(t *testing.T) {
	t1 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	def := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	k := testKey("15/01/2026")
	if v := k.InTimeFormat("02/01/2006", def, []time.Time{t1}); !v.Equal(t1) {
		t.Errorf("got %v", v)
	}
}

// ---------------------------------------------------------------------------
// Range methods
// ---------------------------------------------------------------------------

func TestRangeInt(t *testing.T) {
	if v := testKey("5").RangeInt(0, 1, 10); v != 5 {
		t.Errorf("got %d", v)
	}
	if v := testKey("20").RangeInt(0, 1, 10); v != 0 {
		t.Errorf("got %d, want 0 (out of range)", v)
	}
	if v := testKey("-5").RangeInt(0, 1, 10); v != 0 {
		t.Errorf("got %d, want 0 (below range)", v)
	}
}

func TestRangeInt64(t *testing.T) {
	if v := testKey("5").RangeInt64(0, 1, 10); v != 5 {
		t.Errorf("got %d", v)
	}
	if v := testKey("20").RangeInt64(0, 1, 10); v != 0 {
		t.Errorf("got %d", v)
	}
}

func TestRangeFloat64(t *testing.T) {
	if v := testKey("5.5").RangeFloat64(0, 1.0, 10.0); v != 5.5 {
		t.Errorf("got %f", v)
	}
	if v := testKey("20.0").RangeFloat64(0, 1.0, 10.0); v != 0 {
		t.Errorf("got %f", v)
	}
}

func TestRangeTime(t *testing.T) {
	min := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	max := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	def := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	k := testKey("2026-06-15T00:00:00Z")
	if v := k.RangeTime(def, min, max); v.Month() != 6 {
		t.Errorf("got %v", v)
	}
	k = testKey("2099-01-01T00:00:00Z")
	if v := k.RangeTime(def, min, max); !v.Equal(def) {
		t.Errorf("got %v, want default", v)
	}
}

func TestRangeTimeFormat(t *testing.T) {
	min := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	max := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	def := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	k := testKey("15/06/2026")
	if v := k.RangeTimeFormat("02/01/2006", def, min, max); v.Month() != 6 {
		t.Errorf("got %v", v)
	}
	k = testKey("15/06/2099")
	if v := k.RangeTimeFormat("02/01/2006", def, min, max); !v.Equal(def) {
		t.Errorf("got %v", v)
	}
}

// ---------------------------------------------------------------------------
// Lenient slice methods
// ---------------------------------------------------------------------------

func TestInts(t *testing.T) {
	v := testKey("1,2,3").Ints(",")
	if len(v) != 3 || v[0] != 1 || v[2] != 3 {
		t.Errorf("got %v", v)
	}
	// Skips invalid
	v = testKey("1,bad,3").Ints(",")
	if len(v) != 2 || v[0] != 1 || v[1] != 3 {
		t.Errorf("got %v", v)
	}
}

func TestInt64s(t *testing.T) {
	v := testKey("10,20").Int64s(",")
	if len(v) != 2 || v[0] != 10 || v[1] != 20 {
		t.Errorf("got %v", v)
	}
}

func TestUints(t *testing.T) {
	v := testKey("1,2,3").Uints(",")
	if len(v) != 3 {
		t.Errorf("got %v", v)
	}
	// Negative values skipped
	v = testKey("1,-2,3").Uints(",")
	if len(v) != 2 {
		t.Errorf("expected 2, got %v", v)
	}
}

func TestUint64s(t *testing.T) {
	v := testKey("100,200").Uint64s(",")
	if len(v) != 2 || v[0] != 100 {
		t.Errorf("got %v", v)
	}
}

func TestFloat64s(t *testing.T) {
	v := testKey("1.1,2.2,bad,3.3").Float64s(",")
	if len(v) != 3 || v[1] != 2.2 {
		t.Errorf("got %v", v)
	}
}

func TestBools(t *testing.T) {
	v := testKey("true,false,yes,bad,off").Bools(",")
	if len(v) != 4 {
		t.Errorf("expected 4, got %v", v)
	}
	if v[0] != true || v[1] != false || v[2] != true || v[3] != false {
		t.Errorf("got %v", v)
	}
}

func TestTimes(t *testing.T) {
	v := testKey("2026-01-01T00:00:00Z,bad,2026-06-01T00:00:00Z").Times(",")
	if len(v) != 2 {
		t.Errorf("expected 2, got %v", v)
	}
}

func TestTimesFormat(t *testing.T) {
	v := testKey("01/01/2026,bad,06/01/2026").TimesFormat("02/01/2006", ",")
	if len(v) != 2 {
		t.Errorf("expected 2, got %v", v)
	}
}

func TestEmptySlice(t *testing.T) {
	v := testKey("").Ints(",")
	if len(v) != 0 {
		t.Errorf("expected empty, got %v", v)
	}
}

// ---------------------------------------------------------------------------
// Strict slice methods
// ---------------------------------------------------------------------------

func TestStrictInts(t *testing.T) {
	v, err := testKey("1,2,3").StrictInts(",")
	if err != nil || len(v) != 3 {
		t.Errorf("got %v, %v", v, err)
	}
	_, err = testKey("1,bad,3").StrictInts(",")
	if err == nil {
		t.Error("expected error")
	}
}

func TestStrictInt64s(t *testing.T) {
	v, err := testKey("10,20").StrictInt64s(",")
	if err != nil || len(v) != 2 {
		t.Errorf("got %v, %v", v, err)
	}
	_, err = testKey("10,x").StrictInt64s(",")
	if err == nil {
		t.Error("expected error")
	}
}

func TestStrictUints(t *testing.T) {
	v, err := testKey("1,2").StrictUints(",")
	if err != nil || len(v) != 2 {
		t.Errorf("got %v, %v", v, err)
	}
	_, err = testKey("1,-2").StrictUints(",")
	if err == nil {
		t.Error("expected error for negative uint")
	}
}

func TestStrictUint64s(t *testing.T) {
	_, err := testKey("1,bad").StrictUint64s(",")
	if err == nil {
		t.Error("expected error")
	}
}

func TestStrictFloat64s(t *testing.T) {
	v, err := testKey("1.1,2.2").StrictFloat64s(",")
	if err != nil || len(v) != 2 {
		t.Errorf("got %v, %v", v, err)
	}
	_, err = testKey("1.1,bad").StrictFloat64s(",")
	if err == nil {
		t.Error("expected error")
	}
}

func TestStrictBools(t *testing.T) {
	v, err := testKey("true,false,yes").StrictBools(",")
	if err != nil || len(v) != 3 {
		t.Errorf("got %v, %v", v, err)
	}
	_, err = testKey("true,maybe").StrictBools(",")
	if err == nil {
		t.Error("expected error")
	}
}

func TestStrictTimes(t *testing.T) {
	v, err := testKey("2026-01-01T00:00:00Z,2026-06-01T00:00:00Z").StrictTimes(",")
	if err != nil || len(v) != 2 {
		t.Errorf("got %v, %v", v, err)
	}
	_, err = testKey("2026-01-01T00:00:00Z,bad").StrictTimes(",")
	if err == nil {
		t.Error("expected error")
	}
}

func TestStrictTimesFormat(t *testing.T) {
	v, err := testKey("01/01/2026,06/01/2026").StrictTimesFormat("02/01/2006", ",")
	if err != nil || len(v) != 2 {
		t.Errorf("got %v, %v", v, err)
	}
	_, err = testKey("01/01/2026,bad").StrictTimesFormat("02/01/2006", ",")
	if err == nil {
		t.Error("expected error")
	}
}

// ---------------------------------------------------------------------------
// Valid slice methods (aliases)
// ---------------------------------------------------------------------------

func TestValidAliases(t *testing.T) {
	k := testKey("1,bad,3")
	if len(k.ValidInts(",")) != 2 {
		t.Error("ValidInts")
	}
	if len(k.ValidInt64s(",")) != 2 {
		t.Error("ValidInt64s")
	}
	if len(k.ValidUints(",")) != 2 {
		t.Error("ValidUints")
	}
	if len(k.ValidUint64s(",")) != 2 {
		t.Error("ValidUint64s")
	}

	k2 := testKey("1.1,bad,3.3")
	if len(k2.ValidFloat64s(",")) != 2 {
		t.Error("ValidFloat64s")
	}

	k3 := testKey("true,bad,false")
	if len(k3.ValidBools(",")) != 2 {
		t.Error("ValidBools")
	}

	k4 := testKey("2026-01-01T00:00:00Z,bad")
	if len(k4.ValidTimes(",")) != 1 {
		t.Error("ValidTimes")
	}

	k5 := testKey("01/01/2026,bad")
	if len(k5.ValidTimesFormat("02/01/2006", ",")) != 1 {
		t.Error("ValidTimesFormat")
	}
}

// ---------------------------------------------------------------------------
// Hex / octal parsing via base 0
// ---------------------------------------------------------------------------

func TestHexOctalParsing(t *testing.T) {
	v, err := testKey("0xff").Int()
	if err != nil || v != 255 {
		t.Errorf("hex: got %d, %v", v, err)
	}
	v, err = testKey("0o77").Int()
	if err != nil || v != 63 {
		t.Errorf("octal: got %d, %v", v, err)
	}
}

// ---------------------------------------------------------------------------
// Whitespace in comma-separated values
// ---------------------------------------------------------------------------

func TestSpaceTrimming(t *testing.T) {
	v := testKey("  1 , 2 , 3 ").Ints(",")
	if len(v) != 3 || v[0] != 1 || v[1] != 2 || v[2] != 3 {
		t.Errorf("got %v", v)
	}
}
