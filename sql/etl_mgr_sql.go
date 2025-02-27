package sql

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/denisenkom/go-mssqldb"
	. "github.com/faelmori/getl/etypes"
	. "github.com/faelmori/getl/utils"
	ui "github.com/faelmori/kbx/mods/ui/components"
	"github.com/faelmori/kbx/mods/utils"
	"github.com/faelmori/logz"
	"github.com/goccy/go-json"
	_ "github.com/godror/godror"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"time"
)

func ShowDataTableFromConfig(fileConfigPath string, export bool, exportPath string, outputFormat string) error {
	config, err := LoadConfigFile(fileConfigPath)
	if err != nil {
		return fmt.Errorf("falha ao carregar configuração da fonte: %w", err)
	}

	var sqlQuery string
	if config.SQLQuery != "" {
		sqlQuery = config.SQLQuery
	} else {
		fields := []string{"*"} // Ajuste conforme necessário
		sqlQuery, _, err = BuilExtractdQuery(config, fields)
		if err != nil {
			return fmt.Errorf("falha ao construir a consulta SQL: %w", err)
		}
	}

	handler, err := GetDataTableHandlerFromQuery(config.SourceType, config.SourceConnectionString, sqlQuery)
	if err != nil {
		return err
	}

	if export {
		if exportPath == "" {
			return fmt.Errorf("caminho de exportação não fornecido")
		}

		var data []Data
		for _, row := range handler.Data {
			rowData := make(Data)
			for i, value := range row {
				rowData[handler.Columns[i]] = value
			}
			data = append(data, rowData)
		}

		exportErr := SaveData(exportPath, data, outputFormat)
		if exportErr != nil {
			return fmt.Errorf("falha ao exportar dados para arquivo: %w", exportErr)
		}
		return nil
	}

	customStyles := map[string]lipgloss.Color{
		"header": lipgloss.Color("#01BE85"),
		"row":    lipgloss.Color("#252"),
	}
	return ui.StartTableScreen(handler, customStyles)
}
func ExtractDataWithTypes(dbSQL *sql.DB, config Config) ([]Data, map[string]string, error) {
	var db *sql.DB
	var dbErr error
	if dbSQL == nil {
		db, dbErr = sql.Open(config.SourceType, config.SourceConnectionString)
		if dbErr != nil {
			return nil, nil, logz.ErrorLog("Failed to connect to source database: "+dbErr.Error(), "etl", logz.QUIET)
		}
	} else {
		db = dbSQL
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	_ = logz.InfoLog("Starting data extraction", "etl", logz.QUIET)

	var rows *sql.Rows
	var SQLQueryArgs []interface{}
	var rowsErr error
	var buildQueryErr error

	if config.SQLQuery == "" {
		var fields []string
		var transformationsList []Transformation
		transformationsList = config.Transformations
		for i, t := range transformationsList {
			fields = append(fields, t.SourceField)
			if t.Type == "" {
				transformationsList[i].Type = "string"
			}
		}
		config.SQLQuery, SQLQueryArgs, buildQueryErr = BuilExtractdQuery(config, fields)
		if buildQueryErr != nil {
			return nil, nil, logz.ErrorLog("Failed to build query: "+buildQueryErr.Error(), "etl", logz.QUIET)
		}
	}

	_ = logz.DebugLog("Running query: "+config.SQLQuery, "etl", logz.QUIET)

	if len(SQLQueryArgs) > 0 {
		rows, rowsErr = db.Query(config.SQLQuery, SQLQueryArgs...)
	} else {
		rows, rowsErr = db.Query(config.SQLQuery)
	}

	if rowsErr != nil {
		return nil, nil, logz.ErrorLog("Failed on query execution: "+rowsErr.Error(), "etl", logz.QUIET)
	}

	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var data []Data
	columns, columnsErr := rows.Columns()
	if columnsErr != nil {
		return nil, nil, logz.ErrorLog("Failed to get columns: "+columnsErr.Error(), "etl", logz.QUIET)
	}

	columnTypes, columnTypesErr := rows.ColumnTypes()
	if columnTypesErr != nil {
		return nil, nil, logz.ErrorLog("Failed trying to get column types: "+columnTypesErr.Error(), "etl", logz.QUIET)
	}

	columnTypeMap := make(map[string]string)
	for i, colType := range columnTypes {
		columnTypeMap[columns[i]] = colType.DatabaseTypeName()
	}

	for rows.Next() {
		rowData := make([]interface{}, len(columns))
		rowPointers := make([]interface{}, len(columns))
		for i := range rowData {
			rowPointers[i] = &rowData[i]
		}

		if scanErr := rows.Scan(rowPointers...); scanErr != nil {
			return nil, nil, logz.ErrorLog("Failed to scan row data: "+scanErr.Error(), "etl", logz.QUIET)
		}

		row := make(Data)
		for i, colName := range columns {
			row[colName] = rowData[i]
		}
		data = append(data, row)
	}

	return data, columnTypeMap, nil
}
func EnsureTableExistsWithTypes(db *sql.DB, config Config, fields map[string]string) error {
	if config.DestinationTable == "" {
		return logz.ErrorLog("nome da tabela não informado", "etl", logz.QUIET)
	}

	var createTableQuery string
	var fieldsDest = make(map[string]string)
	createTableQuery = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", config.DestinationTable)
	for fieldName, fieldType := range fields {
		typeName := GetVendorSqlType(
			config.DestinationType,
			fieldType,
		)
		if typeName == "" {
			return logz.ErrorLog(fmt.Sprintf("tipo de campo não mapeado: %s", fieldType), "etl", logz.QUIET)
		}
		if config.UpdateKey == fieldName {
			createTableQuery += fmt.Sprintf("%s %s %s, ", fieldName, typeName, "PRIMARY KEY")
		} else {
			createTableQuery += fmt.Sprintf("%s %s, ", fieldName, typeName)
		}
		fieldsDest[fieldName] = typeName
	}
	createTableQuery = createTableQuery[:len(createTableQuery)-2] + ")"

	_ = logz.DebugLog("Campos de destino: "+config.DestinationType+" - "+fmt.Sprintf("%v", fieldsDest), "etl", logz.QUIET)

	_, createTableQueryErr := db.Exec(createTableQuery)
	if createTableQueryErr != nil {
		return logz.ErrorLog(fmt.Sprintf("falha ao criar a tabela: %v", createTableQueryErr), "etl", logz.QUIET)
	}

	return nil
}
func ExtractData(dbSQL *sql.DB, config Config) ([]Data, []string, error) {
	if config.SQLQuery == "" {
		return nil, nil, logz.ErrorLog("query SQL não informada", "etl", logz.QUIET)
	}

	var db *sql.DB
	var dbErr error
	if dbSQL == nil {
		db, dbErr = sql.Open(config.SourceType, config.SourceConnectionString)
		if dbErr != nil {
			return nil, nil, logz.ErrorLog(fmt.Sprintf("falha ao conectar ao banco de dados: %v", dbErr), "etl", logz.QUIET)
		}
	} else {
		db = dbSQL
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	rows, queryErr := db.Query(config.SQLQuery)
	if queryErr != nil {
		return nil, nil, logz.ErrorLog(fmt.Sprintf("falha ao executar a query SQL: %v", queryErr), "etl", logz.QUIET)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var data []Data
	columns, columnsErr := rows.Columns()
	if columnsErr != nil {
		return nil, nil, logz.ErrorLog(fmt.Sprintf("falha ao obter colunas: %v", columnsErr), "etl", logz.QUIET)
	}

	for rows.Next() {
		rowData := make([]interface{}, len(columns))
		rowPointers := make([]interface{}, len(columns))
		for i := range rowData {
			rowPointers[i] = &rowData[i]
		}

		if scanErr := rows.Scan(rowPointers...); scanErr != nil {
			return nil, nil, logz.ErrorLog(fmt.Sprintf("falha ao escanear os dados da linha: %v", scanErr), "etl", logz.QUIET)
		}

		row := make(Data)
		for i, colName := range columns {
			row[colName] = rowData[i]
		}
		data = append(data, row)
	}

	if config.OutputPath != "" {
		saveDataErr := SaveData(config.OutputPath, data, config.OutputFormat)
		if saveDataErr != nil {
			_ = logz.ErrorLog("Failed to save data: "+saveDataErr.Error(), "etl", logz.QUIET)
		}
	}

	return data, columns, nil
}
func SaveData(filePath string, data []Data, outputFormat string) error {
	if filePath == "" {
		return logz.ErrorLog("caminho do arquivo não informado", "etl", logz.QUIET)
	}

	if outputFormat == "" {
		outputFormat = "json"
	}

	switch outputFormat {
	case "json":
		if saveDataErr := SaveDataToJSON(filePath, data); saveDataErr != nil {
			return logz.ErrorLog("Failed to save data to JSON: "+saveDataErr.Error(), "etl", logz.QUIET)
		}
	case "yaml":
		if saveDataErr := SaveDataToYAML(filePath, data); saveDataErr != nil {
			return logz.ErrorLog("Failed to save data to YAML: "+saveDataErr.Error(), "etl", logz.QUIET)
		}
	case "xml":
		if saveDataErr := SaveDataToXML(filePath, data); saveDataErr != nil {
			return logz.ErrorLog("Failed to save data to XML: "+saveDataErr.Error(), "etl", logz.QUIET)
		}
	default:
		return logz.ErrorLog("formato de saída inválido", "etl", logz.QUIET)
	}

	return nil
}
func SaveDataToXML(filePath string, data []Data) error {
	if filePath == "" {
		return logz.ErrorLog("caminho do arquivo não informado", "etl", logz.QUIET)
	}

	if len(data) == 0 {
		return logz.ErrorLog("dados não informados", "etl", logz.QUIET)
	}

	if ensureFileErr := utils.EnsureFile(filePath, 0644, []string{}); ensureFileErr != nil {
		return logz.ErrorLog("Failed to ensure file: "+ensureFileErr.Error(), "etl", logz.QUIET)
	}

	file, openFileErr := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if openFileErr != nil {
		return logz.ErrorLog("Failed to open file: "+openFileErr.Error(), "etl", logz.QUIET)
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	encoder := xml.NewEncoder(file)

	if encodeErr := encoder.Encode(data); encodeErr != nil {
		return logz.ErrorLog("Failed to encode data: "+encodeErr.Error(), "etl", logz.QUIET)
	}

	return nil
}
func SaveDataToYAML(filePath string, data []Data) error {
	if filePath == "" {
		return logz.ErrorLog("caminho do arquivo não informado", "etl", logz.QUIET)
	}

	if len(data) == 0 {
		return logz.ErrorLog("dados não informados", "etl", logz.QUIET)
	}

	if ensureFileErr := utils.EnsureFile(filePath, 0644, []string{}); ensureFileErr != nil {
		return logz.ErrorLog("Failed to ensure file: "+ensureFileErr.Error(), "etl", logz.QUIET)
	}

	file, openFileErr := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if openFileErr != nil {
		return logz.ErrorLog("Failed to open file: "+openFileErr.Error(), "etl", logz.QUIET)
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	encoder := yaml.NewEncoder(file)

	if encodeErr := encoder.Encode(data); encodeErr != nil {
		return logz.ErrorLog("Failed to encode data: "+encodeErr.Error(), "etl", logz.QUIET)
	}

	return nil
}
func SaveDataToJSON(filePath string, data []Data) error {
	if filePath == "" {
		return logz.ErrorLog("caminho do arquivo não informado", "etl", logz.QUIET)
	}

	if len(data) == 0 {
		return logz.ErrorLog("dados não informados", "etl", logz.QUIET)
	}

	if ensureFileErr := utils.EnsureFile(filePath, 0644, []string{}); ensureFileErr != nil {
		return logz.ErrorLog("Failed to ensure file: "+ensureFileErr.Error(), "etl", logz.QUIET)
	}

	file, openFileErr := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if openFileErr != nil {
		return logz.ErrorLog("Failed to open file: "+openFileErr.Error(), "etl", logz.QUIET)
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	encoder := json.NewEncoder(file)

	if encodeErr := encoder.Encode(data); encodeErr != nil {
		return logz.ErrorLog("Failed to encode data: "+encodeErr.Error(), "etl", logz.QUIET)
	}

	return nil
}
func LoadData(dbSQL *sql.DB, config Config) error {
	var db *sql.DB
	var dbErr error

	if dbSQL == nil {
		db, dbErr = sql.Open(config.DestinationType, config.DestinationConnectionString)
		if dbErr != nil {
			return logz.ErrorLog("Failed to connect to destination database: "+dbErr.Error(), "etl", logz.QUIET)
		}
	} else {
		db = dbSQL
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	var fieldsWithType map[string]string
	var data []Data
	var fieldsErr error

	data, fieldsWithType, fieldsErr = ExtractDataWithTypes(nil, config)
	if fieldsErr != nil {
		return logz.ErrorLog("Failed to extract data: "+fieldsErr.Error(), "etl", logz.QUIET)
	}

	var fieldsDest map[string]string
	var fieldsList []string
	if config.Transformations != nil {
		for _, t := range config.Transformations {
			if t.Type == "" {
				if fieldType, ok := fieldsWithType[t.SourceField]; ok {
					t.Type = fieldType
				} else {
					return logz.ErrorLog("Failed to get field type: "+t.SourceField, "etl", logz.QUIET)
				}
			} else {
				fieldsWithType[t.SourceField] = t.Type
			}
			fieldsDest[t.DestinationField] = t.Type
			fieldsList = append(fieldsList, t.DestinationField)
		}
	}

	fieldsDest = fieldsWithType
	fieldsList = make([]string, 0, len(fieldsDest))
	for field := range fieldsDest {
		fieldsList = append(fieldsList, field)
	}

	if ensureTableExistsWithTypesErr := EnsureTableExistsWithTypes(db, config, fieldsDest); ensureTableExistsWithTypesErr != nil {
		return logz.ErrorLog("Failed to ensure table exists: "+ensureTableExistsWithTypesErr.Error(), "etl", logz.QUIET)
	}

	transformedData, transformedDataErr := ApplyTransformations(data, config.Transformations)
	if transformedDataErr != nil {
		return logz.ErrorLog("Failed to apply transformations: "+transformedDataErr.Error(), "etl", logz.QUIET)
	}

	if config.OutputPath != "" {
		if saveDataErr := SaveData(config.OutputPath, transformedData, config.OutputFormat); saveDataErr != nil {
			return logz.ErrorLog("Failed to save data: "+saveDataErr.Error(), "etl", logz.QUIET)
		}
	}

	tx, txErr := db.Begin()
	if txErr != nil {
		return logz.ErrorLog(fmt.Sprintf("Failed to start transaction: %v", txErr), "etl", logz.QUIET)
	}
	var insertQuery string
	for _, row := range transformedData {
		var columns, values, conlictFallback strings.Builder
		columns.WriteString(fmt.Sprintf("INSERT INTO %s (", config.DestinationTable))
		values.WriteString("VALUES (")
		i := 0
		for col, val := range row {
			if i > 0 {
				columns.WriteString(", ")
				values.WriteString(", ")
			}
			columns.WriteString(col)
			values.WriteString(formatValue(val))
			if config.UpdateKey != "" {
				if i > 0 {
					conlictFallback.WriteString(", ")
				}
				conlictFallback.WriteString(fmt.Sprintf("%s = %s", col, formatValue(val)))
			}
			i++
		}

		// Por hora vou checar só o primeiro campo. Depois implemento o resto da lógica
		var checkQuery strings.Builder
		if config.UpdateKey != "" {
			checkQuery.WriteString(fmt.Sprintf(") ON CONFLICT (%s) DO UPDATE SET %s", config.UpdateKey, conlictFallback.String()))
		} else {
			values.WriteString(")")
			values.WriteString(";")
		}
		columns.WriteString(") ")
		insertQuery = columns.String() + values.String()
		if conlictFallback.Len() > 0 {
			insertQuery += checkQuery.String() + ";"
		} else {
			insertQuery += ";"
		}
		_, err := db.Exec(insertQuery)
		if err != nil {
			_ = tx.Rollback()
			_ = logz.DebugLog(fmt.Sprintf("Failed to execute insert query: %v", insertQuery), "etl", logz.QUIET)
			return logz.ErrorLog("Failed to execute insert query: "+err.Error(), "etl", logz.QUIET)
		}
	}

	if commitErr := tx.Commit(); commitErr != nil {
		_ = logz.DebugLog(fmt.Sprintf("Failed to commit insertion: %v", insertQuery), "etl", logz.QUIET)
		return logz.ErrorLog("Failed to commit transaction: "+commitErr.Error(), "etl", logz.QUIET)
	}

	_ = logz.InfoLog("Dados carregados no banco de destino com sucesso", "etl", logz.QUIET)

	return nil
}
func ExecuteETL(configPath, outputPath, outputFormat string, needCheck bool, checkMethod string) error {
	_ = logz.InfoLog("Iniciando o processo de GETl", "etl", logz.QUIET)

	// Carregar a configuração
	config, loadConfigErr := LoadConfigFile(configPath)
	if loadConfigErr != nil {
		return logz.ErrorLog(fmt.Sprintf("falha ao carregar a configuração: %v", loadConfigErr), "etl", logz.QUIET)
	}

	// Carregar os dados no banco de destino
	if outputPath != "" {
		config.OutputPath = outputPath
	}
	if outputFormat != "" {
		config.OutputFormat = outputFormat
	}

	if needCheck {
		config.NeedCheck = needCheck
		if checkMethod != "" {
			config.CheckMethod = checkMethod
		} else {
			return logz.ErrorLog("método de verificação não informado", "etl", logz.QUIET)
		}
	}

	// Extrair os dados, transformar e carregar no destino
	loadDataErr := LoadData(nil, config)
	if loadDataErr != nil {
		return logz.ErrorLog(fmt.Sprintf("falha ao carregar os dados no destino: %v", loadDataErr), "etl", logz.QUIET)
	}

	_ = logz.InfoLog("Processo de GETl finalizado com sucesso", "etl", logz.QUIET)

	return nil
}
func VacuumDatabase(dbPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("falha ao abrir o banco de dados: %w", err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	_, err = db.Exec("VACUUM")
	if err != nil {
		return fmt.Errorf("falha ao executar VACUUM: %w", err)
	}

	_ = logz.InfoLog("VACUUM executado com sucesso", "etl", logz.QUIET)
	return nil
}
func ExecuteETLJobs() error {
	_ = logz.InfoLog("Iniciando os trabalhos de GETl", "etl", logz.QUIET)

	jobsObj, jobsListErr := GetETLJobs()
	if jobsListErr != nil {
		return logz.ErrorLog(fmt.Sprintf("falha ao buscar os trabalhos de GETl: %v", jobsListErr), "etl", logz.QUIET)
	}

	jobsList := jobsObj.GetJobs()
	for _, job := range jobsList {
		executeErr := ExecuteETL(job.Path(), job.OutputPath(), job.OutputFormat(), job.NeedCheck(), job.CheckMethod())
		if executeErr != nil {
			return logz.ErrorLog(fmt.Sprintf("falha ao executar o trabalho de GETl: %v", executeErr), "etl", logz.QUIET)
		}
	}

	_ = logz.InfoLog("Trabalhos de GETl finalizados com sucesso", "etl", logz.QUIET)

	return nil
}
func formatValue(val interface{}) string {
	if val == nil {
		return "NULL"
	}
	switch v := val.(type) {
	case string:
		return fmt.Sprintf("'%s'", v)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	case time.Time:
		return fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05"))
	default:
	}
	return fmt.Sprintf("'%v'", val)
}
