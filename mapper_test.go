package stanza

import (
	"bytes"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Name mapper tests
// ---------------------------------------------------------------------------

func TestSnackCase(t *testing.T) {
	tests := []struct{ in, want string }{
		{"CamelCase", "camel_case"},
		{"HTTPServer", "http_server"},
		{"ID", "id"},
		{"MyXMLParser", "my_xml_parser"},
		{"Simple", "simple"},
		{"already_snake", "already_snake"},
		{"", ""},
		{"A", "a"},
		{"AB", "ab"},
		{"ABc", "a_bc"},
	}
	for _, tt := range tests {
		got := SnackCase(tt.in)
		if got != tt.want {
			t.Errorf("SnackCase(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestTitleUnderscore(t *testing.T) {
	tests := []struct{ in, want string }{
		{"CamelCase", "Camel_Case"},
		{"HTTPServer", "HTTP_Server"},
		{"Simple", "Simple"},
		{"", ""},
	}
	for _, tt := range tests {
		got := TitleUnderscore(tt.in)
		if got != tt.want {
			t.Errorf("TitleUnderscore(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// MapTo tests
// ---------------------------------------------------------------------------

type basicConfig struct {
	Name    string `ini:"name"`
	Port    int    `ini:"port"`
	Debug   bool   `ini:"debug"`
	Rate    float64
	Timeout time.Duration `ini:"timeout"`
}

func TestMapToBasic(t *testing.T) {
	data := []byte(`
name = myapp
port = 8080
debug = true
Rate = 3.14
timeout = 5s
`)
	f, err := Load(data)
	if err != nil {
		t.Fatal(err)
	}

	var cfg basicConfig
	if err := f.MapTo(&cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.Name != "myapp" {
		t.Errorf("Name = %q", cfg.Name)
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d", cfg.Port)
	}
	if !cfg.Debug {
		t.Error("Debug should be true")
	}
	if cfg.Rate != 3.14 {
		t.Errorf("Rate = %f", cfg.Rate)
	}
	if cfg.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v", cfg.Timeout)
	}
}

type nestedConfig struct {
	Name   string `ini:"name"`
	Server struct {
		Host string `ini:"host"`
		Port int    `ini:"port"`
	} `ini:"server"`
}

func TestMapToNested(t *testing.T) {
	data := []byte(`
name = myapp

[server]
host = localhost
port = 9090
`)
	f, err := Load(data)
	if err != nil {
		t.Fatal(err)
	}

	var cfg nestedConfig
	if err := f.MapTo(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "myapp" {
		t.Errorf("Name = %q", cfg.Name)
	}
	if cfg.Server.Host != "localhost" {
		t.Errorf("Host = %q", cfg.Server.Host)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Port = %d", cfg.Server.Port)
	}
}

type sliceConfig struct {
	Tags  []string `ini:"tags" ini-delim:","`
	Ports []int    `ini:"ports" ini-delim:","`
}

func TestMapToSlice(t *testing.T) {
	data := []byte(`
tags = web, api, prod
ports = 80, 443, 8080
`)
	f, err := Load(data)
	if err != nil {
		t.Fatal(err)
	}

	var cfg sliceConfig
	if err := f.MapTo(&cfg); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Tags) != 3 || cfg.Tags[0] != "web" || cfg.Tags[1] != "api" || cfg.Tags[2] != "prod" {
		t.Errorf("Tags = %v", cfg.Tags)
	}
	if len(cfg.Ports) != 3 || cfg.Ports[0] != 80 || cfg.Ports[2] != 8080 {
		t.Errorf("Ports = %v", cfg.Ports)
	}
}

func TestMapToSkipTag(t *testing.T) {
	type cfg struct {
		Name   string `ini:"name"`
		Secret string `ini:"-"`
	}
	data := []byte(`
name = app
Secret = hidden
`)
	f, err := Load(data)
	if err != nil {
		t.Fatal(err)
	}
	var c cfg
	if err := f.MapTo(&c); err != nil {
		t.Fatal(err)
	}
	if c.Name != "app" {
		t.Errorf("Name = %q", c.Name)
	}
	if c.Secret != "" {
		t.Errorf("Secret should be empty, got %q", c.Secret)
	}
}

func TestMapToWithNameMapper(t *testing.T) {
	type cfg struct {
		ServerName string
		ServerPort int
	}
	data := []byte(`
server_name = myapp
server_port = 3000
`)
	f, err := Load(data)
	if err != nil {
		t.Fatal(err)
	}
	f.NameMapper = SnackCase
	var c cfg
	if err := f.MapTo(&c); err != nil {
		t.Fatal(err)
	}
	if c.ServerName != "myapp" {
		t.Errorf("ServerName = %q", c.ServerName)
	}
	if c.ServerPort != 3000 {
		t.Errorf("ServerPort = %d", c.ServerPort)
	}
}

func TestMapToTime(t *testing.T) {
	type cfg struct {
		Created time.Time `ini:"created"`
	}
	data := []byte(`created = 2026-06-15T10:30:00Z`)
	f, err := Load(data)
	if err != nil {
		t.Fatal(err)
	}
	var c cfg
	if err := f.MapTo(&c); err != nil {
		t.Fatal(err)
	}
	if c.Created.Year() != 2026 || c.Created.Month() != 6 {
		t.Errorf("Created = %v", c.Created)
	}
}

func TestMapToUint(t *testing.T) {
	type cfg struct {
		MaxSize uint64 `ini:"max_size"`
	}
	data := []byte(`max_size = 1024`)
	f, err := Load(data)
	if err != nil {
		t.Fatal(err)
	}
	var c cfg
	if err := f.MapTo(&c); err != nil {
		t.Fatal(err)
	}
	if c.MaxSize != 1024 {
		t.Errorf("MaxSize = %d", c.MaxSize)
	}
}

func TestMapToPointerStruct(t *testing.T) {
	type DB struct {
		Host string `ini:"host"`
	}
	type cfg struct {
		Database *DB `ini:"database"`
	}

	data := []byte(`
[database]
host = db.local
`)
	f, err := Load(data)
	if err != nil {
		t.Fatal(err)
	}
	var c cfg
	if err := f.MapTo(&c); err != nil {
		t.Fatal(err)
	}
	if c.Database == nil {
		t.Fatal("Database should not be nil")
	}
	if c.Database.Host != "db.local" {
		t.Errorf("Host = %q", c.Database.Host)
	}
}

func TestMapToPointerStructMissing(t *testing.T) {
	type DB struct {
		Host string `ini:"host"`
	}
	type cfg struct {
		Database *DB `ini:"database"`
	}
	data := []byte(`name = app`)
	f, err := Load(data)
	if err != nil {
		t.Fatal(err)
	}
	var c cfg
	if err := f.MapTo(&c); err != nil {
		t.Fatal(err)
	}
	if c.Database != nil {
		t.Error("Database should be nil when section missing")
	}
}

func TestMapToNonPointer(t *testing.T) {
	var c basicConfig
	f := Empty()
	if err := f.MapTo(c); err == nil {
		t.Error("expected error for non-pointer")
	}
}

func TestStrictMapTo(t *testing.T) {
	type cfg struct {
		Name string `ini:"name"`
	}
	data := []byte(`
name = app
extra = value
`)
	f, err := Load(data)
	if err != nil {
		t.Fatal(err)
	}
	var c cfg
	if err := f.StrictMapTo(&c); err == nil {
		t.Error("expected error for unmapped key")
	}
}

// ---------------------------------------------------------------------------
// ReflectFrom tests
// ---------------------------------------------------------------------------

func TestReflectFromBasic(t *testing.T) {
	cfg := basicConfig{
		Name:    "myapp",
		Port:    8080,
		Debug:   true,
		Rate:    3.14,
		Timeout: 5 * time.Second,
	}
	f := Empty()
	if err := f.ReflectFrom(&cfg); err != nil {
		t.Fatal(err)
	}

	if f.Section(DefaultSection).Key("name").String() != "myapp" {
		t.Errorf("name = %q", f.Section(DefaultSection).Key("name").String())
	}
	if f.Section(DefaultSection).Key("port").MustInt() != 8080 {
		t.Error("port mismatch")
	}
	if !f.Section(DefaultSection).Key("debug").MustBool() {
		t.Error("debug should be true")
	}
	if f.Section(DefaultSection).Key("timeout").String() != "5s" {
		t.Errorf("timeout = %q", f.Section(DefaultSection).Key("timeout").String())
	}
}

func TestReflectFromNested(t *testing.T) {
	cfg := nestedConfig{Name: "myapp"}
	cfg.Server.Host = "localhost"
	cfg.Server.Port = 9090

	f := Empty()
	if err := f.ReflectFrom(&cfg); err != nil {
		t.Fatal(err)
	}

	if !f.HasSection("server") {
		t.Fatal("missing server section")
	}
	if f.Section("server").Key("host").String() != "localhost" {
		t.Errorf("host = %q", f.Section("server").Key("host").String())
	}
	if f.Section("server").Key("port").MustInt() != 9090 {
		t.Error("port mismatch")
	}
}

func TestReflectFromSlice(t *testing.T) {
	cfg := sliceConfig{
		Tags:  []string{"web", "api"},
		Ports: []int{80, 443},
	}
	f := Empty()
	if err := f.ReflectFrom(&cfg); err != nil {
		t.Fatal(err)
	}

	if f.Section(DefaultSection).Key("tags").String() != "web,api" {
		t.Errorf("tags = %q", f.Section(DefaultSection).Key("tags").String())
	}
	if f.Section(DefaultSection).Key("ports").String() != "80,443" {
		t.Errorf("ports = %q", f.Section(DefaultSection).Key("ports").String())
	}
}

func TestReflectFromOmitempty(t *testing.T) {
	type cfg struct {
		Name  string `ini:"name"`
		Empty string `ini:"empty,omitempty"`
	}
	c := cfg{Name: "app", Empty: ""}
	f := Empty()
	if err := f.ReflectFrom(&c); err != nil {
		t.Fatal(err)
	}
	if f.Section(DefaultSection).HasKey("empty") {
		t.Error("omitempty field should not be written")
	}
	if f.Section(DefaultSection).Key("name").String() != "app" {
		t.Error("name should be written")
	}
}

func TestReflectFromComment(t *testing.T) {
	type cfg struct {
		Port int `ini:"port" ini-comment:"# server port"`
	}
	c := cfg{Port: 8080}
	f := Empty()
	if err := f.ReflectFrom(&c); err != nil {
		t.Fatal(err)
	}
	k, err := f.Section(DefaultSection).GetKey("port")
	if err != nil {
		t.Fatal(err)
	}
	if k.Comment != "# server port" {
		t.Errorf("Comment = %q", k.Comment)
	}
}

func TestReflectFromSkipTag(t *testing.T) {
	type cfg struct {
		Name   string `ini:"name"`
		Secret string `ini:"-"`
	}
	c := cfg{Name: "app", Secret: "hidden"}
	f := Empty()
	if err := f.ReflectFrom(&c); err != nil {
		t.Fatal(err)
	}
	if f.Section(DefaultSection).HasKey("Secret") {
		t.Error("skipped field should not appear")
	}
}

func TestReflectFromTime(t *testing.T) {
	type cfg struct {
		Created time.Time `ini:"created"`
	}
	c := cfg{Created: time.Date(2026, 6, 15, 10, 30, 0, 0, time.UTC)}
	f := Empty()
	if err := f.ReflectFrom(&c); err != nil {
		t.Fatal(err)
	}
	if f.Section(DefaultSection).Key("created").String() != "2026-06-15T10:30:00Z" {
		t.Errorf("created = %q", f.Section(DefaultSection).Key("created").String())
	}
}

func TestReflectFromWithMapper(t *testing.T) {
	type cfg struct {
		ServerName string
	}
	c := cfg{ServerName: "myapp"}
	f := Empty()
	if err := f.ReflectFromWithMapper(&c, SnackCase); err != nil {
		t.Fatal(err)
	}
	if f.Section(DefaultSection).Key("server_name").String() != "myapp" {
		t.Errorf("server_name = %q", f.Section(DefaultSection).Key("server_name").String())
	}
}

// ---------------------------------------------------------------------------
// Roundtrip test
// ---------------------------------------------------------------------------

func TestRoundtrip(t *testing.T) {
	type Config struct {
		Name    string        `ini:"name"`
		Port    int           `ini:"port"`
		Debug   bool          `ini:"debug"`
		Timeout time.Duration `ini:"timeout"`
	}

	original := Config{
		Name:    "myapp",
		Port:    8080,
		Debug:   true,
		Timeout: 30 * time.Second,
	}

	// struct → INI → bytes
	f1 := Empty()
	if err := f1.ReflectFrom(&original); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := f1.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	// bytes → INI → struct
	f2, err := Load(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	var restored Config
	if err := f2.MapTo(&restored); err != nil {
		t.Fatal(err)
	}

	if restored != original {
		t.Errorf("roundtrip mismatch: got %+v, want %+v", restored, original)
	}
}

// ---------------------------------------------------------------------------
// Package-level convenience functions
// ---------------------------------------------------------------------------

func TestPackageLevelMapTo(t *testing.T) {
	data := []byte(`name = app`)
	type cfg struct {
		Name string `ini:"name"`
	}
	var c cfg
	if err := MapTo(&c, data); err != nil {
		t.Fatal(err)
	}
	if c.Name != "app" {
		t.Errorf("Name = %q", c.Name)
	}
}

func TestPackageLevelStrictMapTo(t *testing.T) {
	data := []byte(`
name = app
extra = val
`)
	type cfg struct {
		Name string `ini:"name"`
	}
	var c cfg
	if err := StrictMapTo(&c, data); err == nil {
		t.Error("expected strict error")
	}
}

func TestBoolSlice(t *testing.T) {
	type cfg struct {
		Flags []bool `ini:"flags" ini-delim:","`
	}
	data := []byte(`flags = true, false, yes`)
	f, err := Load(data)
	if err != nil {
		t.Fatal(err)
	}
	var c cfg
	if err := f.MapTo(&c); err != nil {
		t.Fatal(err)
	}
	if len(c.Flags) != 3 || c.Flags[0] != true || c.Flags[1] != false || c.Flags[2] != true {
		t.Errorf("Flags = %v", c.Flags)
	}
}
