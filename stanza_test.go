package stanza

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// resetGlobals restores package-level write globals to their defaults after a
// test that modifies them.
func resetGlobals(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		PrettyEqual = false
		PrettySection = true
		PrettyFormat = true
		DefaultHeader = false
		DefaultFormatLeft = ""
		DefaultFormatRight = ""
	})
}

// writeToString is a convenience wrapper around WriteTo.
func writeToString(f *File) (string, error) {
	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ---------------------------------------------------------------------------
// WriteTo / serialization tests
// ---------------------------------------------------------------------------

func TestWriteToBasic(t *testing.T) {
	resetGlobals(t)
	PrettySection = false

	f := Empty()
	sec, _ := f.NewSection("app")
	sec.NewKey("name", "myapp")
	sec.NewKey("version", "1.0")

	out, err := writeToString(f)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "[app]") {
		t.Errorf("missing section header; got:\n%s", out)
	}
	if !strings.Contains(out, "name=myapp") {
		t.Errorf("missing key=value; got:\n%s", out)
	}
	if !strings.Contains(out, "version=1.0") {
		t.Errorf("missing version; got:\n%s", out)
	}
}

func TestWriteToPrettyEqual(t *testing.T) {
	resetGlobals(t)
	PrettyEqual = true
	PrettySection = false

	f := Empty()
	sec, _ := f.NewSection("cfg")
	sec.NewKey("key", "val")

	out, err := writeToString(f)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "key = val") {
		t.Errorf("PrettyEqual: expected 'key = val'; got:\n%s", out)
	}
	if strings.Contains(out, "key=val") {
		t.Errorf("PrettyEqual: unexpected 'key=val' without spaces; got:\n%s", out)
	}
}

func TestWriteToPrettyEqualFalse(t *testing.T) {
	resetGlobals(t)
	PrettyEqual = false
	PrettySection = false

	f := Empty()
	sec, _ := f.NewSection("cfg")
	sec.NewKey("key", "val")

	out, err := writeToString(f)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "key=val") {
		t.Errorf("expected 'key=val'; got:\n%s", out)
	}
}

func TestWriteToPrettySection(t *testing.T) {
	resetGlobals(t)
	PrettySection = true

	f := Empty()
	f.NewSection("alpha")
	f.NewSection("beta")

	out, err := writeToString(f)
	if err != nil {
		t.Fatal(err)
	}

	// There should be a blank line before [beta].
	lines := strings.Split(out, "\n")
	betaIdx := -1
	for i, l := range lines {
		if l == "[beta]" {
			betaIdx = i
			break
		}
	}
	if betaIdx < 1 {
		t.Fatalf("could not find [beta] in output:\n%s", out)
	}
	if strings.TrimSpace(lines[betaIdx-1]) != "" {
		t.Errorf("PrettySection: expected blank line before [beta]; lines[%d]=%q\n%s", betaIdx-1, lines[betaIdx-1], out)
	}
}

func TestWriteToPrettySectionFalse(t *testing.T) {
	resetGlobals(t)
	PrettySection = false

	f := Empty()
	f.NewSection("alpha")
	f.NewSection("beta")

	out, err := writeToString(f)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(out, "\n")
	alphaIdx := -1
	betaIdx := -1
	for i, l := range lines {
		if l == "[alpha]" {
			alphaIdx = i
		}
		if l == "[beta]" {
			betaIdx = i
		}
	}
	if alphaIdx < 0 || betaIdx < 0 {
		t.Fatalf("sections not found in output:\n%s", out)
	}
	// With PrettySection=false there should be no blank line between sections
	// (they are adjacent or only the keys of alpha separate them).
	if betaIdx-alphaIdx > 1 {
		t.Errorf("PrettySection=false: unexpected gap between sections; alpha=%d beta=%d\n%s", alphaIdx, betaIdx, out)
	}
}

func TestWriteToDefaultHeader(t *testing.T) {
	resetGlobals(t)
	DefaultHeader = true
	PrettySection = false

	f := Empty()
	f.Section(DefaultSection).NewKey("global", "yes")

	out, err := writeToString(f)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "[DEFAULT]") {
		t.Errorf("DefaultHeader: expected [DEFAULT] in output; got:\n%s", out)
	}
	if !strings.Contains(out, "global=yes") {
		t.Errorf("DefaultHeader: expected global key; got:\n%s", out)
	}
}

