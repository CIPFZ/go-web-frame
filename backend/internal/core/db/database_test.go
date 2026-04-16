package db

import (
	"testing"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"go.uber.org/zap"
)

func TestInitDatabaseRejectsUnsupportedDriver(t *testing.T) {
	_, err := InitDatabase(config.Database{Driver: "oracle"}, zap.NewNop())
	if err == nil {
		t.Fatal("InitDatabase() error = nil, want unsupported driver error")
	}
}

func TestNormalizeDriverAliases(t *testing.T) {
	tests := map[string]string{
		"":           "mysql",
		"mysql":      "mysql",
		"postgres":   "postgres",
		"postgresql": "postgres",
		"pgsql":      "postgres",
	}
	for in, want := range tests {
		if got := normalizeDriver(in); got != want {
			t.Fatalf("normalizeDriver(%q) = %q, want %q", in, got, want)
		}
	}
}
