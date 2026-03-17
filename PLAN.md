# stanza — Drop-in Replacement for go-ini/ini

## Overview

**Target:** [go-ini/ini](https://github.com/go-ini/ini) (gopkg.in/ini.v1) — the most popular Go INI file parser
**Module:** `github.com/agentine/stanza`
**License:** MIT
**Go:** 1.25+
**Dependencies:** Zero

## Why Replace go-ini/ini

- 3,537 stars, 5,710 importers on pkg.go.dev
- Single maintainer (Joe Chen) with irregular engagement
- 3.5-year gap between v1.67.0 (Aug 2022) and v1.67.1 (Jan 2026)
- 70 open issues including broken comment/escape handling
- Known bugs: inline comment parsing errors, escaping mechanism broken, SaveTo indentation issues, MapTo uint8 failure, quotation mark auto-insertion
- API surface is large (~80 Key methods) but well-defined and testable

## Architecture

stanza is a single-package Go library with three layers:

```
Loading (Load/LoadSources/LooseLoad/InsensitiveLoad/ShadowLoad/Empty)
    ↓
Data Model (File → Section → Key)
    ↓
Serialization (WriteTo/SaveTo) & Struct Mapping (MapTo/ReflectFrom)
```

The parser reads INI text line-by-line, building a `File` containing ordered `Section`s, each containing ordered `Key`s. Keys support type conversion, validation, slicing, and struct mapping via reflection.

## Public API Surface (100% go-ini/ini compatible)

### Package-Level Loading Functions

```go
func Load(source interface{}, others ...interface{}) (*File, error)
func LoadSources(opts LoadOptions, source interface{}, others ...interface{}) (*File, error)
func LooseLoad(source interface{}, others ...interface{}) (*File, error)
func InsensitiveLoad(source interface{}, others ...interface{}) (*File, error)
func ShadowLoad(source interface{}, others ...interface{}) (*File, error)
func Empty(opts ...LoadOptions) *File
```

Source types: `string` (file path), `[]byte`, `io.Reader`, `io.ReadCloser`.

### Package-Level Struct Mapping

```go
func MapTo(v, source interface{}, others ...interface{}) error
func MapToWithMapper(v interface{}, mapper NameMapper, source interface{}, others ...interface{}) error
func StrictMapTo(v, source interface{}, others ...interface{}) error
func StrictMapToWithMapper(v interface{}, mapper NameMapper, source interface{}, others ...interface{}) error
func ReflectFrom(cfg *File, v interface{}) error
func ReflectFromWithMapper(cfg *File, v interface{}, mapper NameMapper) error
```

### Package-Level Variables & Constants

```go
const DefaultSection = "DEFAULT"

var LineBreak = "\n"         // "\r\n" on Windows
var DefaultHeader = false
var PrettySection = true
var PrettyFormat = true
var PrettyEqual = false
var DefaultFormatLeft = ""
var DefaultFormatRight = ""
```

### Function Types

```go
type NameMapper func(string) string
type ValueMapper func(string) string
type DebugFunc func(message string)

// Built-in NameMappers
var SnackCase NameMapper     // field_name → FIELD_NAME
var TitleUnderscore NameMapper // FieldName → Field_Name
```

### Error Types

```go
type ErrDelimiterNotFound struct { Line string }
type ErrEmptyKeyName struct { Line string }
func IsErrDelimiterNotFound(err error) bool
func IsErrEmptyKeyName(err error) bool
```

### LoadOptions (24 fields)

```go
type LoadOptions struct {
    Loose                      bool     // ignore missing files
    Insensitive                bool     // force all names lowercase
    InsensitiveSections        bool     // force section names lowercase
    InsensitiveKeys            bool     // force key names lowercase
    IgnoreContinuation         bool     // ignore backslash continuation
    IgnoreInlineComment        bool     // treat #/; in values as literal
    SkipUnrecognizableLines    bool     // skip malformed lines
    ShortCircuit               bool     // stop after first source
    AllowBooleanKeys           bool     // keys without values
    AllowShadows               bool     // duplicate key names
    AllowNestedValues          bool     // AWS-style indented values
    AllowPythonMultilineValues bool     // Python configparser multiline
    SpaceBeforeInlineComment   bool     // require space before #/;
    UnescapeValueDoubleQuotes  bool     // unescape \" in values
    UnescapeValueCommentSymbols bool    // unescape \# and \;
    UnparseableSections        []string // raw body sections
    KeyValueDelimiters         string   // default "=:"
    KeyValueDelimiterOnWrite   string   // default "="
    ChildSectionDelimiter      string   // default "."
    PreserveSurroundedQuote    bool     // keep surrounding quotes
    DebugFunc                  DebugFunc
    ReaderBufferSize           int
    AllowNonUniqueSections     bool     // multiple sections same name
    AllowDuplicateShadowValues bool
}
```

### File Type

```go
type File struct {
    BlockMode  bool
    NameMapper NameMapper
    ValueMapper ValueMapper
}

// Section management
func (f *File) NewSection(name string) (*Section, error)
func (f *File) NewSections(names ...string) error
func (f *File) NewRawSection(name, body string) (*Section, error)
func (f *File) GetSection(name string) (*Section, error)
func (f *File) Section(name string) *Section
func (f *File) SectionWithIndex(name string, index int) *Section
func (f *File) Sections() []*Section
func (f *File) SectionsByName(name string) ([]*Section, error)
func (f *File) SectionStrings() []string
func (f *File) HasSection(name string) bool
func (f *File) DeleteSection(name string)
func (f *File) DeleteSectionWithIndex(name string, index int) error
func (f *File) ChildSections(name string) []*Section

// I/O
func (f *File) Append(source interface{}, others ...interface{}) error
func (f *File) Reload() error
func (f *File) WriteTo(w io.Writer) (int64, error)
func (f *File) WriteToIndent(w io.Writer, indent string) (int64, error)
func (f *File) SaveTo(filename string) error
func (f *File) SaveToIndent(filename, indent string) error

// Struct mapping
func (f *File) MapTo(v interface{}) error
func (f *File) StrictMapTo(v interface{}) error
func (f *File) ReflectFrom(v interface{}) error
```

### Section Type

```go
type Section struct {
    Comment string
}

func (s *Section) Name() string
func (s *Section) Body() string
func (s *Section) SetBody(body string)
func (s *Section) NewKey(name, val string) (*Key, error)
func (s *Section) NewBooleanKey(name string) (*Key, error)
func (s *Section) GetKey(name string) (*Key, error)
func (s *Section) Key(name string) *Key
func (s *Section) Keys() []*Key
func (s *Section) KeyStrings() []string
func (s *Section) KeysHash() map[string]string
func (s *Section) ParentKeys() []*Key
func (s *Section) HasKey(name string) bool
func (s *Section) HasValue(value string) bool
func (s *Section) DeleteKey(name string)
func (s *Section) ChildSections() []*Section
func (s *Section) MapTo(v interface{}) error
func (s *Section) StrictMapTo(v interface{}) error
func (s *Section) ReflectFrom(v interface{}) error
```

### Key Type — Value Access & Type Conversion

```go
type Key struct {
    Comment string
}

// Basic access
func (k *Key) Name() string
func (k *Key) Value() string
func (k *Key) String() string
func (k *Key) SetValue(v string)
func (k *Key) ValueWithShadows() []string
func (k *Key) NestedValues() []string
func (k *Key) AddShadow(val string) error
func (k *Key) AddNestedValue(val string) error
func (k *Key) Validate(fn func(string) string) string

// Type conversion (returns error)
func (k *Key) Bool() (bool, error)
func (k *Key) Int() (int, error)
func (k *Key) Int64() (int64, error)
func (k *Key) Uint() (uint, error)
func (k *Key) Uint64() (uint64, error)
func (k *Key) Float64() (float64, error)
func (k *Key) Duration() (time.Duration, error)
func (k *Key) Time() (time.Time, error)
func (k *Key) TimeFormat(format string) (time.Time, error)

// Must methods (default on error)
func (k *Key) MustString(defaultVal string) string
func (k *Key) MustBool(defaultVal ...bool) bool
func (k *Key) MustInt(defaultVal ...int) int
func (k *Key) MustInt64(defaultVal ...int64) int64
func (k *Key) MustUint(defaultVal ...uint) uint
func (k *Key) MustUint64(defaultVal ...uint64) uint64
func (k *Key) MustFloat64(defaultVal ...float64) float64
func (k *Key) MustDuration(defaultVal ...time.Duration) time.Duration
func (k *Key) MustTime(defaultVal ...time.Time) time.Time
func (k *Key) MustTimeFormat(format string, defaultVal ...time.Time) time.Time

// In methods (validate against candidates)
func (k *Key) In(defaultVal string, candidates []string) string
func (k *Key) InInt(defaultVal int, candidates []int) int
func (k *Key) InInt64(defaultVal int64, candidates []int64) int64
func (k *Key) InUint(defaultVal uint, candidates []uint) uint
func (k *Key) InUint64(defaultVal uint64, candidates []uint64) uint64
func (k *Key) InFloat64(defaultVal float64, candidates []float64) float64
func (k *Key) InTime(defaultVal time.Time, candidates []time.Time) time.Time
func (k *Key) InTimeFormat(format string, defaultVal time.Time, candidates []time.Time) time.Time

// Range methods (validate within bounds)
func (k *Key) RangeInt(defaultVal, min, max int) int
func (k *Key) RangeInt64(defaultVal, min, max int64) int64
func (k *Key) RangeFloat64(defaultVal, min, max float64) float64
func (k *Key) RangeTime(defaultVal, min, max time.Time) time.Time
func (k *Key) RangeTimeFormat(format string, defaultVal, min, max time.Time) time.Time

// Slice methods (delimiter-separated, lenient)
func (k *Key) Strings(delim string) []string
func (k *Key) StringsWithShadows(delim string) []string
func (k *Key) Ints(delim string) []int
func (k *Key) Int64s(delim string) []int64
func (k *Key) Uints(delim string) []uint
func (k *Key) Uint64s(delim string) []uint64
func (k *Key) Float64s(delim string) []float64
func (k *Key) Bools(delim string) []bool
func (k *Key) Times(delim string) []time.Time
func (k *Key) TimesFormat(format, delim string) []time.Time

// Strict slice methods (error on first invalid)
func (k *Key) StrictInts(delim string) ([]int, error)
func (k *Key) StrictInt64s(delim string) ([]int64, error)
func (k *Key) StrictUints(delim string) ([]uint, error)
func (k *Key) StrictUint64s(delim string) ([]uint64, error)
func (k *Key) StrictFloat64s(delim string) ([]float64, error)
func (k *Key) StrictBools(delim string) ([]bool, error)
func (k *Key) StrictTimes(delim string) ([]time.Time, error)
func (k *Key) StrictTimesFormat(format, delim string) ([]time.Time, error)

// Valid slice methods (skip invalid entries)
func (k *Key) ValidInts(delim string) []int
func (k *Key) ValidInt64s(delim string) []int64
func (k *Key) ValidUints(delim string) []uint
func (k *Key) ValidUint64s(delim string) []uint64
func (k *Key) ValidFloat64s(delim string) []float64
func (k *Key) ValidBools(delim string) []bool
func (k *Key) ValidTimes(delim string) []time.Time
func (k *Key) ValidTimesFormat(format, delim string) []time.Time
```

### INI Format Features

**Comments:** `#` and `;` as line comments. Inline comments configurable via `SpaceBeforeInlineComment` and `IgnoreInlineComment`.

**Multi-line values:**
1. Backslash continuation: `value = long \` + next line
2. Python-style: indented continuation lines
3. Triple-quoted: `key = """multi\nline"""`

**Key-value delimiters:** `=` and `:` (configurable)

**Boolean values:** `1/t/T/true/TRUE/True/YES/yes/Yes/y/ON/on/On` and inverse

**Section inheritance:** `[parent.child]` with configurable delimiter

**Auto-increment keys:** Key name `-` becomes `#1`, `#2`, etc.

**BOM handling:** UTF-8, UTF-16 LE/BE BOM detection and stripping

### Struct Tag Format

```go
type Config struct {
    Name    string        `ini:"name"`
    Port    int           `ini:"port"`
    Debug   bool          `ini:"debug,omitempty"`
    Tags    []string      `ini:"tags" delim:","`
    Comment string        `comment:"Server name"`
    DB      DatabaseConfig `ini:"database"`  // becomes [database] section
}
```

Tag options: `omitempty`, `allowshadow`, `nonunique`, `extends`

## Bug Fixes Over go-ini/ini

1. **Inline comment parsing** — correctly handle `#` and `;` in values with all config combinations
2. **Escape handling** — proper `\#`, `\;`, `\"` escaping in all contexts
3. **SaveTo indentation** — don't unexpectedly apply indentation
4. **MapTo uint8** — support all integer sizes including uint8
5. **Quotation marks** — don't auto-add quotes to values
6. **Append with io.Reader** — don't prematurely close subsequent readers
7. **Empty key name** — `Section.Key("")` returns proper error instead of nil

## Key Improvements

1. **Better error messages** — include line number, column, and context in all parse errors
2. **Full type safety** — generic-aware Key methods where applicable (Go 1.21+)
3. **Comprehensive test suite** — test every LoadOptions combination
4. **Thread safety** — `File.BlockMode` uses `sync.RWMutex` correctly (fixing edge cases in original)
5. **Modern Go idioms** — proper error wrapping with `%w`, `io.ReadAll` instead of `ioutil`

## Implementation Phases

### Phase 1: Parser & Data Model
- Line-by-line INI parser with all LoadOptions
- File/Section/Key types with ordered storage
- Comment handling (line comments, inline comments, all config variants)
- Multi-line values (backslash, Python-style, triple-quoted)
- Key-value delimiter handling
- Section inheritance (parent.child)
- Boolean keys, auto-increment keys
- BOM detection and stripping
- Error types with line/column info

### Phase 2: Key Type Conversion
- All basic type conversions (Bool/Int/Int64/Uint/Uint64/Float64/Duration/Time/TimeFormat)
- Must* family (default on error)
- In* family (validate against candidates)
- Range* family (validate within bounds)
- Slice families: Strings/Ints/etc. (lenient), Strict* (error), Valid* (skip invalid)
- Shadow and nested value support
- Validate method

### Phase 3: Struct Mapping & Serialization
- MapTo / StrictMapTo with reflection
- ReflectFrom for struct → INI
- Struct tag parsing (ini, comment, delim)
- NameMapper / ValueMapper support
- WriteTo / WriteToIndent / SaveTo / SaveToIndent
- PrettyFormat / PrettySection / PrettyEqual formatting
- Reload and Append

### Phase 4: Polish & Ship
- go-ini/ini test suite ported + new tests for bug fixes
- Benchmarks comparing to go-ini/ini
- Migration guide (drop-in import path change)
- pkg.go.dev documentation
- CI/CD pipeline
- Example programs
