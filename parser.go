package stanza

import (
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"
)

// ---------------------------------------------------------------------------
// Package-level loading functions
// ---------------------------------------------------------------------------

// Load loads an INI file from one or more sources.
// Source types: string (file path), []byte, io.Reader, io.ReadCloser.
func Load(source interface{}, others ...interface{}) (*File, error) {
	return LoadSources(LoadOptions{}, source, others...)
}

// LoadSources loads INI data with explicit options.
func LoadSources(opts LoadOptions, source interface{}, others ...interface{}) (*File, error) {
	f := newFile(opts)
	srcs := append([]interface{}{source}, others...)
	for _, src := range srcs {
		if err := f.loadSource(src); err != nil {
			return nil, err
		}
		if opts.ShortCircuit {
			break
		}
	}
	return f, nil
}

// LooseLoad is a convenience for Load with Loose enabled.
func LooseLoad(source interface{}, others ...interface{}) (*File, error) {
	return LoadSources(LoadOptions{Loose: true}, source, others...)
}

// InsensitiveLoad is a convenience for Load with Insensitive enabled.
func InsensitiveLoad(source interface{}, others ...interface{}) (*File, error) {
	return LoadSources(LoadOptions{Insensitive: true}, source, others...)
}

// ShadowLoad is a convenience for Load with AllowShadows enabled.
func ShadowLoad(source interface{}, others ...interface{}) (*File, error) {
	return LoadSources(LoadOptions{AllowShadows: true}, source, others...)
}

// ---------------------------------------------------------------------------
// Parser internals
// ---------------------------------------------------------------------------

// parseINI parses raw INI data into the given File.
func parseINI(f *File, data []byte) error {
	// Strip UTF-8 BOM if present.
	data = stripBOM(data)

	opts := f.options
	delimiters := opts.KeyValueDelimiters
	if delimiters == "" {
		delimiters = "=:"
	}

	unparseableSet := make(map[string]bool)
	for _, name := range opts.UnparseableSections {
		unparseableSet[name] = true
	}

	p := &parser{
		f:              f,
		opts:           opts,
		delimiters:     delimiters,
		unparseable:    unparseableSet,
		currentSection: f.sectionsByName[DefaultSection],
		autoIncrID:     0,
	}

	return p.parse(data)
}

type parser struct {
	f              *File
	opts           LoadOptions
	delimiters     string
	unparseable    map[string]bool
	currentSection *Section
	autoIncrID     int
	pendingComment string
}

func (p *parser) parse(data []byte) error {
	lines := splitLines(data)
	i := 0
	for i < len(lines) {
		line := lines[i]
		i++

		// Skip empty lines.
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Full-line comment.
		if trimmed[0] == '#' || trimmed[0] == ';' {
			// Check if this is a comment for the next section/key.
			p.handleComment(trimmed)
			continue
		}

		// Section header.
		if trimmed[0] == '[' {
			if err := p.parseSection(trimmed); err != nil {
				if p.opts.SkipUnrecognizableLines {
					continue
				}
				return err
			}
			// If this is an unparseable section, collect raw body.
			if p.unparseable[p.currentSection.name] {
				p.currentSection.isRaw = true
				var body strings.Builder
				for i < len(lines) {
					next := strings.TrimSpace(lines[i])
					if len(next) > 0 && next[0] == '[' {
						break
					}
					body.WriteString(lines[i])
					body.WriteString("\n")
					i++
				}
				p.currentSection.body = strings.TrimRight(body.String(), "\n")
			}
			continue
		}

		// Key-value line.
		consumed, err := p.parseKeyValue(line, lines, i)
		if err != nil {
			if p.opts.SkipUnrecognizableLines {
				continue
			}
			return err
		}
		i += consumed
	}
	return nil
}

func (p *parser) handleComment(line string) {
	if p.pendingComment != "" {
		p.pendingComment += "\n" + line
	} else {
		p.pendingComment = line
	}
}

func (p *parser) consumeComment() string {
	c := p.pendingComment
	p.pendingComment = ""
	return c
}

