package stanza

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Basic type conversions (returns error)
// ---------------------------------------------------------------------------

// Bool returns the boolean value of the key.
func (k *Key) Bool() (bool, error) {
	return parseBool(k.Value())
}

// Int returns the int value of the key.
func (k *Key) Int() (int, error) {
	v, err := strconv.ParseInt(k.Value(), 0, 64)
	return int(v), err
}

// Int64 returns the int64 value of the key.
func (k *Key) Int64() (int64, error) {
	return strconv.ParseInt(k.Value(), 0, 64)
}

// Uint returns the uint value of the key.
func (k *Key) Uint() (uint, error) {
	v, err := strconv.ParseUint(k.Value(), 0, 64)
	return uint(v), err
}

// Uint64 returns the uint64 value of the key.
func (k *Key) Uint64() (uint64, error) {
	return strconv.ParseUint(k.Value(), 0, 64)
}

// Float64 returns the float64 value of the key.
func (k *Key) Float64() (float64, error) {
	return strconv.ParseFloat(k.Value(), 64)
}

// Duration returns the time.Duration value of the key.
func (k *Key) Duration() (time.Duration, error) {
	return time.ParseDuration(k.Value())
}

// Time returns the time.Time value of the key, parsed as RFC3339.
func (k *Key) Time() (time.Time, error) {
	return time.Parse(time.RFC3339, k.Value())
}

// TimeFormat returns the time.Time value of the key, parsed with the given format.
func (k *Key) TimeFormat(format string) (time.Time, error) {
	return time.Parse(format, k.Value())
}

// ---------------------------------------------------------------------------
// Must methods (default on error)
// ---------------------------------------------------------------------------

// MustString returns the value, or defaultVal if empty.
func (k *Key) MustString(defaultVal string) string {
	v := k.Value()
	if v == "" {
		return defaultVal
	}
	return v
}

// MustBool returns the bool value, or defaultVal on error.
func (k *Key) MustBool(defaultVal ...bool) bool {
	v, err := k.Bool()
	if err != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return false
	}
	return v
}

// MustInt returns the int value, or defaultVal on error.
func (k *Key) MustInt(defaultVal ...int) int {
	v, err := k.Int()
	if err != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}
	return v
}

// MustInt64 returns the int64 value, or defaultVal on error.
func (k *Key) MustInt64(defaultVal ...int64) int64 {
	v, err := k.Int64()
	if err != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}
	return v
}

// MustUint returns the uint value, or defaultVal on error.
func (k *Key) MustUint(defaultVal ...uint) uint {
	v, err := k.Uint()
	if err != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}
	return v
}

// MustUint64 returns the uint64 value, or defaultVal on error.
func (k *Key) MustUint64(defaultVal ...uint64) uint64 {
	v, err := k.Uint64()
	if err != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}
	return v
}

// MustFloat64 returns the float64 value, or defaultVal on error.
func (k *Key) MustFloat64(defaultVal ...float64) float64 {
	v, err := k.Float64()
	if err != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}
	return v
}

// MustDuration returns the Duration value, or defaultVal on error.
func (k *Key) MustDuration(defaultVal ...time.Duration) time.Duration {
	v, err := k.Duration()
	if err != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}
	return v
}

// MustTime returns the Time value (RFC3339), or defaultVal on error.
func (k *Key) MustTime(defaultVal ...time.Time) time.Time {
	v, err := k.Time()
	if err != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return time.Time{}
	}
	return v
}

// MustTimeFormat returns the Time value, or defaultVal on error.
func (k *Key) MustTimeFormat(format string, defaultVal ...time.Time) time.Time {
	v, err := k.TimeFormat(format)
	if err != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return time.Time{}
	}
	return v
}

// ---------------------------------------------------------------------------
// In methods (validate against candidates)
// ---------------------------------------------------------------------------

// In returns the value if it's in candidates, otherwise defaultVal.
func (k *Key) In(defaultVal string, candidates []string) string {
	v := k.Value()
	for _, c := range candidates {
		if v == c {
			return v
		}
	}
	return defaultVal
}

// InInt returns the int value if it's in candidates, otherwise defaultVal.
func (k *Key) InInt(defaultVal int, candidates []int) int {
	v := k.MustInt()
	for _, c := range candidates {
		if v == c {
			return v
		}
	}
	return defaultVal
}

// InInt64 returns the int64 value if it's in candidates, otherwise defaultVal.
func (k *Key) InInt64(defaultVal int64, candidates []int64) int64 {
	v := k.MustInt64()
	for _, c := range candidates {
		if v == c {
			return v
		}
	}
	return defaultVal
}

