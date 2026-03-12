package stanza

import (
	"errors"
	"fmt"
	"sync"
)

// Section represents an INI section containing ordered key-value pairs.
type Section struct {
	f    *File
	name string

	Comment string

	// For raw (unparseable) sections.
	isRaw bool
	body  string

	keys     []*Key
	keysByName map[string]*Key

	mu sync.RWMutex
}

func newSection(f *File, name string) *Section {
	return &Section{
		f:          f,
		name:       name,
		keysByName: make(map[string]*Key),
	}
}

// Name returns the section name.
func (s *Section) Name() string {
	return s.name
}

// Body returns the raw body for unparseable sections.
func (s *Section) Body() string {
	return s.body
}

// SetBody sets the raw body for unparseable sections.
func (s *Section) SetBody(body string) {
	s.body = body
}

// NewKey creates a new key in this section.
// If a key with the same name already exists:
//   - if AllowShadows is set, it adds a shadow value
//   - otherwise it overwrites the existing value
func (s *Section) NewKey(name, val string) (*Key, error) {
	if name == "" {
		return nil, errors.New("empty key name")
	}
	if s.f != nil && s.f.BlockMode {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	lookupName := s.keyLookupName(name)
	if existing, ok := s.keysByName[lookupName]; ok {
		if s.f != nil && s.f.options.AllowShadows {
			shadow := newKey(s, name, val)
			existing.shadows = append(existing.shadows, shadow)
			return existing, nil
		}
		existing.value = val
		return existing, nil
	}
	k := newKey(s, name, val)
	s.keys = append(s.keys, k)
	s.keysByName[lookupName] = k
	return k, nil
}

// NewBooleanKey creates a key with no value.
func (s *Section) NewBooleanKey(name string) (*Key, error) {
	if name == "" {
		return nil, errors.New("empty key name")
	}
	if s.f != nil && s.f.BlockMode {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	k := newBooleanKey(s, name)
	lookupName := s.keyLookupName(name)
	s.keys = append(s.keys, k)
	s.keysByName[lookupName] = k
	return k, nil
}

// GetKey returns the key with the given name, or an error if not found.
func (s *Section) GetKey(name string) (*Key, error) {
	if s.f != nil && s.f.BlockMode {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}
	lookupName := s.keyLookupName(name)
	k, ok := s.keysByName[lookupName]
	if !ok {
		return nil, fmt.Errorf("key %q not found in section %q", name, s.name)
	}
	return k, nil
}

// Key returns the key with the given name.
// If the key does not exist, a new empty key is returned (never nil).
func (s *Section) Key(name string) *Key {
	if s.f != nil && s.f.BlockMode {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}
	lookupName := s.keyLookupName(name)
	if k, ok := s.keysByName[lookupName]; ok {
		return k
	}
	k := newKey(s, name, "")
	return k
}

// Keys returns all keys in this section in order.
func (s *Section) Keys() []*Key {
	if s.f != nil && s.f.BlockMode {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}
	return s.keys
}

// KeyStrings returns all key names in order.
func (s *Section) KeyStrings() []string {
	if s.f != nil && s.f.BlockMode {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}
	names := make([]string, len(s.keys))
	for i, k := range s.keys {
		names[i] = k.name
	}
	return names
}

// KeysHash returns a map of key names to values.
func (s *Section) KeysHash() map[string]string {
	if s.f != nil && s.f.BlockMode {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}
	m := make(map[string]string, len(s.keys))
	for _, k := range s.keys {
		m[k.name] = k.value
	}
	return m
}

// HasKey reports whether the section has a key with the given name.
func (s *Section) HasKey(name string) bool {
	if s.f != nil && s.f.BlockMode {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}
	lookupName := s.keyLookupName(name)
	_, ok := s.keysByName[lookupName]
	return ok
}

// HasValue reports whether any key in the section has the given value.
func (s *Section) HasValue(value string) bool {
	if s.f != nil && s.f.BlockMode {
		s.mu.RLock()
		defer s.mu.RUnlock()
	}
	for _, k := range s.keys {
		if k.value == value {
			return true
		}
	}
	return false
}

// DeleteKey deletes the key with the given name.
func (s *Section) DeleteKey(name string) {
	if s.f != nil && s.f.BlockMode {
		s.mu.Lock()
		defer s.mu.Unlock()
	}
	lookupName := s.keyLookupName(name)
	delete(s.keysByName, lookupName)
	for i, k := range s.keys {
		if s.keyLookupName(k.name) == lookupName {
			s.keys = append(s.keys[:i], s.keys[i+1:]...)
			return
		}
	}
}

// ParentKeys returns keys from the parent section (child section inheritance).
func (s *Section) ParentKeys() []*Key {
	if s.f == nil {
		return nil
	}
	delim := s.f.options.ChildSectionDelimiter
	if delim == "" {
		delim = "."
	}
	idx := lastIndex(s.name, delim)
	if idx < 0 {
		return nil
	}
	parentName := s.name[:idx]
	parent, err := s.f.GetSection(parentName)
	if err != nil {
		return nil
	}
	return parent.Keys()
}

// ChildSections returns subsections of this section.
func (s *Section) ChildSections() []*Section {
	if s.f == nil {
		return nil
	}
	delim := s.f.options.ChildSectionDelimiter
	if delim == "" {
		delim = "."
	}
	prefix := s.name + delim
	var children []*Section
	for _, sec := range s.f.sections {
		if len(sec.name) > len(prefix) && sec.name[:len(prefix)] == prefix {
			children = append(children, sec)
		}
	}
	return children
}

// keyLookupName returns the name used for map lookup, applying case folding
// if configured.
func (s *Section) keyLookupName(name string) string {
	if s.f == nil {
		return name
	}
	if s.f.options.Insensitive || s.f.options.InsensitiveKeys {
		return toLower(name)
	}
	return name
}

func lastIndex(s, sep string) int {
	n := len(sep)
	for i := len(s) - n; i >= 0; i-- {
		if s[i:i+n] == sep {
			return i
		}
	}
	return -1
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
