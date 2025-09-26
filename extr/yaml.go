package extr

import (
	"fmt"
	"os"

	. "github.com/kubex-ecosystem/getl/etypes"
	gl "github.com/kubex-ecosystem/getl/internal/module/logger"
	"gopkg.in/yaml.v2"
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
		gl.Log("error", "Failed to open file: "+err.Error())
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&e.data); err != nil {
		gl.Log("error", "Failed to decode YAML: "+err.Error())
		return err
	}

	return nil
}

func (e *YAMLDataTable) LoadData(data []Data) {
	e.data = data
}

func (e *YAMLDataTable) ExtractFile() error {
	file, err := os.Create(e.filePath)
	if err != nil {
		gl.Log("error", "Failed to create file: "+err.Error())
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	if err := encoder.Encode(e.data); err != nil {
		gl.Log("error", "Failed to encode YAML: "+err.Error())
		return err
	}

	return nil
}

func (e *YAMLDataTable) ExtractData(filter map[string]string) ([]Data, error) {
	if len(e.data) == 0 {
		gl.Log("error", "No data to extract")
		return nil, fmt.Errorf("no data to extract")
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
		gl.Log("error", "No data to extract")
		return nil, fmt.Errorf("no data to extract")
	}

	return e.filteredData, nil
}

func (e *YAMLDataTable) ExtractDataByIndex(index int) (Data, error) {
	if len(e.data) == 0 {
		gl.Log("error", "No data to extract")
		return nil, fmt.Errorf("no data to extract")
	}

	if index < 0 || index >= len(e.data) {
		gl.Log("error", "Invalid index")
		return nil, fmt.Errorf("invalid index")
	}

	return e.data[index], nil
}

func (e *YAMLDataTable) ExtractDataByRange(start, end int) ([]Data, error) {
	if len(e.data) == 0 {
		gl.Log("error", "No data to extract")
		return nil, fmt.Errorf("no data to extract")
	}

	if start < 0 || end < 0 || start >= len(e.data) || end >= len(e.data) {
		gl.Log("error", "Invalid range")
		return nil, fmt.Errorf("invalid range")
	}

	return e.data[start:end], nil
}

func (e *YAMLDataTable) ExtractDataByField(field, value string) ([]Data, error) {
	if len(e.data) == 0 {
		gl.Log("error", "No data to extract")
		return nil, fmt.Errorf("no data to extract")
	}

	var filteredData []Data
	for _, row := range e.data {
		if row[field] == value {
			filteredData = append(filteredData, row)
		}
	}

	if len(filteredData) == 0 {
		gl.Log("error", "No data to extract")
		return nil, fmt.Errorf("no data to extract")
	}

	return filteredData, nil
}