func TestWriteToDefaultHeaderFalse(t *testing.T) {
	resetGlobals(t)
	DefaultHeader = false
	PrettySection = false

	f := Empty()
	f.Section(DefaultSection).NewKey("global", "yes")
	f.NewSection("app")
	f.Section("app").NewKey("name", "x")

	out, err := writeToString(f)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "[DEFAULT]") {
		t.Errorf("DefaultHeader=false: unexpected [DEFAULT] in output; got:\n%s", out)
	}
}

func TestWriteToIndent(t *testing.T) {
	resetGlobals(t)
	PrettySection = false

	f := Empty()
	sec, _ := f.NewSection("section")
	sec.NewKey("key", "value")

	var buf bytes.Buffer
	if _, err := f.WriteToIndent(&buf, "\t"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "\tkey=value") {
		t.Errorf("WriteToIndent: expected tab-indented key; got:\n%s", out)
	}
}

func TestSaveTo(t *testing.T) {
	resetGlobals(t)
	PrettySection = false

	f := Empty()
	sec, _ := f.NewSection("saved")
	sec.NewKey("data", "hello")

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "out.ini")

	if err := f.SaveTo(path); err != nil {
		t.Fatal(err)
	}

	rawBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(rawBytes)
	if !strings.Contains(content, "[saved]") {
		t.Errorf("SaveTo: missing section; got:\n%s", content)
	}
	if !strings.Contains(content, "data=hello") {
		t.Errorf("SaveTo: missing key; got:\n%s", content)
	}

	// Read back via Load and verify.
	f2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if f2.Section("saved").Key("data").String() != "hello" {
		t.Errorf("SaveTo round-trip: got %q", f2.Section("saved").Key("data").String())
	}
}

func TestWriteToCommentPreservation(t *testing.T) {
	resetGlobals(t)
	PrettySection = false

	ini := `# Section comment
[section]
# Key comment
key = value
`
	f, err := Load([]byte(ini))
	if err != nil {
		t.Fatal(err)
	}

	out, err := writeToString(f)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "# Section comment") {
		t.Errorf("comment preservation: missing section comment; got:\n%s", out)
	}
	if !strings.Contains(out, "# Key comment") {
		t.Errorf("comment preservation: missing key comment; got:\n%s", out)
	}
}

func TestWriteToKeyValueDelimiterOnWrite(t *testing.T) {
	resetGlobals(t)
	PrettySection = false

	f, err := LoadSources(LoadOptions{
		KeyValueDelimiterOnWrite: ":",
	}, []byte("[section]\nkey=value\n"))
	if err != nil {
		t.Fatal(err)
	}

	out, err := writeToString(f)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "key:value") {
		t.Errorf("KeyValueDelimiterOnWrite: expected 'key:value'; got:\n%s", out)
	}
}

func TestWriteToBooleanKey(t *testing.T) {
	resetGlobals(t)
	PrettySection = false

	f, err := LoadSources(LoadOptions{AllowBooleanKeys: true}, []byte("[flags]\nenabled\ndebug\n"))
	if err != nil {
		t.Fatal(err)
	}

	out, err := writeToString(f)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "enabled") {
		t.Errorf("boolean key: missing 'enabled'; got:\n%s", out)
	}
	// Boolean keys must not have a "=" appended.
	if strings.Contains(out, "enabled=") {
		t.Errorf("boolean key: 'enabled=' should not appear; got:\n%s", out)
	}
	if strings.Contains(out, "debug=") {
		t.Errorf("boolean key: 'debug=' should not appear; got:\n%s", out)
	}
}

func TestWriteToRawSection(t *testing.T) {
	resetGlobals(t)
	PrettySection = false

	f := Empty()
	f.NewRawSection("raw", "this is raw\ncontent here")

	out, err := writeToString(f)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "[raw]") {
		t.Errorf("raw section: missing header; got:\n%s", out)
	}
	if !strings.Contains(out, "this is raw") {
		t.Errorf("raw section: missing body; got:\n%s", out)
	}
}

