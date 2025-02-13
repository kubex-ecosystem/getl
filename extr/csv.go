package extr

import (
	"encoding/csv"
	. "github.com/faelmori/kbx/mods/getl/etypes"
	"github.com/faelmori/kbx/mods/logz"
	"os"
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
		return logz.ErrorLog("File not found: "+e.filePath, "etl", logz.QUIET)
	} else {
		openFile, openFileErr = os.Open(e.filePath)
	}

	if openFileErr != nil {
		return logz.ErrorLog("Failed to open file: "+openFileErr.Error(), "etl", logz.QUIET)
	}
	defer openFile.Close()
	reader := csv.NewReader(openFile)
	records, readErr := reader.ReadAll()
	if readErr != nil {
		return logz.ErrorLog("Failed to read CSV: "+readErr.Error(), "etl", logz.QUIET)
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
		return logz.ErrorLog("Failed to create file: "+err.Error(), "etl", logz.QUIET)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var headers []string
	for key := range e.data[0] {
		headers = append(headers, key)
	}
	if writerErr := writer.Write(headers); writerErr != nil {
		return logz.ErrorLog("Failed to write headers to CSV: "+writerErr.Error(), "etl", logz.QUIET)
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
			return logz.ErrorLog("Failed to write row to CSV: "+writerRowsErr.Error(), "etl", logz.QUIET)
		}
	}
	return nil
}

func (e *CSVDataTable) ExtractData(filter map[string]string) ([]Data, error) {
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

func (e *CSVDataTable) ExtractDataByIndex(index int) (Data, error) {
	if len(e.data) == 0 {
		return nil, logz.ErrorLog("No data to extract", "etl", logz.QUIET)
	}

	if index < 0 || index >= len(e.data) {
		return nil, logz.ErrorLog("Invalid index", "etl", logz.QUIET)
	}

	return e.data[index], nil
}

func (e *CSVDataTable) ExtractDataByRange(start, end int) ([]Data, error) {
	if len(e.data) == 0 {
		return nil, logz.ErrorLog("No data to extract", "etl", logz.QUIET)
	}

	if start < 0 || end < 0 || start >= len(e.data) || end >= len(e.data) {
		return nil, logz.ErrorLog("Invalid range", "etl", logz.QUIET)
	}

	return e.data[start:end], nil
}

func (e *CSVDataTable) ExtractDataByField(field, value string) ([]Data, error) {
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
