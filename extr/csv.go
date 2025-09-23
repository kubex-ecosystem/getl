package extr

import (
	"encoding/csv"
	"fmt"
	"os"

	. "github.com/kubex-ecosystem/getl/etypes"
	"github.com/kubex-ecosystem/logz"
)

type CSVDataTable struct {
	data         []Data
	filePath     string
	filteredData []Data
}

func NewCSVDataTable(data []Data, filePath string) *CSVDataTable {
	return &CSVDataTable{
		data:     data,
		filePath: filePath,
	}
}

func (e *CSVDataTable) LoadFile() error {
	var openFile *os.File
	var openFileErr error

	if _, err := os.Stat(e.filePath); err != nil {
		logz.Error("File not found: "+e.filePath, map[string]interface{}{})
		return err
	} else {
		openFile, openFileErr = os.Open(e.filePath)
	}

	if openFileErr != nil {
		logz.Error("Failed to open file: "+openFileErr.Error(), map[string]interface{}{})
		return openFileErr
	}
	defer openFile.Close()
	reader := csv.NewReader(openFile)
	records, readErr := reader.ReadAll()
	if readErr != nil {
		logz.Error("Failed to read CSV: "+readErr.Error(), map[string]interface{}{})
		return readErr
	}

	for i, row := range records {
		if i == 0 {
			continue
		}
		data := make(Data)
		for j, value := range row {
			data[records[0][j]] = value
		}
		e.data = append(e.data, data)
	}

	return nil
}

func (e *CSVDataTable) LoadData(data []Data) {
	e.data = data
}

func (e *CSVDataTable) ExtractFile() error {
	file, err := os.Create(e.filePath)
	if err != nil {
		logz.Error("Failed to create file: "+err.Error(), map[string]interface{}{})
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var headers []string
	for key := range e.data[0] {
		headers = append(headers, key)
	}
	if writerErr := writer.Write(headers); writerErr != nil {
		logz.Error("Failed to write headers to CSV: "+writerErr.Error(), map[string]interface{}{})
		return writerErr
	}
	for _, row := range e.data {
		var rowData []string
		for _, value := range row {
			if value == nil {
				rowData = append(rowData, "")
				continue
			}
			rowData = append(rowData, value.(string))
		}

		if writerRowsErr := writer.Write(rowData); writerRowsErr != nil {
			logz.Error("Failed to write row to CSV: "+writerRowsErr.Error(), map[string]interface{}{})
			return writerRowsErr
		}
	}
	return nil
}

func (e *CSVDataTable) ExtractData(filter map[string]string) ([]Data, error) {
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

func (e *CSVDataTable) ExtractDataByIndex(index int) (Data, error) {
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

func (e *CSVDataTable) ExtractDataByRange(start, end int) ([]Data, error) {
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

func (e *CSVDataTable) ExtractDataByField(field, value string) ([]Data, error) {
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