func (p *parser) parseSection(line string) error {
	// Find the closing ].
	end := strings.IndexByte(line, ']')
	if end < 0 {
		return fmt.Errorf("unclosed section header: %s", line)
	}
	name := line[1:end]

	if p.opts.Insensitive || p.opts.InsensitiveSections {
		name = toLower(name)
	}

	comment := p.consumeComment()

	lookupName := p.f.sectionLookupName(name)
	if existing, ok := p.f.sectionsByName[lookupName]; ok {
		if !p.f.options.AllowNonUniqueSections {
			p.currentSection = existing
			if comment != "" && existing.Comment == "" {
				existing.Comment = comment
			}
			return nil
		}
	}

	sec := newSection(p.f, name)
	sec.Comment = comment
	p.f.sections = append(p.f.sections, sec)
	if _, ok := p.f.sectionsByName[lookupName]; !ok {
		p.f.sectionsByName[lookupName] = sec
	}
	p.f.sectionList[lookupName] = append(p.f.sectionList[lookupName], sec)
	p.currentSection = sec
	return nil
}

func (p *parser) parseKeyValue(line string, lines []string, nextIdx int) (int, error) {
	consumed := 0

	// Handle backslash line continuation.
	if !p.opts.IgnoreContinuation {
		for strings.HasSuffix(strings.TrimRight(line, " \t"), "\\") {
			line = strings.TrimRight(line, " \t")
			line = line[:len(line)-1]
			if nextIdx+consumed < len(lines) {
				line += strings.TrimSpace(lines[nextIdx+consumed])
				consumed++
			} else {
				break
			}
		}
	}

	// Find the delimiter.
	delimIdx := -1
	delimChar := byte('=')
	for i := 0; i < len(line); i++ {
		c := line[i]
		if strings.IndexByte(p.delimiters, c) >= 0 {
			delimIdx = i
			delimChar = c
			break
		}
	}
	_ = delimChar

	if delimIdx < 0 {
		// No delimiter found.
		if p.opts.AllowBooleanKeys {
			keyName := strings.TrimSpace(line)
			if keyName == "" {
				return consumed, ErrEmptyKeyName{Line: line}
			}
			keyName = p.applyKeyNameMapping(keyName)
			comment := p.consumeComment()
			k, err := p.currentSection.NewBooleanKey(keyName)
			if err != nil {
				return consumed, err
			}
			k.Comment = comment
			return consumed, nil
		}
		return consumed, ErrDelimiterNotFound{Line: line}
	}

	keyName := strings.TrimSpace(line[:delimIdx])
	if keyName == "" {
		return consumed, ErrEmptyKeyName{Line: line}
	}

	// Auto-increment key: "-" becomes "#1", "#2", etc.
	if keyName == "-" {
		p.autoIncrID++
		keyName = fmt.Sprintf("#%d", p.autoIncrID)
	}

	keyName = p.applyKeyNameMapping(keyName)

	value := strings.TrimSpace(line[delimIdx+1:])

	// Strip inline comments (if enabled).
	value = p.stripInlineComment(value)

	// Handle quoted values.
	value = p.handleQuotedValue(value)

	// Handle triple-quoted values.
	if strings.HasPrefix(value, `"""`) {
		value = value[3:]
		if idx := strings.Index(value, `"""`); idx >= 0 {
			value = value[:idx]
		} else {
			// Multi-line triple-quoted value.
			var buf strings.Builder
			buf.WriteString(value)
			for nextIdx+consumed < len(lines) {
				nextLine := lines[nextIdx+consumed]
				consumed++
				if idx := strings.Index(nextLine, `"""`); idx >= 0 {
					buf.WriteString("\n")
					buf.WriteString(nextLine[:idx])
					break
				}
				buf.WriteString("\n")
				buf.WriteString(nextLine)
			}
			value = buf.String()
		}
	}

	// Handle Python-style multiline values.
	if p.opts.AllowPythonMultilineValues {
		for nextIdx+consumed < len(lines) {
			nextLine := lines[nextIdx+consumed]
			if len(nextLine) == 0 {
				break
			}
			// If the next line starts with whitespace, it's a continuation.
			if nextLine[0] == ' ' || nextLine[0] == '\t' {
				value += "\n" + strings.TrimSpace(nextLine)
				consumed++
			} else {
				break
			}
		}
	}

	// Handle nested values (AWS-style).
	var nestedVals []string
	if p.opts.AllowNestedValues {
		for nextIdx+consumed < len(lines) {
			nextLine := lines[nextIdx+consumed]
			if len(nextLine) == 0 {
				break
			}
			if nextLine[0] == ' ' || nextLine[0] == '\t' {
				nestedVals = append(nestedVals, strings.TrimSpace(nextLine))
				consumed++
			} else {
				break
			}
		}
	}

	// Apply unescape options.
	if p.opts.UnescapeValueDoubleQuotes {
		value = strings.ReplaceAll(value, `\"`, `"`)
	}
	if p.opts.UnescapeValueCommentSymbols {
		value = strings.ReplaceAll(value, `\#`, `#`)
		value = strings.ReplaceAll(value, `\;`, `;`)
	}

	// Apply value mapper.
	if p.f.ValueMapper != nil {
		value = p.f.ValueMapper(value)
	}

	comment := p.consumeComment()

	k, err := p.currentSection.NewKey(keyName, value)
	if err != nil {
		return consumed, err
	}
	k.Comment = comment
	k.nestedValues = nestedVals

	if p.opts.DebugFunc != nil {
		p.opts.DebugFunc(fmt.Sprintf("parsed key %q = %q in section [%s]", keyName, value, p.currentSection.name))
	}

	return consumed, nil
}

