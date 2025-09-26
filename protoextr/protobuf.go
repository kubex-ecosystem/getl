package protoextr

import (
	"os"

	gl "github.com/kubex-ecosystem/getl/internal/module/logger"
	"google.golang.org/protobuf/proto"
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
		gl.Log("error", "Failed to open file: "+err.Error())
		return err
	}

	var dataList DataList
	if err := proto.Unmarshal(file, &dataList); err != nil {
		gl.Log("error", "Failed to decode Protobuf: "+err.Error())
		return err
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
		gl.Log("error", "Failed to encode Protobuf: "+err.Error())
		return err
	}

	if err := os.WriteFile(e.filePath, data, 0644); err != nil {
		gl.Log("error", "Failed to write file: "+err.Error())
		return err
	}

	return nil
}
