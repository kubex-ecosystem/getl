package sql

import (
	stdsql "database/sql"
	"os"
	"path/filepath"
	"testing"

	. "github.com/kubex-ecosystem/getl/etypes"
	_ "github.com/mattn/go-sqlite3"
)

func TestLoadDataFromCSVToSQLiteWithUpsert(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "people.csv")
	sqlitePath := filepath.Join(tempDir, "people.db")

	writeCSV := func(content string) {
		t.Helper()
		if err := os.WriteFile(csvPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write csv fixture: %v", err)
		}
	}

	config := Config{
		SourceType:                  "csv",
		SourceConnectionString:      csvPath,
		DestinationType:             "sqlite3",
		DestinationConnectionString: sqlitePath,
		DestinationTable:            "people",
		UpdateKey:                   "id",
		Transformations: []Transformation{
			{SourceField: "id", DestinationField: "id", Operation: "toInt"},
			{SourceField: "name", DestinationField: "full_name", Operation: "uppercase"},
			{SourceField: "city", DestinationField: "city", Operation: "none"},
		},
	}

	writeCSV("\uFEFFid;name;city\n1;Ana;SP\n2;Bruno;RJ\n")
	if err := LoadData(nil, config); err != nil {
		t.Fatalf("LoadData() first run error = %v", err)
	}

	writeCSV("\uFEFFid;name;city\n2;Carla;BH\n3;Diego;POA\n")
	if err := LoadData(nil, config); err != nil {
		t.Fatalf("LoadData() second run error = %v", err)
	}

	db, err := stdsql.Open("sqlite3", sqlitePath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, full_name, city FROM people ORDER BY id")
	if err != nil {
		t.Fatalf("failed to query sqlite table: %v", err)
	}
	defer rows.Close()

	type personRow struct {
		id       int
		fullName string
		city     string
	}

	var result []personRow
	for rows.Next() {
		var row personRow
		if err := rows.Scan(&row.id, &row.fullName, &row.city); err != nil {
			t.Fatalf("failed to scan row: %v", err)
		}
		result = append(result, row)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 rows after upsert, got %d", len(result))
	}

	expected := []personRow{
		{id: 1, fullName: "ANA", city: "SP"},
		{id: 2, fullName: "CARLA", city: "BH"},
		{id: 3, fullName: "DIEGO", city: "POA"},
	}

	for index, row := range result {
		if row != expected[index] {
			t.Fatalf("unexpected row at position %d: got %+v want %+v", index, row, expected[index])
		}
	}
}
