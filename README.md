# stanza

A drop-in replacement for [gopkg.in/ini.v1](https://github.com/go-ini/ini) that fixes the `SaveTo` indentation bug and other long-standing issues. Zero dependencies. Go 1.21+.

## Why stanza?

[go-ini/ini](https://github.com/go-ini/ini) has 3,500+ stars and 5,700+ importers, but its maintenance has been irregular — a 3.5-year gap between releases, 70+ open issues, and several known bugs that affect real-world usage:

- **`SaveTo` indentation bug** — `SaveTo` unexpectedly applied indentation when it should not, corrupting round-tripped files.
- Inline comment parsing errors — `#` and `;` inside values were incorrectly stripped in some configurations.
- `MapTo` failure for `uint8` fields — all unsigned integer sizes now work correctly.
- Auto-inserted quotation marks — stanza never adds quotes you did not put there.
- `Append` with `io.Reader` — subsequent readers were prematurely closed.
- `Section.Key("")` — now returns a proper error instead of panicking.
- Data race in `pendingComment` — the parser's comment accumulator was a package-level global; stanza moves it into the per-parse struct, eliminating the race under concurrent `Load()` calls.

stanza matches the go-ini/ini public API exactly. The migration is a one-line import change.

## Installation

```bash
go get github.com/agentine/stanza
```

## Quick start

```go
package main

import (
    "fmt"
    "github.com/agentine/stanza"
)

func main() {
    f, err := stanza.Load([]byte(`
[server]
host = localhost
port = 8080
debug = true
`))
    if err != nil {
        panic(err)
    }

    fmt.Println(f.Section("server").Key("host").String())   // localhost
    fmt.Println(f.Section("server").Key("port").MustInt())  // 8080
    fmt.Println(f.Section("server").Key("debug").MustBool()) // true
}
```

## Loading

### Functions

```go
// Load from a file path (string), []byte, io.Reader, or io.ReadCloser.
// Pass multiple sources to merge them in order.
f, err := stanza.Load("app.ini")
f, err := stanza.Load("base.ini", "override.ini")
f, err := stanza.Load([]byte("[section]\nkey=val"))
f, err := stanza.Load(reader)

// Load with explicit options.
f, err := stanza.LoadSources(stanza.LoadOptions{
    IgnoreInlineComment: true,
    AllowBooleanKeys:    true,
}, "app.ini")

// Convenience wrappers.
f, err := stanza.LooseLoad("app.ini", "missing-optional.ini") // ignores missing files
f, err := stanza.InsensitiveLoad("app.ini")                   // all names lowercased
f, err := stanza.ShadowLoad("app.ini")                        // allow duplicate keys

// Create an empty file (optionally with options).
f := stanza.Empty()
f := stanza.Empty(stanza.LoadOptions{AllowBooleanKeys: true})
```

### Reload and append

```go
// Re-read all original file-path sources.
err := f.Reload()

// Merge additional sources into an existing file.
err := f.Append("extra.ini", moreBytes)
```

## Data model

### File

```
File
 └── Section ("DEFAULT")
 └── Section ("server")
      └── Key ("host")  →  "localhost"
      └── Key ("port")  →  "8080"
```

A `File` always contains a default section (name `"DEFAULT"`) for keys that appear before any `[section]` header.

```go
// Section access
sec := f.Section("server")          // auto-creates if missing
sec, err := f.GetSection("server")  // returns error if missing
secs := f.Sections()                // all sections, in order
names := f.SectionStrings()         // section names, in order
ok := f.HasSection("server")

// Creating and removing sections
sec, err := f.NewSection("cache")
err = f.NewSections("a", "b", "c")
sec, err := f.NewRawSection("raw", "arbitrary body text")
f.DeleteSection("cache")
err = f.DeleteSectionWithIndex("dup", 1) // for AllowNonUniqueSections

// Non-unique sections (requires AllowNonUniqueSections option)
secs, err := f.SectionsByName("repeated")
sec := f.SectionWithIndex("repeated", 2)

// Child sections  ([parent.child] hierarchy)
children := f.ChildSections("database")

// Struct fields
f.BlockMode   = true        // default; thread-safe via sync.RWMutex
f.NameMapper  = stanza.SnackCase
f.ValueMapper = func(s string) string { return strings.TrimSpace(s) }

// Serialization
n, err := f.WriteTo(w)
n, err := f.WriteToIndent(w, "\t")
err = f.SaveTo("out.ini")
err = f.SaveToIndent("out.ini", "\t")

// Struct mapping
err = f.MapTo(&cfg)
err = f.StrictMapTo(&cfg)
err = f.ReflectFrom(&cfg)
```

### Section

```go
sec.Name()                      // "server"
sec.Comment                     // read/write comment string
sec.Body()                      // raw body (for unparseable sections)
sec.SetBody("raw text")

// Key access
k := sec.Key("port")            // auto-creates empty key if missing
k, err := sec.GetKey("port")   // returns error if missing
ok := sec.HasKey("port")
keys := sec.Keys()
names := sec.KeyStrings()
m := sec.KeysHash()             // map[string]string

ok := sec.HasValue("8080")
sec.DeleteKey("port")

// Creating keys
k, err := sec.NewKey("timeout", "30s")
k, err := sec.NewBooleanKey("verbose")  // a key with no value

// Inheritance
parentKeys := sec.ParentKeys()     // keys from [parent] when name is "parent.child"
children := sec.ChildSections()

// Struct mapping
err = sec.MapTo(&cfg)
err = sec.StrictMapTo(&cfg)
err = sec.ReflectFrom(&cfg)
```

### Key

```go
k.Name()           // key name
k.Value()          // raw string value (same as String())
k.String()         // raw string value
k.SetValue("new")
k.Comment          // read/write comment string
k.IsBooleanKey()   // true for keys with no "=" (AllowBooleanKeys)

// Shadow and nested values
vals := k.ValueWithShadows()          // all values including shadows
err = k.AddShadow("extra-value")
nested := k.NestedValues()            // AWS-style indented sub-values
err = k.AddNestedValue("sub")

// Custom transformation
result := k.Validate(func(s string) string {
    return strings.TrimSpace(s)
})
```

## Parser options

All 24 `LoadOptions` fields:

| Field | Default | Description |
|---|---|---|
| `Loose` | `false` | Ignore missing source files instead of returning an error |
| `Insensitive` | `false` | Force all section and key names to lowercase |
| `InsensitiveSections` | `false` | Force section names to lowercase only |
| `InsensitiveKeys` | `false` | Force key names to lowercase only |
| `IgnoreContinuation` | `false` | Disable backslash `\` line-continuation |
| `IgnoreInlineComment` | `false` | Treat `#` and `;` inside values as literal characters |
| `SkipUnrecognizableLines` | `false` | Silently skip lines that cannot be parsed |
| `ShortCircuit` | `false` | Stop loading after the first source |
| `AllowBooleanKeys` | `false` | Allow keys with no value (e.g. `verbose` with no `=`) |
| `AllowShadows` | `false` | Allow duplicate key names within a section |
| `AllowNestedValues` | `false` | AWS-style indented continuation values |
| `AllowPythonMultilineValues` | `false` | Python configparser-style indented multiline values |
| `SpaceBeforeInlineComment` | `false` | Require a space before `#` or `;` to treat them as inline comments |
| `UnescapeValueDoubleQuotes` | `false` | Unescape `\"` sequences in values |
| `UnescapeValueCommentSymbols` | `false` | Unescape `\#` and `\;` in values |
| `PreserveSurroundedQuote` | `false` | Keep surrounding `"..."` or `'...'` quotes on values |
| `AllowNonUniqueSections` | `false` | Allow multiple sections with the same name |
| `AllowDuplicateShadowValues` | `false` | Allow duplicate shadow values for the same key |
| `UnparseableSections` | `nil` | Section names whose bodies are stored as raw strings |
| `KeyValueDelimiters` | `"=:"` | Characters that separate keys from values |
| `KeyValueDelimiterOnWrite` | `"="` | Delimiter written between key and value |
| `ChildSectionDelimiter` | `"."` | Separator for `parent.child` section hierarchy |
| `DebugFunc` | `nil` | Callback that receives debug messages during parsing |
| `ReaderBufferSize` | `0` | Buffer size for reading sources |

Example:

```go
f, err := stanza.LoadSources(stanza.LoadOptions{
    IgnoreInlineComment:  true,
    AllowBooleanKeys:     true,
    AllowShadows:         true,
    SpaceBeforeInlineComment: true,
    KeyValueDelimiters:   "=",
}, "app.ini")
```

## Key type getters

### Fallible conversions

```go
v, err := k.Bool()
v, err := k.Int()
v, err := k.Int64()
v, err := k.Uint()
v, err := k.Uint64()
v, err := k.Float64()
v, err := k.Duration()          // time.Duration (e.g. "1h30m")
v, err := k.Time()              // time.Time, parsed as RFC 3339
v, err := k.TimeFormat(layout)  // time.Time with custom layout
```

Boolean values recognised (case-insensitive): `1`, `t`, `true`, `yes`, `y`, `on` → `true`; `0`, `f`, `false`, `no`, `n`, `off` → `false`.

### Must variants — return default on error

```go
k.MustString("fallback")
k.MustBool()           // defaults to false
k.MustBool(true)       // explicit default
k.MustInt()
k.MustInt(42)
k.MustInt64()
k.MustUint()
k.MustUint64()
k.MustFloat64()
k.MustDuration()
k.MustTime()
k.MustTimeFormat(layout)
k.MustTimeFormat(layout, defaultTime)
```

## Slice families

All slice methods split the value on a delimiter and trim whitespace from each element.

### Lenient — skip invalid entries

```go
k.Strings(",")      // []string
k.Ints(",")         // []int    — invalid entries silently dropped
k.Int64s(",")
k.Uints(",")
k.Uint64s(",")
k.Float64s(",")
k.Bools(",")
k.Times(",")        // []time.Time, RFC 3339
k.TimesFormat(layout, ",")

// Include shadow values in the split:
k.StringsWithShadows(",")
```

### Strict — return error on first invalid entry

```go
vals, err := k.StrictInts(",")
vals, err := k.StrictInt64s(",")
vals, err := k.StrictUints(",")
vals, err := k.StrictUint64s(",")
vals, err := k.StrictFloat64s(",")
vals, err := k.StrictBools(",")
vals, err := k.StrictTimes(",")
vals, err := k.StrictTimesFormat(layout, ",")
```

### Valid — alias for lenient (skip invalid, explicit intent)

```go
k.ValidInts(",")
k.ValidInt64s(",")
k.ValidUints(",")
k.ValidUint64s(",")
k.ValidFloat64s(",")
k.ValidBools(",")
k.ValidTimes(",")
k.ValidTimesFormat(layout, ",")
```

## Validation

### In — constrain to an allowed set

```go
env := k.In("production", []string{"development", "staging", "production"})
port := k.InInt(8080, []int{80, 443, 8080, 8443})
k.InInt64(defaultVal, candidates)
k.InUint(defaultVal, candidates)
k.InUint64(defaultVal, candidates)
k.InFloat64(defaultVal, candidates)
k.InTime(defaultVal, candidates)
k.InTimeFormat(layout, defaultVal, candidates)
```

### Range — constrain to a numeric or time interval

```go
port := k.RangeInt(8080, 1, 65535)         // returns defaultVal if out of range
k.RangeInt64(defaultVal, min, max)
k.RangeFloat64(defaultVal, min, max)
k.RangeTime(defaultVal, min, max)
k.RangeTimeFormat(layout, defaultVal, min, max)
```

### Validate — custom transform

```go
v := k.Validate(func(s string) string {
    if s == "" {
        return "default"
    }
    return strings.ToUpper(s)
})
```

## Struct mapping

### MapTo — INI to struct

```go
type ServerConfig struct {
    Host    string        `ini:"host"`
    Port    int           `ini:"port"`
    Debug   bool          `ini:"debug,omitempty"`
    Timeout time.Duration `ini:"timeout"`
    Tags    []string      `ini:"tags" ini-delim:","`
}

var cfg ServerConfig
err := f.Section("server").MapTo(&cfg)

// Or map the whole file (top-level fields come from DEFAULT section):
type AppConfig struct {
    Server ServerConfig `ini:"server"` // → [server] section
}
var app AppConfig
err := f.MapTo(&app)
```

`StrictMapTo` returns an error if any INI key in the section has no corresponding struct field.

```go
err := f.StrictMapTo(&app)
err := f.Section("server").StrictMapTo(&cfg)
```

### ReflectFrom — struct to INI

```go
f := stanza.Empty()
err := f.ReflectFrom(&app)
err = f.SaveTo("out.ini")
```

### Struct tags

| Tag | Example | Description |
|---|---|---|
| `ini:"name"` | `ini:"host"` | Override the key/section name |
| `ini:"-"` | `ini:"-"` | Skip this field entirely |
| `ini:",omitempty"` | `ini:"port,omitempty"` | Skip on `ReflectFrom` when value is zero |
| `ini-delim:"sep"` | `ini-delim:"|"` | Delimiter for slice fields (default `,`) |
| `ini-comment:"text"` | `ini-comment:"Server port"` | Comment written above the key or section |

A struct field whose type is a struct (other than `time.Time`) maps to a same-named INI section. A pointer-to-struct field is only mapped if that section exists in the file.

### Package-level mapping helpers

```go
// Load + MapTo in one call.
err := stanza.MapTo(&cfg, "app.ini")
err := stanza.MapToWithMapper(&cfg, stanza.SnackCase, "app.ini")
err := stanza.StrictMapTo(&cfg, "app.ini")
err := stanza.StrictMapToWithMapper(&cfg, stanza.SnackCase, "app.ini")

// Load + ReflectFrom in one call (builds a File from a struct).
err := stanza.ReflectFrom(&cfg, "app.ini")
err := stanza.ReflectFromWithMapper(&cfg, stanza.SnackCase, "app.ini")
```

## Serialization

```go
// Write to any io.Writer.
n, err := f.WriteTo(os.Stdout)
n, err := f.WriteToIndent(os.Stdout, "\t")

// Write to a file (the SaveTo indentation bug in go-ini/ini is fixed here).
err = f.SaveTo("out.ini")
err = f.SaveToIndent("out.ini", "\t")
```

`SaveTo` always writes without indentation. `SaveToIndent` writes with the given indent string prepended to every key line within a section.

## Package-level variables

These control default formatting behaviour for all files.

```go
stanza.DefaultSection = "DEFAULT"  // name of the implicit default section (constant)

stanza.LineBreak      = "\n"       // "\r\n" on Windows
stanza.DefaultHeader  = false      // write [DEFAULT] header even when empty
stanza.PrettySection  = true       // blank line before each section header
stanza.PrettyFormat   = true       // blank line between sections
stanza.PrettyEqual    = false      // use " = " instead of "=" for key-value pairs
stanza.DefaultFormatLeft  = ""     // string prepended to every key name on write
stanza.DefaultFormatRight = ""     // string appended to every key name on write
```

## Name mappers

```go
// SnackCase: "CamelCase" → "camel_case"
f.NameMapper = stanza.SnackCase

// TitleUnderscore: "CamelCase" → "Camel_Case"
f.NameMapper = stanza.TitleUnderscore

// Custom mapper:
f.NameMapper = func(s string) string { return strings.ToLower(s) }
```

A `ValueMapper` transforms every key value during loading:

```go
f.ValueMapper = func(s string) string { return strings.TrimSpace(s) }
```

## INI format features

**Comments** — `#` and `;` begin full-line comments. Inline comments are stripped by default; use `SpaceBeforeInlineComment` to require a preceding space, or `IgnoreInlineComment` to treat `#`/`;` as literals.

**Multi-line values** — three styles:

```ini
# 1. Backslash continuation (disabled by IgnoreContinuation)
value = first line \
        second line

# 2. Triple-quoted
key = """
line one
line two
"""

# 3. Python configparser style (requires AllowPythonMultilineValues)
key = first line
    second line
    third line
```

**Boolean keys** — keys with no `=` (requires `AllowBooleanKeys`):

```ini
[flags]
verbose
readonly
```

**Auto-increment keys** — key name `-` becomes `#1`, `#2`, etc.

**Section inheritance** — `[parent.child]` sections inherit keys from `[parent]` via `sec.ParentKeys()`. The delimiter is configurable via `ChildSectionDelimiter`.

**Shadow keys** — duplicate key names in the same section (requires `AllowShadows`). Retrieve all values with `k.ValueWithShadows()`.

**BOM handling** — UTF-8, UTF-16 LE, and UTF-16 BE byte-order marks are detected and stripped automatically.

## Error types

```go
// Returned when a key-value line has no recognised delimiter.
type ErrDelimiterNotFound struct{ Line string }
stanza.IsErrDelimiterNotFound(err) bool

// Returned when a key name is empty.
type ErrEmptyKeyName struct{ Line string }
stanza.IsErrEmptyKeyName(err) bool
```

## Migration from gopkg.in/ini.v1

The import path is the only thing you need to change. Every exported symbol has the same name and signature.

**sed one-liner:**

```bash
find . -name '*.go' | xargs sed -i 's|gopkg.in/ini.v1|github.com/agentine/stanza|g'
```

**Manual diff:**

```diff
-import "gopkg.in/ini.v1"
+import "github.com/agentine/stanza"
```

Then replace the package qualifier:

```diff
-ini.Load(...)
+stanza.Load(...)
```

Or keep the old qualifier with an alias:

```go
import ini "github.com/agentine/stanza"
```

## License

MIT
