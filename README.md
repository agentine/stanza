# stanza

A zero-dependency Go INI parser. Drop-in replacement for [gopkg.in/ini.v1](https://github.com/go-ini/ini).

## Installation

```bash
go get github.com/agentine/stanza
```

## Quickstart

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

    fmt.Println(f.Section("server").Key("host").String())  // localhost
    fmt.Println(f.Section("server").Key("port").MustInt())  // 8080
    fmt.Println(f.Section("server").Key("debug").MustBool()) // true
}
```

## Migration from gopkg.in/ini.v1

Change your import:

```diff
- import "gopkg.in/ini.v1"
+ import "github.com/agentine/stanza"
```

Replace `ini.` with `stanza.` in your code. The API is fully compatible.

## API Reference

### Loading

| Function | Description |
|---|---|
| `Load(source, others...)` | Load from `[]byte`, filename (`string`), or `io.Reader` |
| `LoadSources(opts, source, others...)` | Load with custom `LoadOptions` |
| `Empty(opts...)` | Create an empty file |

### File

| Method | Description |
|---|---|
| `Section(name)` | Get section (auto-creates if missing) |
| `GetSection(name)` | Get section (returns error if missing) |
| `NewSection(name)` | Create a new section |
| `SectionStrings()` | List all section names |
| `HasSection(name)` | Check if section exists |
| `DeleteSection(name)` | Remove a section |
| `WriteTo(w)` | Write INI to `io.Writer` |
| `SaveTo(filename)` | Write INI to file |
| `SaveToIndent(filename, indent)` | Write with indentation |
| `Reload()` | Re-read original sources |
| `Append(source, others...)` | Merge additional sources |
| `MapTo(v)` | Map all sections to a struct |
| `ReflectFrom(v)` | Build file from a struct |

### Section

| Method | Description |
|---|---|
| `Key(name)` | Get key (auto-creates if missing) |
| `HasKey(name)` | Check if key exists |
| `Keys()` | List all keys |
| `KeyStrings()` | List all key names |
| `KeysHash()` | Get key-value map |
| `DeleteKey(name)` | Remove a key |
| `NewKey(name, value)` | Create or update a key |
| `HasValue(value)` | Check if any key has this value |
| `ParentKeys()` | Inherit keys from parent section |

### Key Type Conversions

| Method | Description |
|---|---|
| `String()` | Raw string value |
| `Bool()` | Parse as bool |
| `Int()` / `Int64()` | Parse as integer |
| `Uint()` / `Uint64()` | Parse as unsigned integer |
| `Float64()` | Parse as float |
| `Duration()` | Parse as `time.Duration` |
| `TimeFormat(layout)` | Parse as `time.Time` |
| `Strings(delim)` | Split as string slice |
| `Ints(delim)` | Split as int slice |
| `Float64s(delim)` | Split as float slice |
| `MustBool(defaults...)` | Bool with default fallback |
| `MustInt(defaults...)` | Int with default fallback |
| `InBool(options...)` | Constrain to allowed values |
| `InInt(options...)` | Constrain to allowed values |
| `RangeBool(min, max)` | Validate in range |
| `RangeInt(min, max)` | Validate in range |
| `ValidBool(fn)` | Custom validation |
| `StrictBool(v, ...)` | Parse with strict values |

### LoadOptions

| Option | Description |
|---|---|
| `Loose` | Ignore missing source files |
| `Insensitive` | Lowercase all section/key names |
| `InsensitiveSections` | Lowercase section names only |
| `InsensitiveKeys` | Lowercase key names only |
| `IgnoreContinuation` | Disable backslash line-continuation |
| `IgnoreInlineComment` | Treat `#` and `;` in values as literal |
| `SpaceBeforeInlineComment` | Require space before inline comment |
| `AllowBooleanKeys` | Allow keys without values |
| `AllowShadows` | Allow duplicate key names |
| `AllowNestedValues` | Enable indented sub-values |
| `AllowPythonMultilineValues` | Enable Python-style multiline values |
| `AllowNonUniqueSections` | Allow duplicate section names |
| `SkipUnrecognizableLines` | Silently skip unparseable lines |
| `UnescapeValueDoubleQuotes` | Unescape `\"` in values |
| `UnescapeValueCommentSymbols` | Unescape `\#` and `\;` in values |
| `PreserveSurroundedQuote` | Keep surrounding quotes on values |
| `KeyValueDelimiters` | Custom key-value separators (default `"="`) |
| `UnparseableSections` | Section names stored as raw text |

### Struct Mapping

```go
type Config struct {
    Name  string `ini:"name"`
    Port  int    `ini:"port"`
    Debug bool   `ini:"debug"`
}

var cfg Config
err := f.MapTo(&cfg)

// Or build from struct:
f := stanza.Empty()
err := f.ReflectFrom(&cfg)
```

Supported struct tags: `ini:"key_name"`, `ini:"-"` (skip), `ini:",omitempty"`.

### NameMapper

```go
// Convert CamelCase field names to snake_case keys:
f.NameMapper = stanza.SnackCase

// Convert CamelCase to Title_Underscore:
f.NameMapper = stanza.TitleUnderscore
```

## License

MIT
