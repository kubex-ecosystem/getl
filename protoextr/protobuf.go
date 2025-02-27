package protoextr

import (
	"github.com/faelmori/logz"
	"google.golang.org/protobuf/proto"
	"os"
)

type ProtobufDataTable struct {
	data     []*Data
	filePath string
}

func NewProtobufDataTable(data []*Data, filePath string) *ProtobufDataTable {
	return &ProtobufDataTable{
		data:     data,
		filePath: filePath,
	}
}

func (e *ProtobufDataTable) LoadFile() error {
	file, err := os.ReadFile(e.filePath)
	if err != nil {
		return logz.ErrorLog("Failed to open file: "+err.Error(), "etl", logz.QUIET)
	}

	var dataList DataList
	if err := proto.Unmarshal(file, &dataList); err != nil {
		return logz.ErrorLog("Failed to decode Protobuf: "+err.Error(), "etl", logz.QUIET)
	}

	e.data = dataList.Data
	return nil
}

func (e *ProtobufDataTable) LoadData(data []*Data) {
	e.data = data
}

func (e *ProtobufDataTable) ExtractFile() error {
	dataList := &DataList{Data: e.data}

	data, err := proto.Marshal(dataList)
	if err != nil {
		return logz.ErrorLog("Failed to encode Protobuf: "+err.Error(), "etl", logz.QUIET)
	}

	if err := os.WriteFile(e.filePath, data, 0644); err != nil {
		return logz.ErrorLog("Failed to write file: "+err.Error(), "etl", logz.QUIET)
	}

	return nil
}