func (p *parser) applyKeyNameMapping(name string) string {
	if p.opts.Insensitive || p.opts.InsensitiveKeys {
		name = toLower(name)
	}
	return name
}

func (p *parser) stripInlineComment(value string) string {
	if p.opts.IgnoreInlineComment {
		return value
	}

	// If the value is quoted, don't strip.
	if len(value) > 1 && (value[0] == '"' || value[0] == '\'') {
		return value
	}

	bestIdx := -1
	for _, c := range []byte{'#', ';'} {
		// Scan for the first UNESCAPED occurrence of the comment char.
		offset := 0
		for {
			idx := strings.IndexByte(value[offset:], c)
			if idx < 0 {
				break
			}
			idx += offset
			// Check for escaped comment.
			if idx > 0 && value[idx-1] == '\\' {
				offset = idx + 1
				continue
			}
			if p.opts.SpaceBeforeInlineComment {
				if idx > 0 && value[idx-1] == ' ' {
					if bestIdx < 0 || idx < bestIdx {
						bestIdx = idx
					}
					break
				}
				// No preceding space — keep scanning for a later match.
				offset = idx + 1
				continue
			} else {
				if bestIdx < 0 || idx < bestIdx {
					bestIdx = idx
				}
				break
			}
		}
	}
	if bestIdx >= 0 {
		if p.opts.SpaceBeforeInlineComment {
			value = strings.TrimRight(value[:bestIdx-1], " \t")
		} else {
			value = strings.TrimRight(value[:bestIdx], " \t")
		}
	}
	return value
}

func (p *parser) handleQuotedValue(value string) string {
	if len(value) < 2 {
		return value
	}

	if p.opts.PreserveSurroundedQuote {
		return value
	}

	// Strip surrounding double or single quotes.
	if (value[0] == '"' && value[len(value)-1] == '"') ||
		(value[0] == '\'' && value[len(value)-1] == '\'') {
		return value[1 : len(value)-1]
	}
	return value
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func splitLines(data []byte) []string {
	// Normalize line endings.
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	data = bytes.ReplaceAll(data, []byte("\r"), []byte("\n"))
	s := string(data)
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	// Remove trailing empty line from Split.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func stripBOM(data []byte) []byte {
	// UTF-8 BOM: EF BB BF
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return data[3:]
	}
	// UTF-16 LE BOM: FF FE
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		return decodeUTF16LE(data[2:])
	}
	// UTF-16 BE BOM: FE FF
	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		return decodeUTF16BE(data[2:])
	}
	return data
}

func decodeUTF16LE(data []byte) []byte {
	var buf bytes.Buffer
	for i := 0; i+1 < len(data); i += 2 {
		r := rune(data[i]) | rune(data[i+1])<<8
		if r >= 0xD800 && r <= 0xDBFF && i+3 < len(data) {
			lo := rune(data[i+2]) | rune(data[i+3])<<8
			if lo >= 0xDC00 && lo <= 0xDFFF {
				r = (r-0xD800)*0x400 + (lo - 0xDC00) + 0x10000
				i += 2
			}
		}
		var b [utf8.UTFMax]byte
		n := utf8.EncodeRune(b[:], r)
		buf.Write(b[:n])
	}
	return buf.Bytes()
}

func decodeUTF16BE(data []byte) []byte {
	var buf bytes.Buffer
	for i := 0; i+1 < len(data); i += 2 {
		r := rune(data[i])<<8 | rune(data[i+1])
		if r >= 0xD800 && r <= 0xDBFF && i+3 < len(data) {
			lo := rune(data[i+2])<<8 | rune(data[i+3])
			if lo >= 0xDC00 && lo <= 0xDFFF {
				r = (r-0xD800)*0x400 + (lo - 0xDC00) + 0x10000
				i += 2
			}
		}
		var b [utf8.UTFMax]byte
		n := utf8.EncodeRune(b[:], r)
		buf.Write(b[:n])
	}
	return buf.Bytes()
}
