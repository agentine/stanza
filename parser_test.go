package stanza

import (
	"bytes"
	"strings"
	"testing"
)

const sampleINI = `# comment
[section]
key = value
name = stanza
flag = true

[database]
host = localhost
port = 5432
# inline db name
dbname = mydb
`

func TestLoadBytes(t *testing.T) {
	f, err := Load([]byte(sampleINI))
	if err != nil {
		t.Fatal(err)
	}

	if !f.HasSection("section") {
		t.Fatal("section not found")
	}
	if !f.HasSection("database") {
		t.Fatal("database not found")
	}

	sec := f.Section("section")
	if sec.Key("key").String() != "value" {
		t.Errorf("got %q, want %q", sec.Key("key").String(), "value")
	}
	if sec.Key("name").String() != "stanza" {
		t.Errorf("got %q, want %q", sec.Key("name").String(), "stanza")
	}

	db := f.Section("database")
	if db.Key("host").String() != "localhost" {
		t.Errorf("got %q, want %q", db.Key("host").String(), "localhost")
	}
	if db.Key("port").String() != "5432" {
		t.Errorf("got %q, want %q", db.Key("port").String(), "5432")
	}
	if db.Key("dbname").String() != "mydb" {
		t.Errorf("got %q, want %q", db.Key("dbname").String(), "mydb")
	}
}

func TestSectionStrings(t *testing.T) {
	f, err := Load([]byte(sampleINI))
	if err != nil {
		t.Fatal(err)
	}
	names := f.SectionStrings()
	// DEFAULT, section, database
	if len(names) != 3 {
		t.Fatalf("expected 3 sections, got %d: %v", len(names), names)
	}
}