// InUint returns the uint value if it's in candidates, otherwise defaultVal.
func (k *Key) InUint(defaultVal uint, candidates []uint) uint {
	v := k.MustUint()
	for _, c := range candidates {
		if v == c {
			return v
		}
	}
	return defaultVal
}

// InUint64 returns the uint64 value if it's in candidates, otherwise defaultVal.
func (k *Key) InUint64(defaultVal uint64, candidates []uint64) uint64 {
	v := k.MustUint64()
	for _, c := range candidates {
		if v == c {
			return v
		}
	}
	return defaultVal
}

// InFloat64 returns the float64 value if it's in candidates, otherwise defaultVal.
func (k *Key) InFloat64(defaultVal float64, candidates []float64) float64 {
	v := k.MustFloat64()
	for _, c := range candidates {
		if v == c {
			return v
		}
	}
	return defaultVal
}

// InTime returns the Time value if it's in candidates, otherwise defaultVal.
func (k *Key) InTime(defaultVal time.Time, candidates []time.Time) time.Time {
	v := k.MustTime()
	for _, c := range candidates {
		if v.Equal(c) {
			return v
		}
	}
	return defaultVal
}

// InTimeFormat returns the Time value if it's in candidates, otherwise defaultVal.
func (k *Key) InTimeFormat(format string, defaultVal time.Time, candidates []time.Time) time.Time {
	v := k.MustTimeFormat(format)
	for _, c := range candidates {
		if v.Equal(c) {
			return v
		}
	}
	return defaultVal
}

// ---------------------------------------------------------------------------
// Range methods (validate within bounds)
// ---------------------------------------------------------------------------

// RangeInt returns the int value clamped to [min, max], or defaultVal on error.
func (k *Key) RangeInt(defaultVal, min, max int) int {
	v := k.MustInt(defaultVal)
	if v < min || v > max {
		return defaultVal
	}
	return v
}

// RangeInt64 returns the int64 value clamped to [min, max].
func (k *Key) RangeInt64(defaultVal, min, max int64) int64 {
	v := k.MustInt64(defaultVal)
	if v < min || v > max {
		return defaultVal
	}
	return v
}

// RangeFloat64 returns the float64 value clamped to [min, max].
func (k *Key) RangeFloat64(defaultVal, min, max float64) float64 {
	v := k.MustFloat64(defaultVal)
	if v < min || v > max {
		return defaultVal
	}
	return v
}

// RangeTime returns the Time value if within [min, max].
func (k *Key) RangeTime(defaultVal, min, max time.Time) time.Time {
	v := k.MustTime(defaultVal)
	if v.Before(min) || v.After(max) {
		return defaultVal
	}
	return v
}

// RangeTimeFormat returns the Time value if within [min, max].
func (k *Key) RangeTimeFormat(format string, defaultVal, min, max time.Time) time.Time {
	v := k.MustTimeFormat(format, defaultVal)
	if v.Before(min) || v.After(max) {
		return defaultVal
	}
	return v
}

// ---------------------------------------------------------------------------
// Slice methods (lenient — skip invalid)
// ---------------------------------------------------------------------------

// Ints splits by delim and parses as int, skipping errors.
func (k *Key) Ints(delim string) []int {
	parts := k.Strings(delim)
	var result []int
	for _, p := range parts {
		v, err := strconv.Atoi(p)
		if err == nil {
			result = append(result, v)
		}
	}
	return result
}

// Int64s splits by delim and parses as int64, skipping errors.
func (k *Key) Int64s(delim string) []int64 {
	parts := k.Strings(delim)
	var result []int64
	for _, p := range parts {
		v, err := strconv.ParseInt(p, 0, 64)
		if err == nil {
			result = append(result, v)
		}
	}
	return result
}

// Uints splits by delim and parses as uint, skipping errors.
func (k *Key) Uints(delim string) []uint {
	parts := k.Strings(delim)
	var result []uint
	for _, p := range parts {
		v, err := strconv.ParseUint(p, 0, 64)
		if err == nil {
			result = append(result, uint(v))
		}
	}
	return result
}

// Uint64s splits by delim and parses as uint64, skipping errors.
func (k *Key) Uint64s(delim string) []uint64 {
	parts := k.Strings(delim)
	var result []uint64
	for _, p := range parts {
		v, err := strconv.ParseUint(p, 0, 64)
		if err == nil {
			result = append(result, v)
		}
	}
	return result
}

// Float64s splits by delim and parses as float64, skipping errors.
func (k *Key) Float64s(delim string) []float64 {
	parts := k.Strings(delim)
	var result []float64
	for _, p := range parts {
		v, err := strconv.ParseFloat(p, 64)
		if err == nil {
			result = append(result, v)
		}
	}
	return result
}