func TestWriteToRawSectionFromLoad(t *testing.T) {
	resetGlobals(t)
	PrettySection = false

	ini := `[raw]
unparsed body line
another line
[normal]
key = value
`
	f, err := LoadSources(LoadOptions{UnparseableSections: []string{"raw"}}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}

	out, err := writeToString(f)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "[raw]") {
		t.Errorf("raw section output: missing [raw]; got:\n%s", out)
	}
	if !strings.Contains(out, "unparsed body line") {
		t.Errorf("raw section output: missing body; got:\n%s", out)
	}
	if !strings.Contains(out, "[normal]") {
		t.Errorf("raw section output: missing [normal]; got:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// LoadOptions tests (not already covered in parser_test.go)
// ---------------------------------------------------------------------------

func TestLooseLoad(t *testing.T) {
	// A non-existent file should not return an error when Loose=true.
	_, err := LoadSources(LoadOptions{Loose: true}, "/nonexistent/path/that/does/not/exist.ini")
	if err != nil {
		t.Errorf("Loose: expected no error for missing file, got: %v", err)
	}
}

func TestLooseLoadFalse(t *testing.T) {
	// Without Loose, a missing file must return an error.
	_, err := Load("/nonexistent/path/that/does/not/exist.ini")
	if err == nil {
		t.Error("expected error for missing file without Loose")
	}
}

func TestLoadOptionsIgnoreInlineComment(t *testing.T) {
	ini := "[s]\nkey = val#notcomment\n"
	f, err := LoadSources(LoadOptions{IgnoreInlineComment: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	if f.Section("s").Key("key").String() != "val#notcomment" {
		t.Errorf("IgnoreInlineComment: got %q", f.Section("s").Key("key").String())
	}
}

func TestLoadOptionsSpaceBeforeInlineComment(t *testing.T) {
	ini := "[s]\nwithspace = val ;comment\nnospace = val;notcomment\n"
	f, err := LoadSources(LoadOptions{SpaceBeforeInlineComment: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	if got := f.Section("s").Key("withspace").String(); got != "val" {
		t.Errorf("SpaceBeforeInlineComment withspace: got %q, want %q", got, "val")
	}
	if got := f.Section("s").Key("nospace").String(); got != "val;notcomment" {
		t.Errorf("SpaceBeforeInlineComment nospace: got %q, want %q", got, "val;notcomment")
	}
}

func TestLoadOptionsAllowBooleanKeys(t *testing.T) {
	ini := "[flags]\nfeature_x\nfeature_y\n"
	f, err := LoadSources(LoadOptions{AllowBooleanKeys: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	k, err := f.Section("flags").GetKey("feature_x")
	if err != nil {
		t.Fatal(err)
	}
	if !k.IsBooleanKey() {
		t.Error("AllowBooleanKeys: expected boolean key")
	}
	if f.Section("flags").Key("feature_y").IsBooleanKey() != true {
		t.Error("AllowBooleanKeys: feature_y should be boolean")
	}
}

func TestLoadOptionsAllowShadows(t *testing.T) {
	ini := "[s]\nport = 80\nport = 443\nport = 8080\n"
	f, err := LoadSources(LoadOptions{AllowShadows: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	vals := f.Section("s").Key("port").ValueWithShadows()
	if len(vals) != 3 {
		t.Fatalf("AllowShadows: expected 3 values, got %d: %v", len(vals), vals)
	}
	if vals[0] != "80" || vals[1] != "443" || vals[2] != "8080" {
		t.Errorf("AllowShadows: wrong values: %v", vals)
	}
}

func TestLoadOptionsAllowNestedValues(t *testing.T) {
	ini := "[s]\nkey = top\n  sub1\n  sub2\nother = plain\n"
	f, err := LoadSources(LoadOptions{AllowNestedValues: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	k := f.Section("s").Key("key")
	if k.String() != "top" {
		t.Errorf("AllowNestedValues: primary value: got %q", k.String())
	}
	nested := k.NestedValues()
	if len(nested) != 2 {
		t.Fatalf("AllowNestedValues: expected 2 nested, got %d", len(nested))
	}
	if nested[0] != "sub1" || nested[1] != "sub2" {
		t.Errorf("AllowNestedValues: got %v", nested)
	}
	if f.Section("s").Key("other").String() != "plain" {
		t.Errorf("AllowNestedValues: other key broken")
	}
}

func TestLoadOptionsAllowPythonMultilineValues(t *testing.T) {
	ini := "[s]\naddress = 123 Main St\n  Suite 100\n  Anytown, USA\nnext = done\n"
	f, err := LoadSources(LoadOptions{AllowPythonMultilineValues: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	got := f.Section("s").Key("address").String()
	want := "123 Main St\nSuite 100\nAnytown, USA"
	if got != want {
		t.Errorf("AllowPythonMultilineValues: got %q, want %q", got, want)
	}
	if f.Section("s").Key("next").String() != "done" {
		t.Errorf("AllowPythonMultilineValues: subsequent key broken")
	}
}

func TestLoadOptionsIgnoreContinuation(t *testing.T) {
	// With IgnoreContinuation the trailing backslash is kept as-is.
	// We also enable SkipUnrecognizableLines so the continuation line
	// ("world") that would normally be the joined value is simply dropped.
	ini := "[s]\nkey = hello \\\nworld = kept\n"
	f, err := LoadSources(LoadOptions{
		IgnoreContinuation: true,
	}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	val := f.Section("s").Key("key").String()
	if !strings.HasSuffix(val, `\`) {
		t.Errorf("IgnoreContinuation: expected trailing backslash, got %q", val)
	}
	// The next line is now a separate key=value pair, not joined.
	if f.Section("s").Key("world").String() != "kept" {
		t.Errorf("IgnoreContinuation: next line should be its own key; got %q", f.Section("s").Key("world").String())
	}
}

func TestLoadOptionsSkipUnrecognizableLines(t *testing.T) {
	// Without AllowBooleanKeys, lines without delimiters are normally errors.
	// With SkipUnrecognizableLines they are silently ignored.
	ini := "[s]\ngood = ok\nthis line is garbage with no delimiter\nbad line two\nother = also ok\n"
	f, err := LoadSources(LoadOptions{SkipUnrecognizableLines: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	if f.Section("s").Key("good").String() != "ok" {
		t.Errorf("SkipUnrecognizableLines: good key: got %q", f.Section("s").Key("good").String())
	}
	if f.Section("s").Key("other").String() != "also ok" {
		t.Errorf("SkipUnrecognizableLines: other key: got %q", f.Section("s").Key("other").String())
	}
}

func TestLoadOptionsUnescapeValueDoubleQuotes(t *testing.T) {
	ini := "[s]\nkey = say \\\"hello\\\"\n"
	f, err := LoadSources(LoadOptions{
		IgnoreInlineComment:       true,
		UnescapeValueDoubleQuotes: true,
	}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	got := f.Section("s").Key("key").String()
	want := `say "hello"`
	if got != want {
		t.Errorf("UnescapeValueDoubleQuotes: got %q, want %q", got, want)
	}
}

func TestLoadOptionsUnescapeValueCommentSymbols(t *testing.T) {
	ini := "[s]\nhash = value \\# with hash\nsemicolon = value \\; with semi\n"
	f, err := LoadSources(LoadOptions{
		IgnoreInlineComment:         true,
		UnescapeValueCommentSymbols: true,
	}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	if got := f.Section("s").Key("hash").String(); got != "value # with hash" {
		t.Errorf("UnescapeValueCommentSymbols hash: got %q", got)
	}
	if got := f.Section("s").Key("semicolon").String(); got != "value ; with semi" {
		t.Errorf("UnescapeValueCommentSymbols semi: got %q", got)
	}
}

func TestLoadOptionsAllowNonUniqueSections(t *testing.T) {
	ini := "[db]\nhost = primary\n[db]\nhost = replica\n"
	f, err := LoadSources(LoadOptions{AllowNonUniqueSections: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	sections, err := f.SectionsByName("db")
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 2 {
		t.Fatalf("AllowNonUniqueSections: expected 2 sections, got %d", len(sections))
	}
	if sections[0].Key("host").String() != "primary" {
		t.Errorf("first section host: got %q", sections[0].Key("host").String())
	}
	if sections[1].Key("host").String() != "replica" {
		t.Errorf("second section host: got %q", sections[1].Key("host").String())
	}
}

func TestLoadOptionsInsensitive(t *testing.T) {
	ini := "[MySection]\nMyKey = MyValue\n"
	f, err := LoadSources(LoadOptions{Insensitive: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	// Both section and key should be accessible in lowercase.
	if !f.HasSection("mysection") {
		t.Error("Insensitive: section not found via lowercase")
	}
	if f.Section("mysection").Key("mykey").String() != "MyValue" {
		t.Errorf("Insensitive: key lookup failed: got %q", f.Section("mysection").Key("mykey").String())
	}
}

func TestLoadOptionsInsensitiveSections(t *testing.T) {
	ini := "[MySection]\nMyKey = value\n"
	f, err := LoadSources(LoadOptions{InsensitiveSections: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	if !f.HasSection("mysection") {
		t.Error("InsensitiveSections: section not found via lowercase")
	}
	// Key name should still be case-sensitive.
	if !f.Section("mysection").HasKey("MyKey") {
		t.Error("InsensitiveSections: key should still be case-sensitive")
	}
}

func TestLoadOptionsInsensitiveKeys(t *testing.T) {
	ini := "[Section]\nMyKey = value\n"
	f, err := LoadSources(LoadOptions{InsensitiveKeys: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	// Section name should still be case-sensitive.
	if !f.HasSection("Section") {
		t.Error("InsensitiveKeys: section should be case-sensitive")
	}
	// Key name should be lowercase.
	if f.Section("Section").Key("mykey").String() != "value" {
		t.Errorf("InsensitiveKeys: key lookup failed: got %q", f.Section("Section").Key("mykey").String())
	}
}

func TestLoadOptionsKeyValueDelimitersColon(t *testing.T) {
	// Use only ":" as delimiter; "=" should NOT split the key name.
	// Lines without ":" would be errors, so only include colon-delimited lines.
	ini := "[s]\nkey: colon value\nother: also colon\n"
	f, err := LoadSources(LoadOptions{KeyValueDelimiters: ":"}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	if f.Section("s").Key("key").String() != "colon value" {
		t.Errorf("colon delimiter key: got %q", f.Section("s").Key("key").String())
	}
	if f.Section("s").Key("other").String() != "also colon" {
		t.Errorf("colon delimiter other: got %q", f.Section("s").Key("other").String())
	}
}

func TestLoadOptionsPreserveSurroundedQuote(t *testing.T) {
	ini := "[s]\nkey = \"quoted value\"\nkey2 = 'single quoted'\n"
	f, err := LoadSources(LoadOptions{PreserveSurroundedQuote: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	if got := f.Section("s").Key("key").String(); got != `"quoted value"` {
		t.Errorf("PreserveSurroundedQuote double: got %q", got)
	}
	if got := f.Section("s").Key("key2").String(); got != `'single quoted'` {
		t.Errorf("PreserveSurroundedQuote single: got %q", got)
	}
}

// ---------------------------------------------------------------------------
// Roundtrip tests
// ---------------------------------------------------------------------------

func TestRoundtripBasic(t *testing.T) {
	resetGlobals(t)
	PrettySection = false
	PrettyEqual = false

	original := "[server]\nhost=localhost\nport=8080\n[database]\nname=mydb\n"
	f1, err := Load([]byte(original))
	if err != nil {
		t.Fatal(err)
	}

	out, err := writeToString(f1)
	if err != nil {
		t.Fatal(err)
	}

	f2, err := Load([]byte(out))
	if err != nil {
		t.Fatalf("roundtrip re-parse failed: %v\noutput was:\n%s", err, out)
	}

	if f2.Section("server").Key("host").String() != "localhost" {
		t.Errorf("roundtrip: server.host: got %q", f2.Section("server").Key("host").String())
	}
	if f2.Section("server").Key("port").String() != "8080" {
		t.Errorf("roundtrip: server.port: got %q", f2.Section("server").Key("port").String())
	}
	if f2.Section("database").Key("name").String() != "mydb" {
		t.Errorf("roundtrip: database.name: got %q", f2.Section("database").Key("name").String())
	}
}

func TestRoundtripPrettyEqual(t *testing.T) {
	resetGlobals(t)
	PrettySection = false
	PrettyEqual = true

	original := "[cfg]\nalpha=one\nbeta=two\n"
	f1, err := Load([]byte(original))
	if err != nil {
		t.Fatal(err)
	}

	out, err := writeToString(f1)
	if err != nil {
		t.Fatal(err)
	}

	// After write, the file uses " = " spacing.
	f2, err := Load([]byte(out))
	if err != nil {
		t.Fatalf("PrettyEqual roundtrip re-parse failed: %v\noutput:\n%s", err, out)
	}
	if f2.Section("cfg").Key("alpha").String() != "one" {
		t.Errorf("PrettyEqual roundtrip alpha: got %q", f2.Section("cfg").Key("alpha").String())
	}
	if f2.Section("cfg").Key("beta").String() != "two" {
		t.Errorf("PrettyEqual roundtrip beta: got %q", f2.Section("cfg").Key("beta").String())
	}
}

func TestRoundtripWithComments(t *testing.T) {
	resetGlobals(t)
	PrettySection = false

	ini := "# top comment\n[section]\n# key comment\nvalue=42\n"
	f1, err := Load([]byte(ini))
	if err != nil {
		t.Fatal(err)
	}

	out, err := writeToString(f1)
	if err != nil {
		t.Fatal(err)
	}

	f2, err := Load([]byte(out))
	if err != nil {
		t.Fatalf("comments roundtrip re-parse failed: %v\noutput:\n%s", err, out)
	}
	if f2.Section("section").Key("value").String() != "42" {
		t.Errorf("comments roundtrip: value: got %q", f2.Section("section").Key("value").String())
	}
}

func TestRoundtripSaveTo(t *testing.T) {
	resetGlobals(t)
	PrettySection = false

	ini := "[net]\naddr=127.0.0.1\nport=9090\n"
	f1, err := Load([]byte(ini))
	if err != nil {
		t.Fatal(err)
	}

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "roundtrip.ini")
	if err := f1.SaveTo(path); err != nil {
		t.Fatal(err)
	}

	f2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if f2.Section("net").Key("addr").String() != "127.0.0.1" {
		t.Errorf("SaveTo roundtrip addr: got %q", f2.Section("net").Key("addr").String())
	}
	if f2.Section("net").Key("port").String() != "9090" {
		t.Errorf("SaveTo roundtrip port: got %q", f2.Section("net").Key("port").String())
	}
}

func TestRoundtripMultipleSections(t *testing.T) {
	resetGlobals(t)
	PrettySection = false
	PrettyEqual = false

	ini := "[a]\nx=1\n[b]\ny=2\n[c]\nz=3\n"
	f1, err := Load([]byte(ini))
	if err != nil {
		t.Fatal(err)
	}

	out, err := writeToString(f1)
	if err != nil {
		t.Fatal(err)
	}

	f2, err := Load([]byte(out))
	if err != nil {
		t.Fatalf("multi-section roundtrip: %v\noutput:\n%s", err, out)
	}
	for _, tc := range []struct{ section, key, want string }{
		{"a", "x", "1"},
		{"b", "y", "2"},
		{"c", "z", "3"},
	} {
		got := f2.Section(tc.section).Key(tc.key).String()
		if got != tc.want {
			t.Errorf("roundtrip [%s].%s: got %q, want %q", tc.section, tc.key, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Bug fix tests
// ---------------------------------------------------------------------------

func TestDeleteSection_AllowNonUniqueSections(t *testing.T) {
	ini := `[sec]
key1 = a

[sec]
key2 = b

[sec]
key3 = c
`
	f, err := LoadSources(LoadOptions{AllowNonUniqueSections: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}

	// Should have 3 sections with name "sec".
	secs, err2 := f.SectionsByName("sec")
	if err2 != nil {
		t.Fatal(err2)
	}
	if len(secs) != 3 {
		t.Fatalf("expected 3 sections named 'sec', got %d", len(secs))
	}

	// Delete the first one.
	f.DeleteSection("sec")

	// Should have 2 remaining.
	secs, err2 = f.SectionsByName("sec")
	if err2 != nil {
		t.Fatal(err2)
	}
	if len(secs) != 2 {
		t.Fatalf("expected 2 sections named 'sec' after delete, got %d", len(secs))
	}

	// The section should still be accessible.
	if !f.HasSection("sec") {
		t.Error("HasSection('sec') should still return true")
	}

	// First remaining section should have key2.
	sec := f.Section("sec")
	if sec.Key("key2").String() != "b" {
		t.Errorf("expected key2=b in first remaining section, got %q", sec.Key("key2").String())
	}
}

func TestStripInlineComment_EscapedThenUnescaped(t *testing.T) {
	ini := `[section]
key = value \# not a comment # real comment
`
	f, err := Load([]byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	got := f.Section("section").Key("key").String()
	// The escaped \# should be preserved, but text after the unescaped # should be stripped.
	if !strings.Contains(got, `\#`) {
		t.Errorf("expected escaped \\# to be preserved, got %q", got)
	}
	if strings.Contains(got, "real comment") {
		t.Errorf("expected unescaped # comment to be stripped, got %q", got)
	}
}

func TestSectionMapTo(t *testing.T) {
	ini := `[app]
name = TestApp
port = 8080
`
	f, err := Load([]byte(ini))
	if err != nil {
		t.Fatal(err)
	}

	type Config struct {
		Name string `ini:"name"`
		Port int    `ini:"port"`
	}
	var cfg Config
	err = f.Section("app").MapTo(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "TestApp" {
		t.Errorf("expected Name=TestApp, got %q", cfg.Name)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected Port=8080, got %d", cfg.Port)
	}
}

func TestSectionStrictMapTo(t *testing.T) {
	ini := `[app]
name = TestApp
unknown_key = oops
`
	f, err := Load([]byte(ini))
	if err != nil {
		t.Fatal(err)
	}

	type Config struct {
		Name string `ini:"name"`
	}
	var cfg Config
	err = f.Section("app").StrictMapTo(&cfg)
	if err == nil {
		t.Error("expected error for unmapped key in StrictMapTo")
	}
}

func TestSectionReflectFrom(t *testing.T) {
	f := Empty()
	sec, err := f.NewSection("app")
	if err != nil {
		t.Fatal(err)
	}

	type Config struct {
		Name string `ini:"name"`
		Port int    `ini:"port"`
	}
	cfg := Config{Name: "MyApp", Port: 9090}
	err = sec.ReflectFrom(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	if sec.Key("name").String() != "MyApp" {
		t.Errorf("expected name=MyApp, got %q", sec.Key("name").String())
	}
	if sec.Key("port").String() != "9090" {
		t.Errorf("expected port=9090, got %q", sec.Key("port").String())
	}
}

func TestReflectFromPackageLevel(t *testing.T) {
	type Config struct {
		Name string `ini:"name"`
	}
	cfg := Config{Name: "test"}
	// ReflectFrom requires a loadable source; use empty bytes.
	err := ReflectFrom(&cfg, []byte(""))
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadCloserClosed(t *testing.T) {
	// Verify that io.ReadCloser sources get closed.
	data := []byte("[section]\nkey = val\n")
	rc := &trackingReadCloser{Reader: bytes.NewReader(data)}
	f, err := Load(rc)
	if err != nil {
		t.Fatal(err)
	}
	if !rc.closed {
		t.Error("ReadCloser should have been closed after Load")
	}
	if f.Section("section").Key("key").String() != "val" {
		t.Error("expected key=val")
	}
}

func TestStripInlineComment_SpaceBeforeInlineComment(t *testing.T) {
	ini := `[section]
key = val#tag # real comment
`
	f, err := LoadSources(LoadOptions{SpaceBeforeInlineComment: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}
	got := f.Section("section").Key("key").String()
	// "val#tag" has no space before #, so it should be preserved.
	// " # real comment" has a space before #, so it should be stripped.
	if got != "val#tag" {
		t.Errorf("expected 'val#tag', got %q", got)
	}
}

func TestDeleteSectionWithIndex_UpdatesSectionsByName(t *testing.T) {
	ini := `[sec]
key1 = a

[sec]
key2 = b

[sec]
key3 = c
`
	f, err := LoadSources(LoadOptions{AllowNonUniqueSections: true}, []byte(ini))
	if err != nil {
		t.Fatal(err)
	}

	// Delete the first section (index 0)
	err = f.DeleteSectionWithIndex("sec", 0)
	if err != nil {
		t.Fatal(err)
	}

	// f.Section("sec") should now return the formerly-second section (key2=b)
	sec := f.Section("sec")
	if sec.Key("key2").String() != "b" {
		t.Errorf("expected key2=b after deleting index 0, got %q", sec.Key("key2").String())
	}

	// Verify HasSection still works
	if !f.HasSection("sec") {
		t.Error("HasSection('sec') should still return true")
	}
}

type trackingReadCloser struct {
	*bytes.Reader
	closed bool
}

func (t *trackingReadCloser) Close() error {
	t.closed = true
	return nil
}
