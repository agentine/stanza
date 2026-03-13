package stanza

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

// File represents an INI file with ordered sections.
type File struct {
	// BlockMode enables thread-safe access via read-write mutexes.
	// Defaults to true.
	BlockMode bool

	// NameMapper maps struct field names to INI key names.
	NameMapper NameMapper

	// ValueMapper transforms key values during loading.
	ValueMapper ValueMapper

	options LoadOptions

	sections     []*Section
	sectionsByName map[string]*Section
	// For AllowNonUniqueSections: name → []*Section
	sectionList map[string][]*Section

	sources []dataSource

	mu sync.RWMutex
}

// dataSource tracks where the file was loaded from for Reload().
type dataSource struct {
	kind string // "file", "bytes", "reader"
	path string // only for "file"
	data []byte // only for "bytes"
}

func newFile(opts LoadOptions) *File {
	f := &File{
		BlockMode:      true,
		options:        opts,
		sectionsByName: make(map[string]*Section),
		sectionList:    make(map[string][]*Section),
	}
	// Always create a default section.
	sec := newSection(f, DefaultSection)
	f.sections = append(f.sections, sec)
	f.sectionsByName[DefaultSection] = sec
	f.sectionList[DefaultSection] = []*Section{sec}
	return f
}

// Empty creates an empty File with the given options.
func Empty(opts ...LoadOptions) *File {
	var o LoadOptions
	if len(opts) > 0 {
		o = opts[0]
	}
	return newFile(o)
}

// ---------------------------------------------------------------------------
// Section management
// ---------------------------------------------------------------------------

// NewSection creates a new section with the given name.
func (f *File) NewSection(name string) (*Section, error) {
	if f.BlockMode {
		f.mu.Lock()
		defer f.mu.Unlock()
	}
	lookupName := f.sectionLookupName(name)
	if _, ok := f.sectionsByName[lookupName]; ok {
		if !f.options.AllowNonUniqueSections {
			return f.sectionsByName[lookupName], nil
		}
	}
	sec := newSection(f, name)
	f.sections = append(f.sections, sec)
	if _, ok := f.sectionsByName[lookupName]; !ok {
		f.sectionsByName[lookupName] = sec
	}
	f.sectionList[lookupName] = append(f.sectionList[lookupName], sec)
	return sec, nil
}

// NewSections creates multiple sections by name.
func (f *File) NewSections(names ...string) error {
	for _, name := range names {
		if _, err := f.NewSection(name); err != nil {
			return err
		}
	}
	return nil
}

// NewRawSection creates a section with a raw (unparsed) body.
func (f *File) NewRawSection(name, body string) (*Section, error) {
	sec, err := f.NewSection(name)
	if err != nil {
		return nil, err
	}
	sec.isRaw = true
	sec.body = body
	return sec, nil
}

// GetSection returns the section with the given name, or an error.
func (f *File) GetSection(name string) (*Section, error) {
	if f.BlockMode {
		f.mu.RLock()
		defer f.mu.RUnlock()
	}
	lookupName := f.sectionLookupName(name)
	sec, ok := f.sectionsByName[lookupName]
	if !ok {
		return nil, fmt.Errorf("section %q not found", name)
	}
	return sec, nil
}

// Section returns the section with the given name.
// If the section does not exist, a new empty section is created.
func (f *File) Section(name string) *Section {
	if f.BlockMode {
		f.mu.RLock()
		defer f.mu.RUnlock()
	}
	lookupName := f.sectionLookupName(name)
	if sec, ok := f.sectionsByName[lookupName]; ok {
		return sec
	}
	// Create a new section without the lock (avoid deadlock with NewSection).
	if f.BlockMode {
		f.mu.RUnlock()
		sec, _ := f.NewSection(name)
		f.mu.RLock()
		return sec
	}
	sec := newSection(f, name)
	f.sections = append(f.sections, sec)
	f.sectionsByName[lookupName] = sec
	f.sectionList[lookupName] = append(f.sectionList[lookupName], sec)
	return sec
}

// SectionWithIndex returns the i-th section with the given name
// (for AllowNonUniqueSections).
func (f *File) SectionWithIndex(name string, index int) *Section {
	if f.BlockMode {
		f.mu.RLock()
		defer f.mu.RUnlock()
	}
	lookupName := f.sectionLookupName(name)
	list := f.sectionList[lookupName]
	if index < 0 || index >= len(list) {
		return newSection(f, name)
	}
	return list[index]
}

