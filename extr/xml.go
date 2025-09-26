package extr

import (
	"encoding/xml"
	"fmt"
	"os"

	. "github.com/kubex-ecosystem/getl/etypes"
	gl "github.com/kubex-ecosystem/getl/internal/module/logger"
)

type XMLRow struct {
	XMLName xml.Name `xml:"row"`
	Fields  []XMLField
}

type XMLField struct {
	XMLName xml.Name `xml:"field"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",chardata"`
}

type XMLData struct {
	XMLName xml.Name `xml:"data"`
	Rows    []XMLRow `xml:"row"`
}

type XMLDataTable struct {
	data         []Data
	filePath     string
	filteredData []Data
}

func NewXMLDataTable(data []Data, filePath string) *XMLDataTable {
	return &XMLDataTable{
		data:     data,
		filePath: filePath,
	}
}

func (e *XMLDataTable) LoadFile() error {
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
	decoder := xml.NewDecoder(openFile)

	var xmlData XMLData
	if decodeErr := decoder.Decode(&xmlData); decodeErr != nil {
		gl.Log("error", "Failed to decode data: "+decodeErr.Error())
		return decodeErr
	}

	for _, xmlRow := range xmlData.Rows {
		var row Data
		for _, xmlField := range xmlRow.Fields {
			row[xmlField.Name] = xmlField.Value
		}
		e.data = append(e.data, row)
	}

	return nil
}

func (e *XMLDataTable) LoadData(data []Data) {
	e.data = data
}

func (e *XMLDataTable) ExtractFile() error {
	var xmlData XMLData

	for _, row := range e.data {
		var xmlRow XMLRow
		for key, value := range row {
			xmlField := XMLField{
				Name:  key,
				Value: value.(string),
			}
			xmlRow.Fields = append(xmlRow.Fields, xmlField)
		}
		xmlData.Rows = append(xmlData.Rows, xmlRow)
	}

	file, err := os.Create(e.filePath)
	if err != nil {
		gl.Log("error", "Failed to create file: "+err.Error())
		return err
	}
	defer file.Close()

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	if err := encoder.Encode(xmlData); err != nil {
		gl.Log("error", "Failed to write XML: "+err.Error())
		return err
	}

	return nil
}

func (e *XMLDataTable) ExtractData(filter map[string]string) ([]Data, error) {
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

func (e *XMLDataTable) ExtractDataByIndex(index int) (Data, error) {
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

func (e *XMLDataTable) ExtractDataByRange(start, end int) ([]Data, error) {
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

func (e *XMLDataTable) ExtractDataByField(field, value string) ([]Data, error) {
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
