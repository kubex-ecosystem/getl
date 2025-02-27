package extr

import (
	. "github.com/faelmori/getl/etypes"
	"github.com/faelmori/logz"
	"github.com/pelletier/go-toml/v2"
	"os"
)

type TOMLDataTable struct {
	data         []Data
	filePath     string
	filteredData []Data
}

func NewTOMLDataTable(data []Data, filePath string) *TOMLDataTable {
	return &TOMLDataTable{
		data:     data,
		filePath: filePath,
	}
}

func (e *TOMLDataTable) LoadFile() error {
	file, err := os.Open(e.filePath)
	if err != nil {
		return logz.ErrorLog("Failed to open file: "+err.Error(), "etl", logz.QUIET)
	}
	defer file.Close()

	decoder := toml.NewDecoder(file)
	if err := decoder.Decode(&e.data); err != nil {
		return logz.ErrorLog("Failed to decode TOML: "+err.Error(), "etl", logz.QUIET)
	}

	return nil
}

func (e *TOMLDataTable) LoadData(data []Data) {
	e.data = data
}

func (e *TOMLDataTable) ExtractFile() error {
	file, err := os.Create(e.filePath)
	if err != nil {
		return logz.ErrorLog("Failed to create file: "+err.Error(), "etl", logz.QUIET)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(e.data); err != nil {
		return logz.ErrorLog("Failed to encode TOML: "+err.Error(), "etl", logz.QUIET)
	}

	return nil
}

func (e *TOMLDataTable) ExtractData(filter map[string]string) ([]Data, error) {
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

func (e *TOMLDataTable) ExtractDataByIndex(index int) (Data, error) {
	if len(e.data) == 0 {
		return nil, logz.ErrorLog("No data to extract", "etl", logz.QUIET)
	}

	if index < 0 || index >= len(e.data) {
		return nil, logz.ErrorLog("Invalid index", "etl", logz.QUIET)
	}

	return e.data[index], nil
}

func (e *TOMLDataTable) ExtractDataByRange(start, end int) ([]Data, error) {
	if len(e.data) == 0 {
		return nil, logz.ErrorLog("No data to extract", "etl", logz.QUIET)
	}

	if start < 0 || end < 0 || start >= len(e.data) || end >= len(e.data) {
		return nil, logz.ErrorLog("Invalid range", "etl", logz.QUIET)
	}

	return e.data[start:end], nil
}

func (e *TOMLDataTable) ExtractDataByField(field, value string) ([]Data, error) {
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
