package extr

import (
	"encoding/json"
	. "github.com/faelmori/kbx/mods/getl/etypes"
	"github.com/faelmori/kbx/mods/logz"
	"os"
)

type JSONDataTable struct {
	data         []Data
	filePath     string
	filteredData []Data
}

func NewJSONDataTable(data []Data, filePath string) *JSONDataTable {
	return &JSONDataTable{
		data:     data,
		filePath: filePath,
	}
}

func (e *JSONDataTable) LoadFile() error {
	var openFile *os.File
	var openFileErr error

	if _, err := os.Stat(e.filePath); err != nil {
		return logz.ErrorLog("File not found: "+e.filePath, "etl", logz.QUIET)
	} else {
		openFile, openFileErr = os.Open(e.filePath)
	}

	if openFileErr != nil {
		return logz.ErrorLog("Failed to open file: "+openFileErr.Error(), "etl", logz.QUIET)
	}
	defer openFile.Close()
	decoder := json.NewDecoder(openFile)

	if decodeErr := decoder.Decode(&e.data); decodeErr != nil {
		return logz.ErrorLog("Failed to decode data: "+decodeErr.Error(), "etl", logz.QUIET)
	}

	return nil
}

func (e *JSONDataTable) LoadData(data []Data) {
	e.data = data
}

func (e *JSONDataTable) ExtractFile() error {
	var createFile *os.File
	var createFileErr error

	if len(e.data) == 0 {
		return logz.ErrorLog("No data to extract", "etl", logz.QUIET)
	}

	if _, err := os.Stat(e.filePath); err == nil {
		return logz.ErrorLog("File already exists: "+e.filePath, "etl", logz.QUIET)
	} else {
		createFile, createFileErr = os.Create(e.filePath)
	}

	if createFileErr != nil {
		return logz.ErrorLog("Failed to create file: "+createFileErr.Error(), "etl", logz.QUIET)
	}
	defer createFile.Close()
	encoder := json.NewEncoder(createFile)
	encoder.SetIndent("", "  ")

	if encodeErr := encoder.Encode(e.data); encodeErr != nil {
		return logz.ErrorLog("Failed to encode data: "+encodeErr.Error(), "etl", logz.QUIET)
	}

	return nil
}

func (e *JSONDataTable) ExtractData(filter map[string]string) ([]Data, error) {
	if len(e.data) == 0 {
		return nil, logz.ErrorLog("No data to extract", "etl", logz.QUIET)
	}

	if len(filter) == 0 {
		e.filteredData = e.data
	}

	for _, row := range e.data {
		for key, value := range filter {
			if row[key] == value {
				e.filteredData = append(e.filteredData, row)
			}
		}
	}

	if len(e.filteredData) == 0 {
		return nil, logz.ErrorLog("No data to extract", "etl", logz.QUIET)
	}

	return e.filteredData, nil
}

func (e *JSONDataTable) ExtractDataByIndex(index int) (Data, error) {
	if len(e.data) == 0 {
		return nil, logz.ErrorLog("No data to extract", "etl", logz.QUIET)
	}

	if index < 0 || index >= len(e.data) {
		return nil, logz.ErrorLog("Invalid index", "etl", logz.QUIET)
	}

	return e.data[index], nil
}

func (e *JSONDataTable) ExtractDataByRange(start, end int) ([]Data, error) {
	if len(e.data) == 0 {
		return nil, logz.ErrorLog("No data to extract", "etl", logz.QUIET)
	}

	if start < 0 || end < 0 || start >= len(e.data) || end >= len(e.data) {
		return nil, logz.ErrorLog("Invalid range", "etl", logz.QUIET)
	}

	return e.data[start:end], nil
}

func (e *JSONDataTable) ExtractDataByField(field, value string) ([]Data, error) {
	if len(e.data) == 0 {
		return nil, logz.ErrorLog("No data to extract", "etl", logz.QUIET)
	}

	var filteredData []Data
	for _, row := range e.data {
		if row[field] == value {
			filteredData = append(filteredData, row)
		}
	}

	if len(filteredData) == 0 {
		return nil, logz.ErrorLog("No data to extract", "etl", logz.QUIET)
	}

	return filteredData, nil
}