// Bools splits by delim and parses as bool, skipping errors.
func (k *Key) Bools(delim string) []bool {
	parts := k.Strings(delim)
	var result []bool
	for _, p := range parts {
		v, err := parseBool(p)
		if err == nil {
			result = append(result, v)
		}
	}
	return result
}

// Times splits by delim and parses as time.Time (RFC3339), skipping errors.
func (k *Key) Times(delim string) []time.Time {
	return k.TimesFormat(time.RFC3339, delim)
}

// TimesFormat splits by delim and parses with format, skipping errors.
func (k *Key) TimesFormat(format, delim string) []time.Time {
	parts := k.Strings(delim)
	var result []time.Time
	for _, p := range parts {
		v, err := time.Parse(format, p)
		if err == nil {
			result = append(result, v)
		}
	}
	return result
}

// ---------------------------------------------------------------------------
// Strict slice methods (error on first invalid)
// ---------------------------------------------------------------------------

// StrictInts splits by delim and parses as int, returning error on any invalid.
func (k *Key) StrictInts(delim string) ([]int, error) {
	parts := k.Strings(delim)
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid int %q: %w", p, err)
		}
		result = append(result, v)
	}
	return result, nil
}

// StrictInt64s splits by delim and parses as int64.
func (k *Key) StrictInt64s(delim string) ([]int64, error) {
	parts := k.Strings(delim)
	result := make([]int64, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.ParseInt(p, 0, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid int64 %q: %w", p, err)
		}
		result = append(result, v)
	}
	return result, nil
}

// StrictUints splits by delim and parses as uint.
func (k *Key) StrictUints(delim string) ([]uint, error) {
	parts := k.Strings(delim)
	result := make([]uint, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.ParseUint(p, 0, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid uint %q: %w", p, err)
		}
		result = append(result, uint(v))
	}
	return result, nil
}

// StrictUint64s splits by delim and parses as uint64.
func (k *Key) StrictUint64s(delim string) ([]uint64, error) {
	parts := k.Strings(delim)
	result := make([]uint64, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.ParseUint(p, 0, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid uint64 %q: %w", p, err)
		}
		result = append(result, v)
	}
	return result, nil
}

// StrictFloat64s splits by delim and parses as float64.
func (k *Key) StrictFloat64s(delim string) ([]float64, error) {
	parts := k.Strings(delim)
	result := make([]float64, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float64 %q: %w", p, err)
		}
		result = append(result, v)
	}
	return result, nil
}

// StrictBools splits by delim and parses as bool.
func (k *Key) StrictBools(delim string) ([]bool, error) {
	parts := k.Strings(delim)
	result := make([]bool, 0, len(parts))
	for _, p := range parts {
		v, err := parseBool(p)
		if err != nil {
			return nil, fmt.Errorf("invalid bool %q: %w", p, err)
		}
		result = append(result, v)
	}
	return result, nil
}

// StrictTimes splits by delim and parses as time.Time (RFC3339).
func (k *Key) StrictTimes(delim string) ([]time.Time, error) {
	return k.StrictTimesFormat(time.RFC3339, delim)
}

// StrictTimesFormat splits by delim and parses with format.
func (k *Key) StrictTimesFormat(format, delim string) ([]time.Time, error) {
	parts := k.Strings(delim)
	result := make([]time.Time, 0, len(parts))
	for _, p := range parts {
		v, err := time.Parse(format, p)
		if err != nil {
			return nil, fmt.Errorf("invalid time %q: %w", p, err)
		}
		result = append(result, v)
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Valid slice methods (skip invalid entries)
// These are aliases for the lenient methods.
// ---------------------------------------------------------------------------

// ValidInts is like Ints — skips invalid entries.
func (k *Key) ValidInts(delim string) []int         { return k.Ints(delim) }
func (k *Key) ValidInt64s(delim string) []int64      { return k.Int64s(delim) }
func (k *Key) ValidUints(delim string) []uint        { return k.Uints(delim) }
func (k *Key) ValidUint64s(delim string) []uint64    { return k.Uint64s(delim) }
func (k *Key) ValidFloat64s(delim string) []float64  { return k.Float64s(delim) }
func (k *Key) ValidBools(delim string) []bool        { return k.Bools(delim) }
func (k *Key) ValidTimes(delim string) []time.Time   { return k.Times(delim) }
func (k *Key) ValidTimesFormat(format, delim string) []time.Time {
	return k.TimesFormat(format, delim)
}

// ---------------------------------------------------------------------------
// Bool parser (matching go-ini/ini)
// ---------------------------------------------------------------------------

func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "t", "true", "yes", "y", "on":
		return true, nil
	case "0", "f", "false", "no", "n", "off":
		return false, nil
	}
	return false, fmt.Errorf("cannot parse %q as bool", s)
}
