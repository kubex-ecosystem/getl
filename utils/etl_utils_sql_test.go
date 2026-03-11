package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigFileExpandsEnvironmentVariables(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	t.Setenv("TEST_GETL_SOURCE", filepath.Join(tempDir, "source.csv"))
	t.Setenv("TEST_GETL_DSN", "postgres://user:pass@localhost:5432/postgres?sslmode=disable")
	t.Setenv("TEST_GETL_TABLE", "sankhya_catalog.tdd_tabelas")

	content := `{
		"sourceType": "csv",
		"sourceConnectionString": "${TEST_GETL_SOURCE}",
		"destinationType": "postgres",
		"destinationConnectionString": "${TEST_GETL_DSN}",
		"destinationTable": "${TEST_GETL_TABLE}"
	}`

	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write config fixture: %v", err)
	}

	config, err := LoadConfigFile(configPath)
	if err != nil {
		t.Fatalf("LoadConfigFile() error = %v", err)
	}

	if got, want := config.SourceConnectionString, os.Getenv("TEST_GETL_SOURCE"); got != want {
		t.Fatalf("unexpected sourceConnectionString: got %q want %q", got, want)
	}
	if got, want := config.DestinationConnectionString, os.Getenv("TEST_GETL_DSN"); got != want {
		t.Fatalf("unexpected destinationConnectionString: got %q want %q", got, want)
	}
	if got, want := config.DestinationTable, os.Getenv("TEST_GETL_TABLE"); got != want {
		t.Fatalf("unexpected destinationTable: got %q want %q", got, want)
	}
}
