package stanza

import (
	"bytes"
	"testing"
)

var benchINI = []byte(`
[general]
app_name = MyApp
version = 1.2.3
debug = false

[server]
host = 0.0.0.0
port = 8080
timeout = 30s
max_connections = 1000

[database]
driver = postgres
host = db.example.com
port = 5432
user = admin
password = secret
name = mydb
pool_size = 20
ssl_mode = require

[cache]
enabled = true
driver = redis
host = cache.example.com
port = 6379
ttl = 300

[logging]
level = info
file = /var/log/app.log
max_size = 100
max_backups = 3
`)

func BenchmarkLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := Load(benchINI)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLoadLargeFile(b *testing.B) {
	// Build a larger INI with 50 sections, 10 keys each.
	var buf bytes.Buffer
	for s := 0; s < 50; s++ {
		buf.WriteString("[section_")
		buf.WriteString(string(rune('A' + s%26)))
		buf.WriteString("]\n")
		for k := 0; k < 10; k++ {
			buf.WriteString("key_")
			buf.WriteByte(byte('0' + k))
			buf.WriteString(" = value_")
			buf.WriteByte(byte('0' + k))
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}
	data := buf.Bytes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Load(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteTo(b *testing.B) {
	f, err := Load(benchINI)
	if err != nil {
		b.Fatal(err)
	}
	var buf bytes.Buffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_, err := f.WriteTo(&buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkKeyAccess(b *testing.B) {
	f, err := Load(benchINI)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = f.Section("server").Key("host").String()
		_ = f.Section("database").Key("port").MustInt()
		_ = f.Section("cache").Key("enabled").MustBool()
	}
}

func BenchmarkMapTo(b *testing.B) {
	type Config struct {
		Server struct {
			Host           string `ini:"host"`
			Port           int    `ini:"port"`
			Timeout        string `ini:"timeout"`
			MaxConnections int    `ini:"max_connections"`
		} `ini:"server"`
	}
	f, err := Load(benchINI)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cfg Config
		_ = f.MapTo(&cfg)
	}
}
