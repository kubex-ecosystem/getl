package extr

import (
	"encoding/json"
	"fmt"
	"os"

	. "github.com/kubex-ecosystem/getl/etypes"
	gl "github.com/kubex-ecosystem/getl/internal/module/logger"
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
		gl.Log("error", "File not found: "+e.filePath)
		return err
	} else {
		openFile, openFileErr = os.Open(e.filePath)
	}

	if openFileErr != nil {
		gl.Log("error", "Failed to open file: "+openFileErr.Error())
		return openFileErr
	}
	defer openFile.Close()
	decoder := json.NewDecoder(openFile)

	if decodeErr := decoder.Decode(&e.data); decodeErr != nil {
		gl.Log("error", "Failed to decode data: "+decodeErr.Error())
		return decodeErr
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
		gl.Log("error", "No data to extract")
		return fmt.Errorf("no data to extract")
	}

	if _, err := os.Stat(e.filePath); err == nil {
		gl.Log("error", "File already exists: "+e.filePath)
		return fmt.Errorf("file already exists")
	} else {
		createFile, createFileErr = os.Create(e.filePath)
	}

	if createFileErr != nil {
		gl.Log("error", "Failed to create file: "+createFileErr.Error())
		return fmt.Errorf("failed to create file: %s", createFileErr.Error())
	}
	defer createFile.Close()
	encoder := json.NewEncoder(createFile)
	encoder.SetIndent("", "  ")

	if encodeErr := encoder.Encode(e.data); encodeErr != nil {
		gl.Log("error", "Failed to encode data: "+encodeErr.Error())
		return encodeErr
	}

	return nil
}

func (e *JSONDataTable) ExtractData(filter map[string]string) ([]Data, error) {
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

func (e *JSONDataTable) ExtractDataByIndex(index int) (Data, error) {
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

func (e *JSONDataTable) ExtractDataByRange(start, end int) ([]Data, error) {
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

func (e *JSONDataTable) ExtractDataByField(field, value string) ([]Data, error) {
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
