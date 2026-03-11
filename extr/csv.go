package extr

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	. "github.com/kubex-ecosystem/getl/etypes"
	gl "github.com/kubex-ecosystem/getl/internal/module/logger"
)

type CSVDataTable struct {
	data         []Data
	headers      []string
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

	reader := csv.NewReader(openFile)
	reader.Comma = detectCSVDelimiter(openFile)
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	records, readErr := reader.ReadAll()
	if readErr != nil {
		gl.Log("error", "Failed to read CSV: "+readErr.Error())
		return readErr
	}
	if len(records) == 0 {
		return fmt.Errorf("csv vazio: %s", e.filePath)
	}

	e.headers = normalizeCSVHeaders(records[0])

	for i, row := range records {
		if i == 0 {
			continue
		}
		data := make(Data)
		for j, header := range e.headers {
			if j >= len(row) {
				data[header] = ""
				continue
			}
			data[header] = strings.TrimSpace(stripUTF8BOM(row[j]))
		}
		e.data = append(e.data, data)
	}

	return nil
}

func (e *CSVDataTable) LoadData(data []Data) {
	e.data = data
}

func (e *CSVDataTable) Headers() []string {
	return append([]string(nil), e.headers...)
}

func (e *CSVDataTable) ExtractFile() error {
	file, err := os.Create(e.filePath)
	if err != nil {
		gl.Log("error", "Failed to create file: "+err.Error())
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var headers []string
	if len(e.headers) > 0 {
		headers = append(headers, e.headers...)
	} else {
		for key := range e.data[0] {
			headers = append(headers, key)
		}
	}
	if writerErr := writer.Write(headers); writerErr != nil {
		gl.Log("error", "Failed to write headers to CSV: "+writerErr.Error())
		return writerErr
	}
	for _, row := range e.data {
		var rowData []string
		for _, header := range headers {
			value := row[header]
			if value == nil {
				rowData = append(rowData, "")
				continue
			}
			rowData = append(rowData, fmt.Sprintf("%v", value))
		}

		if writerRowsErr := writer.Write(rowData); writerRowsErr != nil {
			gl.Log("error", "Failed to write row to CSV: "+writerRowsErr.Error())
			return writerRowsErr
		}
	}
	return nil
}

func (e *CSVDataTable) ExtractData(filter map[string]string) ([]Data, error) {
	if len(e.data) == 0 {
		gl.Log("error", "No data to extract")
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
		gl.Log("error", "No data to extract")
		return nil, fmt.Errorf("No data to extract")
	}

	return e.filteredData, nil
}

func (e *CSVDataTable) ExtractDataByIndex(index int) (Data, error) {
	if len(e.data) == 0 {
		gl.Log("error", "No data to extract")
		return nil, fmt.Errorf("No data to extract")
	}

	if index < 0 || index >= len(e.data) {
		gl.Log("error", "Invalid index")
		return nil, fmt.Errorf("Invalid index")
	}

	return e.data[index], nil
}

func (e *CSVDataTable) ExtractDataByRange(start, end int) ([]Data, error) {
	if len(e.data) == 0 {
		gl.Log("error", "No data to extract")
		return nil, fmt.Errorf("No data to extract")
	}

	if start < 0 || end < 0 || start >= len(e.data) || end >= len(e.data) {
		gl.Log("error", "Invalid range")
		return nil, fmt.Errorf("Invalid range")
	}

	return e.data[start:end], nil
}

func (e *CSVDataTable) ExtractDataByField(field, value string) ([]Data, error) {
	if len(e.data) == 0 {
		gl.Log("error", "No data to extract")
		return nil, fmt.Errorf("No data to extract")
	}

	var filteredData []Data
	for _, row := range e.data {
		if row[field] == value {
			filteredData = append(filteredData, row)
		}
	}

	if len(filteredData) == 0 {
		gl.Log("error", "No data to extract")
		return nil, fmt.Errorf("No data to extract")
	}

	return filteredData, nil
}

func detectCSVDelimiter(file *os.File) rune {
	if _, err := file.Seek(0, 0); err != nil {
		return ','
	}

	reader := bufio.NewReader(file)
	firstLine, err := reader.ReadString('\n')
	if err != nil && firstLine == "" {
		_, _ = file.Seek(0, 0)
		return ','
	}
	_, _ = file.Seek(0, 0)

	firstLine = stripUTF8BOM(firstLine)
	candidates := []rune{';', ',', '\t', '|'}
	best := ','
	bestCount := -1

	for _, candidate := range candidates {
		count := strings.Count(firstLine, string(candidate))
		if count > bestCount {
			best = candidate
			bestCount = count
		}
	}

	return best
}

func normalizeCSVHeaders(headers []string) []string {
	normalized := make([]string, 0, len(headers))
	for _, header := range headers {
		header = strings.TrimSpace(stripUTF8BOM(header))
		normalized = append(normalized, header)
	}
	return normalized
}

func stripUTF8BOM(value string) string {
	return strings.TrimPrefix(value, "\uFEFF")
}
