package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDotEnvLoadsEnvFile(t *testing.T) {
	t.Chdir(t.TempDir())

	key := uniqueEnvKey("LOADS")
	if err := os.WriteFile(dotEnvFileName, []byte(fmt.Sprintf("%s=from-dotenv\n", key)), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	if err := LoadDotEnv(); err != nil {
		t.Fatalf("LoadDotEnv returned error: %v", err)
	}

	if got := os.Getenv(key); got != "from-dotenv" {
		t.Fatalf("expected %q, got %q", "from-dotenv", got)
	}
}

func TestLoadDotEnvIgnoresMissingFile(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := LoadDotEnv(); err != nil {
		t.Fatalf("LoadDotEnv returned error for missing file: %v", err)
	}
}

func TestLoadDotEnvDoesNotOverrideExistingEnv(t *testing.T) {
	t.Chdir(t.TempDir())

	key := uniqueEnvKey("OVERRIDE")
	if err := os.WriteFile(dotEnvFileName, []byte(fmt.Sprintf("%s=from-dotenv\n", key)), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	t.Setenv(key, "from-env")

	if err := LoadDotEnv(); err != nil {
		t.Fatalf("LoadDotEnv returned error: %v", err)
	}

	if got := os.Getenv(key); got != "from-env" {
		t.Fatalf("expected existing env to win, got %q", got)
	}
}

func TestLoadDotEnvReturnsParseError(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.WriteFile(dotEnvFileName, []byte("BROKEN='unterminated\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	if err := LoadDotEnv(); err == nil {
		t.Fatal("expected parse error, got nil")
	}
}

func TestLoadDotEnvFromPathIgnoresMissingFile(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), ".env")

	if err := loadDotEnv(missingPath); err != nil {
		t.Fatalf("loadDotEnv returned error for missing file: %v", err)
	}
}

func uniqueEnvKey(suffix string) string {
	return fmt.Sprintf("SCALA_BOT_TEST_%s_%d", suffix, time.Now().UnixNano())
}