// Sections returns all sections in order.
func (f *File) Sections() []*Section {
	if f.BlockMode {
		f.mu.RLock()
		defer f.mu.RUnlock()
	}
	return f.sections
}

// SectionsByName returns all sections with the given name.
func (f *File) SectionsByName(name string) ([]*Section, error) {
	if f.BlockMode {
		f.mu.RLock()
		defer f.mu.RUnlock()
	}
	lookupName := f.sectionLookupName(name)
	list, ok := f.sectionList[lookupName]
	if !ok || len(list) == 0 {
		return nil, fmt.Errorf("section %q not found", name)
	}
	return list, nil
}

// SectionStrings returns the names of all sections in order.
func (f *File) SectionStrings() []string {
	if f.BlockMode {
		f.mu.RLock()
		defer f.mu.RUnlock()
	}
	seen := make(map[string]bool)
	var names []string
	for _, sec := range f.sections {
		if !seen[sec.name] {
			names = append(names, sec.name)
			seen[sec.name] = true
		}
	}
	return names
}

// HasSection reports whether the file has a section with the given name.
func (f *File) HasSection(name string) bool {
	if f.BlockMode {
		f.mu.RLock()
		defer f.mu.RUnlock()
	}
	lookupName := f.sectionLookupName(name)
	_, ok := f.sectionsByName[lookupName]
	return ok
}

// DeleteSection deletes the first section with the given name.
// When AllowNonUniqueSections is set, only the first section is
// removed and remaining duplicates are preserved.
func (f *File) DeleteSection(name string) {
	if f.BlockMode {
		f.mu.Lock()
		defer f.mu.Unlock()
	}
	lookupName := f.sectionLookupName(name)
	list := f.sectionList[lookupName]
	if len(list) <= 1 {
		// Last (or only) section with this name — remove from maps entirely.
		delete(f.sectionsByName, lookupName)
		delete(f.sectionList, lookupName)
	} else {
		// Remove the first entry from the list; update sectionsByName to
		// point to the new first section.
		f.sectionList[lookupName] = list[1:]
		f.sectionsByName[lookupName] = list[1]
	}
	for i, sec := range f.sections {
		if f.sectionLookupName(sec.name) == lookupName {
			f.sections = append(f.sections[:i], f.sections[i+1:]...)
			return
		}
	}
}

// DeleteSectionWithIndex deletes the i-th section with the given name.
func (f *File) DeleteSectionWithIndex(name string, index int) error {
	if f.BlockMode {
		f.mu.Lock()
		defer f.mu.Unlock()
	}
	lookupName := f.sectionLookupName(name)
	list := f.sectionList[lookupName]
	if index < 0 || index >= len(list) {
		return fmt.Errorf("section %q index %d out of range", name, index)
	}
	target := list[index]
	f.sectionList[lookupName] = append(list[:index], list[index+1:]...)
	if len(f.sectionList[lookupName]) == 0 {
		delete(f.sectionsByName, lookupName)
		delete(f.sectionList, lookupName)
	} else if index == 0 {
		f.sectionsByName[lookupName] = f.sectionList[lookupName][0]
	}
	for i, sec := range f.sections {
		if sec == target {
			f.sections = append(f.sections[:i], f.sections[i+1:]...)
			break
		}
	}
	return nil
}

// ChildSections returns all child sections of the named section.
func (f *File) ChildSections(name string) []*Section {
	sec, err := f.GetSection(name)
	if err != nil {
		return nil
	}
	return sec.ChildSections()
}

// ---------------------------------------------------------------------------
// I/O
// ---------------------------------------------------------------------------

// Append loads additional sources and merges them into this file.
func (f *File) Append(source interface{}, others ...interface{}) error {
	srcs := append([]interface{}{source}, others...)
	for _, src := range srcs {
		if err := f.loadSource(src); err != nil {
			return err
		}
	}
	return nil
}

