package extr

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCSVDataTableLoadFileDetectsDelimiterAndBOM(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "sample.csv")
	content := "\uFEFFid;name;city\n1;Ana;SP\n2;Bruno;RJ\n"

	if err := os.WriteFile(csvPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write csv fixture: %v", err)
	}

	table := NewCSVDataTable(nil, csvPath)
	if err := table.LoadFile(); err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	rows, err := table.ExtractData(nil)
	if err != nil {
		t.Fatalf("ExtractData() error = %v", err)
	}

	headers := table.Headers()
	if len(headers) != 3 {
		t.Fatalf("expected 3 headers, got %d", len(headers))
	}
	if headers[0] != "id" {
		t.Fatalf("expected first header to be cleaned from BOM, got %q", headers[0])
	}

	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0]["name"] != "Ana" {
		t.Fatalf("expected first row name Ana, got %v", rows[0]["name"])
	}
	if rows[1]["city"] != "RJ" {
		t.Fatalf("expected second row city RJ, got %v", rows[1]["city"])
	}
}
