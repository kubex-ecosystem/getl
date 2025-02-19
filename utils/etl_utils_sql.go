package utils

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elgris/sqrl"
	. "github.com/faelmori/getl/etypes"
	"github.com/faelmori/kbx/mods/logz"
	"github.com/faelmori/kbx/mods/utils"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

func ApplyTransformations(data []Data, transformations []Transformation) ([]Data, error) {
	if transformations == nil {
		return data, nil
	}

	transformedData := make([]Data, len(data))
	for i, row := range data {
		transformedRow := make(Data)
		for _, t := range transformations {
			value, exists := row[t.SourceField]
			if !exists {
				return nil, fmt.Errorf("campo fonte não encontrado: %s", t.SourceField)
			}

			switch t.Operation {
			case "copy", "none":
				transformedRow[t.DestinationField] = value
			case "uppercase":
				if strValue, ok := value.(string); ok {
					transformedRow[t.DestinationField] = strings.ToUpper(strValue)
				} else {
					return nil, fmt.Errorf("valor não é uma string: %v", value)
				}
			case "base64":
				if strValue, ok := value.(string); ok {
					transformedRow[t.DestinationField] = base64.StdEncoding.EncodeToString([]byte(strValue))
				} else {
					return nil, fmt.Errorf("valor não é uma string: %v", value)
				}
			case "toInt":
				if strValue, ok := value.(string); ok {
					intValue, err := strconv.Atoi(strValue)
					if err != nil {
						return nil, fmt.Errorf("falha ao converter para inteiro: %w", err)
					}
					transformedRow[t.DestinationField] = intValue
				} else {
					return nil, fmt.Errorf("valor não é uma string: %v", value)
				}
			default:
				return nil, fmt.Errorf("operação desconhecida: %s", t.Operation)
			}
		}
		transformedData[i] = transformedRow
	}

	return transformedData, nil
}
func LoadFieldsFromTransformConfig(fileConfigPath string) (Fields, error) {
	var config Config

	_ = logz.InfoLog("Loading fields from file: "+fileConfigPath, "etl", logz.QUIET)
	fileData, fileDataErr := os.ReadFile(fileConfigPath)
	if fileDataErr != nil {
		return nil, logz.ErrorLog("failed to load file: "+fileDataErr.Error(), "etl")
	}
	_ = logz.InfoLog("File loaded successfully", "etl", logz.QUIET)

	_ = logz.InfoLog("Unmarshalling file data", "etl", logz.QUIET)
	unmarshalErr := json.Unmarshal(fileData, &config)
	if unmarshalErr != nil {
		return nil, logz.ErrorLog("334: "+unmarshalErr.Error(), "etl")
	}
	_ = logz.InfoLog("File data unmarshalled successfully", "etl", logz.QUIET)

	_ = logz.InfoLog("Creating fields map", "etl", logz.QUIET)
	var fields Fields

	for _, t := range config.Transformations {
		if fields == nil {
			fields = make(Fields)
		}

		if fields[t.SPath] == nil {
			fields[t.SPath] = []Field{}
		}

		fields[t.SPath] = append(fields[t.SPath], Field{"name": t.SourceField})

		if fields[t.DPath] == nil {
			fields[t.DPath] = []Field{}
		}

		fields[t.DPath] = append(fields[t.DPath], Field{"name": t.DestinationField})
	}

	_ = logz.InfoLog("Fields map created successfully: "+config.SourceType+" -> "+config.DestinationType, "etl", logz.QUIET)
	if maps.Values(fields) == nil {
		return nil, errors.New("failed to create sourceFields map: " + config.SourceType)
	}

	return fields, nil
}
func BuilExtractdQuery(config Config, fields []string) (string, []interface{}, error) {
	query := sqrl.Select(fields...).From(config.SourceTable)

	for _, join := range config.Joins {
		switch strings.ToUpper(join.JoinType) {
		case "INNER":
			query = query.Join(join.Table + " ON " + join.Condition)
		case "LEFT":
			query = query.LeftJoin(join.Table + " ON " + join.Condition)
		case "RIGHT":
			query = query.RightJoin(join.Table + " ON " + join.Condition)
		default:
			return "", nil, fmt.Errorf("tipo de join desconhecido: %s", join.JoinType)
		}
	}

	if config.Where != "" {
		query = query.Where(config.Where)
	}

	if config.OrderBy != "" {
		query = query.OrderBy(config.OrderBy)
	}

	return query.ToSql()
}
func LoadConfigFile(fileConfigPath string) (Config, error) {
	fileData, err := os.ReadFile(fileConfigPath)
	if err != nil {
		return Config{}, fmt.Errorf("falha ao ler o arquivo de configuração: %w", err)
	}

	var config Config
	if unmarshalErr := json.Unmarshal(fileData, &config); unmarshalErr != nil {
		return Config{}, fmt.Errorf("falha ao processar JSON de configuração: %w", unmarshalErr)
	}

	// Verificação de campos obrigatórios
	requiredFields := []string{"sourceType", "sourceConnectionString", "destinationType", "destinationConnectionString"}
	for _, field := range requiredFields {
		if reflect.ValueOf(config).FieldByName(strings.ToTitle(field)).String() == "" {
			return Config{}, fmt.Errorf("campo obrigatório ausente na configuração: %s", field)
		}
	}

	return config, nil
}
func GetDataTableHandlerFromQuery(sourceType, sourceConnectionString, sqlQuery string) (*TableHandler, error) {
	db, err := sql.Open(sourceType, sourceConnectionString)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao banco de dados de origem: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("falha ao executar a consulta SQL: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("falha ao obter colunas: %w", err)
	}

	var data [][]string
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("falha ao escanear linha: %w", err)
		}

		var row []string
		for _, value := range values {
			row = append(row, fmt.Sprintf("%v", value))
		}
		data = append(data, row)
	}

	return &TableHandler{Columns: columns, Data: data}, nil
}
func GenerateConfigTemplate(filePath string) error {
	config := Config{
		SourceType:                  "sqlite,postgres,mysql,oracle,sqlserver",
		SourceConnectionString:      "connection_string_for_source_database",
		SourceTable:                 "origin_table_name",
		DestinationType:             "sqlite,postgres,mysql,oracle,sqlserver",
		DestinationConnectionString: "connection_string_for_destination_database",
		DestinationTable:            "destination_table_name",
		SQLQuery:                    "SELECT * FROM your_table",
		OutputPath:                  "output_file_path",
		OutputFormat:                "json,csv,xml,parquet",
		Transformations: []Transformation{
			{
				SourceField:      "campo_origem",
				DestinationField: "campo_destino",
				Operation:        "none",
				SPath:            "caminho_origem",
				DPath:            "caminho_destino",
				Type:             "string",
			},
		},
		Joins: []Join{
			{
				Table:     "nome_da_tabela_join",
				Condition: "condicao_de_join",
				JoinType:  "INNER",
			},
		},
		Where:        "where_clause",
		OrderBy:      "order_by_clause",
		Triggers:     []Trigger{},
		LogTable:     "log_table_name",
		SyncInterval: "sync_interval",
		KafkaURL:     "kafka_broker_url",
		KafkaTopic:   "kafka_topic_name",
		KafkaGroupID: "kafka_group_id",
	}

	if filePath == "" {
		homeFilePath, filePathErr := utils.GetWorkDir()
		if filePathErr != nil {
			return fmt.Errorf("falha ao obter o diretório HOME: %w", filePathErr)
		}
		filePath = homeFilePath + "/.kubex/example_config.json"
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("falha ao criar o arquivo: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("falha ao codificar o JSON: %w", err)
	}

	return nil
}
func GetETLJobs() (JobList, error) {
	cwd, cwdErr := utils.GetWorkDir()
	if cwdErr != nil {
		return nil, logz.ErrorLog("failed to get current working directory: "+cwdErr.Error(), "etl", logz.QUIET)
	}
	jobsCwd := filepath.Join(cwd, "jobs")

	files, filesErr := os.ReadDir(jobsCwd)
	if filesErr != nil {
		return nil, logz.ErrorLog("failed to read jobs directory: "+filesErr.Error(), "etl", logz.QUIET)
	}

	var jobs VJobList
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(jobsCwd, file.Name())
		job, jobErr := LoadJobFromFile(filePath)
		if jobErr != nil {
			return nil, logz.ErrorLog("failed to load job from file: "+jobErr.Error(), "etl", logz.QUIET)
		}

		vJob := *job

		jobs.VJobs = append(jobs.VJobs, vJob)
	}

	return &jobs, nil
}
func LoadJobFromFile(filePath string) (*VJob, error) {
	fileData, fileDataErr := os.ReadFile(filePath)
	if fileDataErr != nil {
		return nil, logz.ErrorLog("failed to load file: "+fileDataErr.Error(), "etl", logz.QUIET)
	}

	var job VJob
	if unmarshalErr := json.Unmarshal(fileData, &job); unmarshalErr != nil {
		return nil, logz.ErrorLog("failed to unmarshal file data: "+unmarshalErr.Error(), "etl", logz.QUIET)
	}

	return &job, nil
}
