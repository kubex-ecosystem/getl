package extr

import (
	. "github.com/faelmori/getl/etypes"
	"github.com/faelmori/logz"
	"gopkg.in/yaml.v2"
	"os"
)

type YAMLDataTable struct {
	data         []Data
	filePath     string
	filteredData []Data
}

func NewYAMLDataTable(data []Data, filePath string) *YAMLDataTable {
	return &YAMLDataTable{
		data:     data,
		filePath: filePath,
	}
}

func (e *YAMLDataTable) LoadFile() error {
	file, err := os.Open(e.filePath)
	if err != nil {
		return logz.ErrorLog("Failed to open file: "+err.Error(), "etl", logz.QUIET)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&e.data); err != nil {
		return logz.ErrorLog("Failed to decode YAML: "+err.Error(), "etl", logz.QUIET)
	}

	return nil
}

func (e *YAMLDataTable) LoadData(data []Data) {
	e.data = data
}

func (e *YAMLDataTable) ExtractFile() error {
	file, err := os.Create(e.filePath)
	if err != nil {
		return logz.ErrorLog("Failed to create file: "+err.Error(), "etl", logz.QUIET)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	if err := encoder.Encode(e.data); err != nil {
		return logz.ErrorLog("Failed to encode YAML: "+err.Error(), "etl", logz.QUIET)
	}

	return nil
}

func (e *YAMLDataTable) ExtractData(filter map[string]string) ([]Data, error) {
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

func (e *YAMLDataTable) ExtractDataByIndex(index int) (Data, error) {
	if len(e.data) == 0 {
		return nil, logz.ErrorLog("No data to extract", "etl", logz.QUIET)
	}

	if index < 0 || index >= len(e.data) {
		return nil, logz.ErrorLog("Invalid index", "etl", logz.QUIET)
	}

	return e.data[index], nil
}

func (e *YAMLDataTable) ExtractDataByRange(start, end int) ([]Data, error) {
	if len(e.data) == 0 {
		return nil, logz.ErrorLog("No data to extract", "etl", logz.QUIET)
	}

	if start < 0 || end < 0 || start >= len(e.data) || end >= len(e.data) {
		return nil, logz.ErrorLog("Invalid range", "etl", logz.QUIET)
	}

	return e.data[start:end], nil
}

func (e *YAMLDataTable) ExtractDataByField(field, value string) ([]Data, error) {
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
