package stanza

import (
	"strings"
	"sync"
)

// Key represents an INI key-value pair.
type Key struct {
	s               *Section
	name            string
	value           string
	Comment         string
	isBooleanKey    bool
	isAutoIncrement bool

	shadows  []*Key
	nestedValues []string

	mu sync.RWMutex
}

func newKey(s *Section, name, value string) *Key {
	return &Key{
		s:     s,
		name:  name,
		value: value,
	}
}

func newBooleanKey(s *Section, name string) *Key {
	return &Key{
		s:            s,
		name:         name,
		isBooleanKey: true,
	}
}

// Name returns the key name.
func (k *Key) Name() string {
	return k.name
}

// Value returns the key value.
func (k *Key) Value() string {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.value
}

// String returns the key value (same as Value).
func (k *Key) String() string {
	return k.Value()
}

// SetValue sets the key value.
func (k *Key) SetValue(v string) {
	if k.s != nil && k.s.f != nil && k.s.f.BlockMode {
		k.mu.Lock()
		defer k.mu.Unlock()
	}
	k.value = v
}

// ValueWithShadows returns all values including shadows.
func (k *Key) ValueWithShadows() []string {
	if k.s != nil && k.s.f != nil && k.s.f.BlockMode {
		k.mu.RLock()
		defer k.mu.RUnlock()
	}
	vals := make([]string, 1+len(k.shadows))
	vals[0] = k.value
	for i, shadow := range k.shadows {
		vals[i+1] = shadow.value
	}
	return vals
}

// NestedValues returns any nested (indented) values for this key.
func (k *Key) NestedValues() []string {
	if k.s != nil && k.s.f != nil && k.s.f.BlockMode {
		k.mu.RLock()
		defer k.mu.RUnlock()
	}
	return k.nestedValues
}

// AddShadow adds a shadow value to this key.
func (k *Key) AddShadow(val string) error {
	if k.s == nil || k.s.f == nil {
		return nil
	}
	if k.s.f.BlockMode {
		k.mu.Lock()
		defer k.mu.Unlock()
	}
	shadow := newKey(k.s, k.name, val)
	k.shadows = append(k.shadows, shadow)
	return nil
}

// AddNestedValue adds a nested value.
func (k *Key) AddNestedValue(val string) error {
	if k.s != nil && k.s.f != nil && k.s.f.BlockMode {
		k.mu.Lock()
		defer k.mu.Unlock()
	}
	k.nestedValues = append(k.nestedValues, val)
	return nil
}

// Validate passes the value through fn and returns the result.
func (k *Key) Validate(fn func(string) string) string {
	return fn(k.Value())
}

// Strings splits the value by the given delimiter and trims whitespace.
func (k *Key) Strings(delim string) []string {
	val := k.Value()
	if val == "" {
		return []string{}
	}
	parts := strings.Split(val, delim)
	result := make([]string, len(parts))
	for i, p := range parts {
		result[i] = strings.TrimSpace(p)
	}
	return result
}

// StringsWithShadows returns all values (including shadows) split by delim.
func (k *Key) StringsWithShadows(delim string) []string {
	vals := k.ValueWithShadows()
	var result []string
	for _, v := range vals {
		if v == "" {
			continue
		}
		for _, p := range strings.Split(v, delim) {
			result = append(result, strings.TrimSpace(p))
		}
	}
	return result
}

// IsBooleanKey reports whether this key has no value (e.g., "flag" with no "=").
func (k *Key) IsBooleanKey() bool {
	return k.isBooleanKey
}
