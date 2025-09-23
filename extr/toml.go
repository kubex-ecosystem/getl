package extr

import (
	"fmt"
	"os"

	. "github.com/kubex-ecosystem/getl/etypes"
	"github.com/kubex-ecosystem/logz"
	"github.com/pelletier/go-toml/v2"
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
		return err //logz.Error("Failed to open file: "+err.Error(), map[string]interface{}{})
	}
	defer file.Close()

	decoder := toml.NewDecoder(file)
	if err := decoder.Decode(&e.data); err != nil {
		return err //logz.Error("Failed to decode TOML: "+err.Error(), map[string]interface{}{})
	}

	return nil
}

func (e *TOMLDataTable) LoadData(data []Data) {
	e.data = data
}

func (e *TOMLDataTable) ExtractFile() error {
	file, err := os.Create(e.filePath)
	if err != nil {
		return err //logz.Error("Failed to create file: "+err.Error(), map[string]interface{}{})
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(e.data); err != nil {
		return err //logz.Error("Failed to encode TOML: "+err.Error(), map[string]interface{}{})
	}

	return nil
}

func (e *TOMLDataTable) ExtractData(filter map[string]string) ([]Data, error) {
	if len(e.data) == 0 {
		logz.Error("No data to extract", map[string]interface{}{})
		return nil, fmt.Errorf("No data to extract")
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
		logz.Error("No data to extract", map[string]interface{}{})
		return nil, fmt.Errorf("No data to extract")
	}

	return e.filteredData, nil
}

func (e *TOMLDataTable) ExtractDataByIndex(index int) (Data, error) {
	if len(e.data) == 0 {
		logz.Error("No data to extract", map[string]interface{}{})
		return nil, fmt.Errorf("No data to extract")
	}

	if index < 0 || index >= len(e.data) {
		logz.Error("Invalid index", map[string]interface{}{})
		return nil, fmt.Errorf("Invalid index")
	}

	return e.data[index], nil
}

func (e *TOMLDataTable) ExtractDataByRange(start, end int) ([]Data, error) {
	if len(e.data) == 0 {
		logz.Error("No data to extract", map[string]interface{}{})
		return nil, fmt.Errorf("No data to extract")
	}

	if start < 0 || end < 0 || start >= len(e.data) || end >= len(e.data) {
		logz.Error("Invalid range", map[string]interface{}{})
		return nil, fmt.Errorf("Invalid range")
	}

	return e.data[start:end], nil
}

func (e *TOMLDataTable) ExtractDataByField(field, value string) ([]Data, error) {
	if len(e.data) == 0 {
		logz.Error("No data to extract", map[string]interface{}{})
		return nil, fmt.Errorf("No data to extract")
	}

	var filteredData []Data
	for _, row := range e.data {
		if row[field] == value {
			filteredData = append(filteredData, row)
		}
	}

	if len(filteredData) == 0 {
		logz.Error("No data to extract", map[string]interface{}{})
		return nil, fmt.Errorf("No data to extract")
	}

	return filteredData, nil
}
