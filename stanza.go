package stanza

import (
	"fmt"
	"runtime"
)

// DefaultSection is the name of the implicit default section that holds
// keys appearing before any explicit [section] header.
const DefaultSection = "DEFAULT"

// LineBreak is the line ending used when writing INI files.
// It defaults to "\n" on Unix and "\r\n" on Windows.
var LineBreak = detectLineBreak()

func detectLineBreak() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

// DefaultHeader controls whether a [DEFAULT] header is written for
// keys in the default section.
var DefaultHeader = false

// PrettySection adds a blank line before each section header when writing.
var PrettySection = true

// PrettyFormat adds a blank line between sections when writing.
var PrettyFormat = true

// PrettyEqual uses " = " instead of "=" when writing key-value pairs.
var PrettyEqual = false

// DefaultFormatLeft is the string prepended to keys when writing.
var DefaultFormatLeft = ""

// DefaultFormatRight is the string appended to keys when writing.
var DefaultFormatRight = ""

// NameMapper maps struct field names to INI key names.
type NameMapper func(string) string

// ValueMapper transforms key values during loading.
type ValueMapper func(string) string

// DebugFunc is a callback for debug messages during parsing.
type DebugFunc func(message string)

// LoadOptions configures the parser's behaviour.
type LoadOptions struct {
	// Loose ignores missing source files instead of returning an error.
	Loose bool
	// Insensitive forces all section and key names to lowercase.
	Insensitive bool
	// InsensitiveSections forces section names to lowercase.
	InsensitiveSections bool
	// InsensitiveKeys forces key names to lowercase.
	InsensitiveKeys bool
	// IgnoreContinuation disables backslash line-continuation.
	IgnoreContinuation bool
	// IgnoreInlineComment treats # and ; inside values as literal characters.
	IgnoreInlineComment bool
	// SkipUnrecognizableLines silently skips lines that cannot be parsed.
	SkipUnrecognizableLines bool
	// ShortCircuit stops loading after the first source.
	ShortCircuit bool
	// AllowBooleanKeys permits keys without values (e.g., "flag" with no "=").
	AllowBooleanKeys bool
	// AllowShadows permits duplicate key names within a section.
	AllowShadows bool
	// AllowNestedValues enables AWS-style indented continuation values.
	AllowNestedValues bool
	// AllowPythonMultilineValues enables Python configparser-style
	// indented multiline values.
	AllowPythonMultilineValues bool
	// SpaceBeforeInlineComment requires a space before # or ;
	// for them to be recognised as inline comments.
	SpaceBeforeInlineComment bool
	// UnescapeValueDoubleQuotes unescapes \" sequences in values.
	UnescapeValueDoubleQuotes bool
	// UnescapeValueCommentSymbols unescapes \# and \; in values.
	UnescapeValueCommentSymbols bool
	// UnparseableSections lists section names whose bodies should be
	// stored as raw strings instead of parsed as key-value pairs.
	UnparseableSections []string
	// KeyValueDelimiters is the set of characters that separate keys
	// from values (default "=:").
	KeyValueDelimiters string
	// KeyValueDelimiterOnWrite is the delimiter used when writing
	// key-value pairs (default "=").
	KeyValueDelimiterOnWrite string
	// ChildSectionDelimiter is the separator for parent.child section
	// hierarchy (default ".").
	ChildSectionDelimiter string
	// PreserveSurroundedQuote keeps surrounding quotes on values
	// instead of stripping them.
	PreserveSurroundedQuote bool
	// DebugFunc receives debug messages during parsing.
	DebugFunc DebugFunc
	// ReaderBufferSize sets the buffer size for reading sources.
	ReaderBufferSize int
	// AllowNonUniqueSections permits multiple sections with the same name.
	AllowNonUniqueSections bool
	// AllowDuplicateShadowValues permits duplicate shadow values for
	// the same key.
	AllowDuplicateShadowValues bool
}

// SnackCase is a NameMapper that converts CamelCase to snake_case.
var SnackCase NameMapper

// TitleUnderscore is a NameMapper that inserts underscores before
// uppercase letters in CamelCase names.
var TitleUnderscore NameMapper

// ---------------------------------------------------------------------------
// Error types
// ---------------------------------------------------------------------------

// ErrDelimiterNotFound is returned when a key-value line lacks a
// recognised delimiter.
type ErrDelimiterNotFound struct {
	Line string
}

func (e ErrDelimiterNotFound) Error() string {
	return fmt.Sprintf("key-value delimiter not found: %s", e.Line)
}

// IsErrDelimiterNotFound reports whether err is an ErrDelimiterNotFound.
func IsErrDelimiterNotFound(err error) bool {
	_, ok := err.(ErrDelimiterNotFound)
	return ok
}

// ErrEmptyKeyName is returned when a line has an empty key name.
type ErrEmptyKeyName struct {
	Line string
}

func (e ErrEmptyKeyName) Error() string {
	return fmt.Sprintf("empty key name: %s", e.Line)
}

// IsErrEmptyKeyName reports whether err is an ErrEmptyKeyName.
func IsErrEmptyKeyName(err error) bool {
	_, ok := err.(ErrEmptyKeyName)
	return ok
}
