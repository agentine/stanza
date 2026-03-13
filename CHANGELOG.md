# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-03-13

Initial release of **stanza** — a zero-dependency INI/config file parser for Go that replaces [go-ini/ini](https://github.com/go-ini/ini).

### Added

- **Parser** (`parser.go`) — full INI parser with support for: sections `[section]`, keys `key = value`, `key: value`, `key=value`, comments (`;`, `#`), inline comments, multi-line values (backslash continuation), BOM stripping, and configurable options.
- **`LoadOptions`** — `AllowNonUniqueSections`, `AllowNonUniqueKeys`, `AllowBooleanKeys`, `InsensitiveSections`, `InsensitiveKeys`, `IgnoreInlineComment`, `SpaceBeforeInlineComment`, `PreserveSurroundingSpace`, `StrictMode`, `SkipUnrecognizableLines`, `AllowShadows`, `AllowNestedValues`, `AllowPythonMultilineValues`.
- **`File`** (`file.go`) — top-level object: `Load`, `LooseLoad`, `Empty`, `ShadowLoad`; `Section`, `HasSection`, `Sections`, `SectionsByName`, `NewSection`, `DeleteSection`, `DeleteSectionWithIndex`; `Keys`, `SaveTo`, `WriteTo`, `Reload`.
- **`Section`** (`section.go`) — `Key`, `HasKey`, `Keys`, `KeyStrings`, `NewKey`, `DeleteKey`, `ParentKeys`; section comment support.
- **`Key`** (`key.go`) — typed getters: `String`, `Int`, `Int64`, `Uint`, `Uint64`, `Float64`, `Bool`, `Duration`, `Time`; `In`, `RangeInt`, `Strings`, `Ints`, `Int64s`, `Uints`, `Uint64s`, `Float64s`, `Bools`, `StringsWithShadows`; `SetValue`, `Comment`, `AddShadow`.
- **Key type conversions** (`key_convert.go`) — full set of typed conversion methods with validation.
- **Struct mapping** (`mapper.go`, `name_mapper.go`) — `MapTo`, `MapToWithMapper`, `ReflectFrom`, `ReflectFromWithMapper`; built-in name mappers: `TitleCaseMapper`, `SnakeCaseMapper`, `AllCapsUnderscore`.
- **Benchmarks** (`bench_test.go`) — load and parse throughput benchmarks.
- **go-ini/ini compatible API** — drop-in replacement for the most common go-ini/ini usage patterns.

### Fixed

- **CRITICAL data race** — `pendingComment` moved from package-level global variable to `parser` struct field; eliminates race under concurrent `Load()` calls and cross-file comment leakage.
- `DeleteSection` with `AllowNonUniqueSections` — preserved remaining duplicate sections instead of nuking all entries for a name.
- `DeleteSectionWithIndex` — fixed stale pointer after slice removal.
- `stripInlineComment` with `SpaceBeforeInlineComment` — changed `break` to `continue` to correctly scan past unescaped `#` without stopping early.
- Section mapping methods (`MapTo` / `ReflectFrom`) — correctly handle pointer-to-struct and embedded struct fields.