func TestDefaultSection(t *testing.T) {
	ini := `key = global_value
[section]
key = section_value
`
	f, err := Load([]byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	if f.Section(DefaultSection).Key("key").String() != "global_value" {
		t.Error("default section key mismatch")
	}
	if f.Section("section").Key("key").String() != "section_value" {
		t.Error("section key mismatch")
	}
}

func TestBooleanKeys(t *testing.T) {
	ini := `[section]
flag
enabled
`
	f, err := LoadSources(LoadOptions{AllowBooleanKeys: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	k, err := f.Section("section").GetKey("flag")
	if err != nil {
		t.Fatal(err)
	}
	if !k.IsBooleanKey() {
		t.Error("expected boolean key")
	}
}

func TestAutoIncrementKeys(t *testing.T) {
	ini := `[section]
- = a
- = b
- = c
`
	f, err := Load([]byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	sec := f.Section("section")
	if sec.Key("#1").String() != "a" {
		t.Errorf("auto-incr #1: got %q", sec.Key("#1").String())
	}
	if sec.Key("#2").String() != "b" {
		t.Errorf("auto-incr #2: got %q", sec.Key("#2").String())
	}
	if sec.Key("#3").String() != "c" {
		t.Errorf("auto-incr #3: got %q", sec.Key("#3").String())
	}
}

func TestColonDelimiter(t *testing.T) {
	ini := `[section]
key: value
`
	f, err := Load([]byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	if f.Section("section").Key("key").String() != "value" {
		t.Error("colon delimiter not parsed")
	}
}

func TestInlineComments(t *testing.T) {
	ini := `[section]
key = value # comment
key2 = value2 ; another comment
`
	f, err := Load([]byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	if f.Section("section").Key("key").String() != "value" {
		t.Errorf("got %q", f.Section("section").Key("key").String())
	}
}

func TestIgnoreInlineComment(t *testing.T) {
	ini := `[section]
key = value # not a comment
`
	f, err := LoadSources(LoadOptions{IgnoreInlineComment: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	if f.Section("section").Key("key").String() != "value # not a comment" {
		t.Errorf("got %q", f.Section("section").Key("key").String())
	}
}

func TestSpaceBeforeInlineComment(t *testing.T) {
	ini := `[section]
key = value#not-a-comment
key2 = value2 # comment
`
	f, err := LoadSources(LoadOptions{SpaceBeforeInlineComment: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	if f.Section("section").Key("key").String() != "value#not-a-comment" {
		t.Errorf("key: got %q", f.Section("section").Key("key").String())
	}
	if f.Section("section").Key("key2").String() != "value2" {
		t.Errorf("key2: got %q", f.Section("section").Key("key2").String())
	}
}

func TestQuotedValues(t *testing.T) {
	ini := `[section]
key = "hello world"
key2 = 'hello world'
`
	f, err := Load([]byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	if f.Section("section").Key("key").String() != "hello world" {
		t.Errorf("double-quoted: got %q", f.Section("section").Key("key").String())
	}
	if f.Section("section").Key("key2").String() != "hello world" {
		t.Errorf("single-quoted: got %q", f.Section("section").Key("key2").String())
	}
}

func TestPreserveSurroundedQuote(t *testing.T) {
	ini := `[section]
key = "hello world"
`
	f, err := LoadSources(LoadOptions{PreserveSurroundedQuote: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	if f.Section("section").Key("key").String() != `"hello world"` {
		t.Errorf("got %q", f.Section("section").Key("key").String())
	}
}

func TestTripleQuotedMultiline(t *testing.T) {
	ini := "[section]\nkey = \"\"\"line1\nline2\nline3\"\"\"\n"
	f, err := Load([]byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	expected := "line1\nline2\nline3"
	if f.Section("section").Key("key").String() != expected {
		t.Errorf("got %q, want %q", f.Section("section").Key("key").String(), expected)
	}
}

func TestBackslashContinuation(t *testing.T) {
	ini := `[section]
key = hello \
world
`
	f, err := Load([]byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	if f.Section("section").Key("key").String() != "hello world" {
		t.Errorf("got %q", f.Section("section").Key("key").String())
	}
}

func TestIgnoreContinuation(t *testing.T) {
	ini := `[section]
key = hello \
`
	f, err := LoadSources(LoadOptions{IgnoreContinuation: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	val := f.Section("section").Key("key").String()
	if !strings.HasSuffix(val, `\`) {
		t.Errorf("expected trailing backslash, got %q", val)
	}
}

func TestInsensitive(t *testing.T) {
	ini := `[Section]
Key = value
`
	f, err := LoadSources(LoadOptions{Insensitive: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	if !f.HasSection("section") {
		t.Error("case-insensitive section lookup failed")
	}
	if f.Section("section").Key("key").String() != "value" {
		t.Error("case-insensitive key lookup failed")
	}
}

func TestPythonMultilineValues(t *testing.T) {
	ini := `[section]
key = line1
  line2
  line3
other = normal
`
	f, err := LoadSources(LoadOptions{AllowPythonMultilineValues: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	expected := "line1\nline2\nline3"
	if f.Section("section").Key("key").String() != expected {
		t.Errorf("got %q, want %q", f.Section("section").Key("key").String(), expected)
	}
	if f.Section("section").Key("other").String() != "normal" {
		t.Errorf("other: got %q", f.Section("section").Key("other").String())
	}
}

func TestNestedValues(t *testing.T) {
	ini := `[section]
key = value
  nested1
  nested2
`
	f, err := LoadSources(LoadOptions{AllowNestedValues: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	k := f.Section("section").Key("key")
	nested := k.NestedValues()
	if len(nested) != 2 {
		t.Fatalf("expected 2 nested values, got %d", len(nested))
	}
	if nested[0] != "nested1" || nested[1] != "nested2" {
		t.Errorf("nested values: %v", nested)
	}
}

func TestShadowValues(t *testing.T) {
	ini := `[section]
key = a
key = b
key = c
`
	f, err := LoadSources(LoadOptions{AllowShadows: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	k := f.Section("section").Key("key")
	vals := k.ValueWithShadows()
	if len(vals) != 3 {
		t.Fatalf("expected 3 shadow values, got %d", len(vals))
	}
	if vals[0] != "a" || vals[1] != "b" || vals[2] != "c" {
		t.Errorf("shadow values: %v", vals)
	}
}

func TestWriteTo(t *testing.T) {
	f := Empty()
	sec, _ := f.NewSection("server")
	sec.NewKey("host", "localhost")
	sec.NewKey("port", "8080")

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "[server]") {
		t.Error("missing section header")
	}
	if !strings.Contains(output, "host=localhost") && !strings.Contains(output, "host = localhost") {
		t.Errorf("missing key-value: %s", output)
	}
}

func TestDeleteSection(t *testing.T) {
	f, err := Load([]byte(sampleINI))
	if err != nil {
		t.Fatal(err)
	}
	f.DeleteSection("database")
	if f.HasSection("database") {
		t.Error("section not deleted")
	}
}

func TestDeleteKey(t *testing.T) {
	f, err := Load([]byte(sampleINI))
	if err != nil {
		t.Fatal(err)
	}
	sec := f.Section("section")
	sec.DeleteKey("key")
	if sec.HasKey("key") {
		t.Error("key not deleted")
	}
}

func TestSectionComment(t *testing.T) {
	ini := `# Section comment
[section]
key = value
`
	f, err := Load([]byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	sec := f.Section("section")
	if sec.Comment != "# Section comment" {
		t.Errorf("comment: got %q", sec.Comment)
	}
}

func TestKeyComment(t *testing.T) {
	ini := `[section]
# Key comment
key = value
`
	f, err := Load([]byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	k := f.Section("section").Key("key")
	if k.Comment != "# Key comment" {
		t.Errorf("comment: got %q", k.Comment)
	}
}

func TestUTF8BOM(t *testing.T) {
	bom := []byte{0xEF, 0xBB, 0xBF}
	ini := append(bom, []byte("[section]\nkey = value\n")...)
	f, err := Load(ini)
	if err != nil {
		t.Fatal(err)
	}
	if f.Section("section").Key("key").String() != "value" {
		t.Error("BOM not stripped")
	}
}

func TestUnparseableSection(t *testing.T) {
	ini := `[raw]
this is raw content
not key = value pairs
just text

[normal]
key = value
`
	f, err := LoadSources(LoadOptions{UnparseableSections: []string{"raw"}}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	sec := f.Section("raw")
	if !strings.Contains(sec.Body(), "this is raw content") {
		t.Errorf("raw body: %q", sec.Body())
	}
	if f.Section("normal").Key("key").String() != "value" {
		t.Error("normal section broken after raw")
	}
}

func TestSkipUnrecognizableLines(t *testing.T) {
	ini := `[section]
key = value
this line has no delimiter and no boolean key
other = works
`
	f, err := LoadSources(LoadOptions{SkipUnrecognizableLines: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	if f.Section("section").Key("key").String() != "value" {
		t.Error("key mismatch")
	}
	if f.Section("section").Key("other").String() != "works" {
		t.Error("other mismatch")
	}
}

func TestUnescapeOptions(t *testing.T) {
	ini := `[section]
key = hello \"world\"
key2 = value \# not comment
`
	f, err := LoadSources(LoadOptions{
		IgnoreInlineComment:        true,
		UnescapeValueDoubleQuotes:  true,
		UnescapeValueCommentSymbols: true,
	}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	if f.Section("section").Key("key").String() != `hello "world"` {
		t.Errorf("unescape double quotes: got %q", f.Section("section").Key("key").String())
	}
	if f.Section("section").Key("key2").String() != "value # not comment" {
		t.Errorf("unescape comment: got %q", f.Section("section").Key("key2").String())
	}
}

func TestChildSections(t *testing.T) {
	ini := `[parent]
key = value
[parent.child1]
key = value1
[parent.child2]
key = value2
`
	f, err := Load([]byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	children := f.ChildSections("parent")
	if len(children) != 2 {
		t.Fatalf("expected 2 child sections, got %d", len(children))
	}
}

func TestParentKeys(t *testing.T) {
	ini := `[parent]
inherited = yes
[parent.child]
own = value
`
	f, err := Load([]byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	child := f.Section("parent.child")
	parentKeys := child.ParentKeys()
	if len(parentKeys) != 1 {
		t.Fatalf("expected 1 parent key, got %d", len(parentKeys))
	}
	if parentKeys[0].Name() != "inherited" {
		t.Errorf("parent key name: got %q", parentKeys[0].Name())
	}
}

func TestEmpty(t *testing.T) {
	f := Empty()
	sec, err := f.NewSection("test")
	if err != nil {
		t.Fatal(err)
	}
	sec.NewKey("key", "value")
	if f.Section("test").Key("key").String() != "value" {
		t.Error("empty file key mismatch")
	}
}

func TestKeysHash(t *testing.T) {
	f, err := Load([]byte(sampleINI))
	if err != nil {
		t.Fatal(err)
	}
	hash := f.Section("database").KeysHash()
	if hash["host"] != "localhost" {
		t.Error("hash mismatch")
	}
}

func TestHasValue(t *testing.T) {
	f, err := Load([]byte(sampleINI))
	if err != nil {
		t.Fatal(err)
	}
	if !f.Section("database").HasValue("localhost") {
		t.Error("HasValue should return true")
	}
	if f.Section("database").HasValue("nonexistent") {
		t.Error("HasValue should return false")
	}
}

func TestNonUniqueSections(t *testing.T) {
	ini := `[section]
key1 = a
[section]
key2 = b
`
	f, err := LoadSources(LoadOptions{AllowNonUniqueSections: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	sections, err := f.SectionsByName("section")
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}
}