// Reload re-reads all original sources.
func (f *File) Reload() error {
	// Reset sections, keep only default.
	f.sections = nil
	f.sectionsByName = make(map[string]*Section)
	f.sectionList = make(map[string][]*Section)
	sec := newSection(f, DefaultSection)
	f.sections = append(f.sections, sec)
	f.sectionsByName[DefaultSection] = sec
	f.sectionList[DefaultSection] = []*Section{sec}

	for _, ds := range f.sources {
		switch ds.kind {
		case "file":
			if err := f.loadSource(ds.path); err != nil {
				return err
			}
		case "bytes":
			if err := f.loadSource(ds.data); err != nil {
				return err
			}
		}
	}
	return nil
}

// WriteTo writes the INI representation to w.
func (f *File) WriteTo(w io.Writer) (int64, error) {
	return f.WriteToIndent(w, "")
}

// WriteToIndent writes the INI representation with indentation.
func (f *File) WriteToIndent(w io.Writer, indent string) (int64, error) {
	if f.BlockMode {
		f.mu.RLock()
		defer f.mu.RUnlock()
	}

	var buf bytes.Buffer
	delim := f.options.KeyValueDelimiterOnWrite
	if delim == "" {
		delim = "="
	}
	if PrettyEqual {
		delim = " " + delim + " "
	}

	for i, sec := range f.sections {
		if sec.name == DefaultSection {
			if !DefaultHeader && len(sec.keys) == 0 {
				continue
			}
			if DefaultHeader {
				if i > 0 && PrettySection {
					buf.WriteString(LineBreak)
				}
				if sec.Comment != "" {
					buf.WriteString(sec.Comment)
					buf.WriteString(LineBreak)
				}
				buf.WriteString("[" + sec.name + "]")
				buf.WriteString(LineBreak)
			}
		} else {
			if i > 0 && PrettySection {
				buf.WriteString(LineBreak)
			}
			if sec.Comment != "" {
				buf.WriteString(sec.Comment)
				buf.WriteString(LineBreak)
			}
			buf.WriteString("[" + sec.name + "]")
			buf.WriteString(LineBreak)
		}

		if sec.isRaw {
			buf.WriteString(sec.body)
			buf.WriteString(LineBreak)
			continue
		}

		for _, k := range sec.keys {
			if k.Comment != "" {
				buf.WriteString(k.Comment)
				buf.WriteString(LineBreak)
			}
			if k.isBooleanKey {
				buf.WriteString(indent + DefaultFormatLeft + k.name + DefaultFormatRight)
			} else {
				buf.WriteString(indent + DefaultFormatLeft + k.name + DefaultFormatRight + delim + k.value)
			}
			buf.WriteString(LineBreak)
		}

		if PrettyFormat && i < len(f.sections)-1 {
			// blank line handled by PrettySection above
		}
	}

	n, err := w.Write(buf.Bytes())
	return int64(n), err
}

// SaveTo writes the INI representation to a file.
func (f *File) SaveTo(filename string) error {
	return f.SaveToIndent(filename, "")
}

// SaveToIndent writes the INI representation to a file with indentation.
func (f *File) SaveToIndent(filename, indent string) error {
	var buf bytes.Buffer
	if _, err := f.WriteToIndent(&buf, indent); err != nil {
		return err
	}
	return os.WriteFile(filename, buf.Bytes(), 0o644)
}

// loadSource loads a single source into the file.
func (f *File) loadSource(source interface{}) error {
	var data []byte
	var err error
	switch v := source.(type) {
	case string:
		data, err = os.ReadFile(v)
		if err != nil {
			if f.options.Loose {
				return nil
			}
			return err
		}
		f.sources = append(f.sources, dataSource{kind: "file", path: v})
	case []byte:
		data = v
		f.sources = append(f.sources, dataSource{kind: "bytes", data: v})
	case io.ReadCloser:
		defer v.Close()
		data, err = io.ReadAll(v)
		if err != nil {
			return err
		}
		f.sources = append(f.sources, dataSource{kind: "bytes", data: data})
	case io.Reader:
		data, err = io.ReadAll(v)
		if err != nil {
			return err
		}
		f.sources = append(f.sources, dataSource{kind: "bytes", data: data})
	default:
		return errors.New("unsupported source type")
	}

	if f.ValueMapper != nil {
		// Parse with value mapper applied later in parser.
	}

	return parseINI(f, data)
}

func (f *File) sectionLookupName(name string) string {
	if f.options.Insensitive || f.options.InsensitiveSections {
		return toLower(name)
	}
	return name
}
