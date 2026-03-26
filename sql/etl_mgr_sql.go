// Package sql provides SQL-related functions for the ETL process
package sql

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"database/sql"
	"encoding/json"
	"encoding/xml"

	"gopkg.in/yaml.v3"

	"github.com/charmbracelet/lipgloss"
	"github.com/kubex-ecosystem/getl/etypes"
	"github.com/kubex-ecosystem/getl/extr"
	"github.com/kubex-ecosystem/getl/utils"

	gl "github.com/kubex-ecosystem/logz"
	ui "github.com/kubex-ecosystem/xtui/components"

	_ "github.com/denisenkom/go-mssqldb" // Microsoft SQL Server
	_ "github.com/godror/godror"         // Oracle
	_ "github.com/lib/pq"                // PostgreSQL
	_ "github.com/mattn/go-sqlite3"      // SQLite
)

// ShowDataTableFromConfig shows the data table from the source
func ShowDataTableFromConfig(fileConfigPath string, export bool, exportPath string, outputFormat string) error {
	config, err := utils.LoadConfigFile(fileConfigPath)
	if err != nil {
		return fmt.Errorf("falha ao carregar configuração da fonte: %v", err)
	}

	var sqlQuery string
	if config.SQLQuery != "" {
		sqlQuery = config.SQLQuery
	} else {
		fields := []string{"*"} // Ajuste conforme necessário
		sqlQuery, _, err = utils.BuilExtractdQuery(config, fields)
		if err != nil {
			return fmt.Errorf("falha ao construir a consulta SQL: %v", err)
		}
	}

	handler, err := utils.GetDataTableHandlerFromQuery(config.SourceType, config.SourceConnectionString, sqlQuery)
	if err != nil {
		return err
	}

	if export {
		if exportPath == "" {
			return fmt.Errorf("caminho de exportação não fornecido")
		}

		var data []etypes.Data
		for _, row := range handler.Data {
			rowData := make(etypes.Data)
			for i, value := range row {
				rowData[handler.Columns[i]] = value
			}
			data = append(data, rowData)
		}

		exportErr := SaveData(exportPath, data, outputFormat)
		if exportErr != nil {
			return fmt.Errorf("falha ao exportar dados para arquivo: %v", exportErr)
		}
		return nil
	}

	customStyles := map[string]lipgloss.Color{
		"header": lipgloss.Color("#01BE85"),
		"row":    lipgloss.Color("#252"),
	}
	return ui.StartTableScreen(handler, customStyles)
}

// inferTypeFromValue analyzes actual data values to infer better SQL types
func inferTypeFromValue(value interface{}) string {
	if value == nil {
		return "TEXT"
	}

	// Use reflection to get the underlying type
	valueType := reflect.TypeOf(value)
	if valueType.Kind() == reflect.Ptr {
		if reflect.ValueOf(value).IsNil() {
			return "TEXT"
		}
		valueType = valueType.Elem()
	}

	switch valueType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "INTEGER"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "INTEGER"
	case reflect.Float32, reflect.Float64:
		return "REAL"
	case reflect.Bool:
		return "INTEGER"
	case reflect.String:
		// Try to parse as number
		strVal := value.(string)
		if _, err := strconv.ParseInt(strVal, 10, 64); err == nil {
			return "INTEGER"
		}
		if _, err := strconv.ParseFloat(strVal, 64); err == nil {
			return "REAL"
		}
		return "TEXT"
	case reflect.Slice:
		if valueType.Elem().Kind() == reflect.Uint8 {
			return "BLOB"
		}
		return "TEXT"
	default:
		return "TEXT"
	}
}

func normalizeDriverName(driver string) string {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "postgresql", "pg":
		return "postgres"
	case "sqlite":
		return "sqlite3"
	default:
		return strings.ToLower(strings.TrimSpace(driver))
	}
}

func isCSVSource(sourceType string) bool {
	return normalizeDriverName(sourceType) == "csv"
}

func promoteInferredType(current, candidate string) string {
	if candidate == "" {
		return current
	}
	if current == "" {
		return candidate
	}
	if current == candidate {
		return current
	}
	if current == "TEXT" || candidate == "TEXT" {
		return "TEXT"
	}
	if (current == "INTEGER" && candidate == "REAL") || (current == "REAL" && candidate == "INTEGER") {
		return "REAL"
	}
	return "TEXT"
}

func extractCSVDataWithTypes(config etypes.Config) ([]etypes.Data, map[string]string, error) {
	csvPath := strings.TrimSpace(config.SourceConnectionString)
	if csvPath == "" {
		return nil, nil, fmt.Errorf("sourceConnectionString deve apontar para o arquivo CSV")
	}

	table := extr.NewCSVDataTable(nil, csvPath)
	if err := table.LoadFile(); err != nil {
		return nil, nil, err
	}

	data, err := table.ExtractData(nil)
	if err != nil {
		return nil, nil, err
	}

	columnTypes := make(map[string]string)
	for _, header := range table.Headers() {
		columnTypes[header] = "TEXT"
	}

	for _, row := range data {
		for column, value := range row {
			inferredType := inferTypeFromValue(value)
			if inferredType == "TEXT" && strings.TrimSpace(gl.Sprintf("%v", value)) == "" {
				continue
			}
			columnTypes[column] = promoteInferredType(columnTypes[column], inferredType)
		}
	}

	return data, columnTypes, nil
}

func resolveDestinationFieldTypes(sourceTypes map[string]string, transformations []etypes.Transformation) (map[string]string, error) {
	if len(transformations) == 0 {
		fields := make(map[string]string, len(sourceTypes))
		for field, fieldType := range sourceTypes {
			fields[field] = fieldType
		}
		return fields, nil
	}

	fields := make(map[string]string, len(transformations))
	for _, transformation := range transformations {
		sourceFieldType := sourceTypes[transformation.SourceField]
		if sourceFieldType == "" {
			return nil, fmt.Errorf("tipo do campo fonte não encontrado: %s", transformation.SourceField)
		}

		destinationField := transformation.DestinationField
		if destinationField == "" {
			destinationField = transformation.SourceField
		}

		destinationType := transformation.Type
		if destinationType == "" {
			switch strings.ToLower(strings.TrimSpace(transformation.Operation)) {
			case "", "copy", "none":
				destinationType = sourceFieldType
			case "uppercase", "base64":
				destinationType = "TEXT"
			case "toint":
				destinationType = "INTEGER"
			default:
				destinationType = sourceFieldType
			}
		}

		fields[destinationField] = destinationType
	}

	return fields, nil
}

func buildPlaceholder(driver string, index int) string {
	switch normalizeDriverName(driver) {
	case "postgres":
		return gl.Sprintf("$%d", index)
	case "sqlserver", "mssql":
		return gl.Sprintf("@p%d", index)
	case "oracle", "godror":
		return gl.Sprintf(":%d", index)
	default:
		return "?"
	}
}

func buildConflictClause(driver, updateKey string, columns []string) (string, error) {
	if updateKey == "" {
		return "", nil
	}

	switch normalizeDriverName(driver) {
	case "postgres", "sqlite3":
		assignments := make([]string, 0, len(columns))
		for _, column := range columns {
			if column == updateKey {
				continue
			}
			assignments = append(assignments, gl.Sprintf("%s = EXCLUDED.%s", column, column))
		}
		if len(assignments) == 0 {
			return gl.Sprintf(" ON CONFLICT (%s) DO NOTHING", updateKey), nil
		}
		return gl.Sprintf(" ON CONFLICT (%s) DO UPDATE SET %s", updateKey, strings.Join(assignments, ", ")), nil
	default:
		return "", gl.Errorf("UpdateKey ainda não suportado para destino %s", driver)
	}
}

func executeInsertBatch(tx *sql.Tx, config etypes.Config, data []etypes.Data) error {
	for _, row := range data {
		columns := make([]string, 0, len(row))
		for column := range row {
			columns = append(columns, column)
		}
		sort.Strings(columns)

		placeholders := make([]string, 0, len(columns))
		args := make([]interface{}, 0, len(columns))
		for index, column := range columns {
			placeholders = append(placeholders, buildPlaceholder(config.DestinationType, index+1))
			args = append(args, row[column])
		}

		conflictClause, err := buildConflictClause(config.DestinationType, config.UpdateKey, columns)
		if err != nil {
			return err
		}

		insertQuery := gl.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)%s",
			config.DestinationTable,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "),
			conflictClause,
		)

		if _, err := tx.Exec(insertQuery, args...); err != nil {
			return fmt.Errorf("falha ao executar insert: %w", err)
		}
	}

	return nil
}

// ExtractDataWithTypes extracts data from the source with types
func ExtractDataWithTypes(dbSQL *sql.DB, config etypes.Config) ([]etypes.Data, map[string]string, error) {
	config.SourceType = normalizeDriverName(config.SourceType)
	config.DestinationType = normalizeDriverName(config.DestinationType)

	if isCSVSource(config.SourceType) {
		return extractCSVDataWithTypes(config)
	}

	var db *sql.DB
	var dbErr error
	shouldCloseDB := false
	if dbSQL == nil {
		db, dbErr = sql.Open(config.SourceType, config.SourceConnectionString)
		if dbErr != nil {
			gl.Log("error", "Failed to connect to source database: "+dbErr.Error())
			return nil, nil, dbErr
		}
		shouldCloseDB = true
	} else {
		db = dbSQL
	}
	if shouldCloseDB {
		defer func(db *sql.DB) {
			_ = db.Close()
		}(db)
	}

	gl.Log("info", "Starting data extraction")

	var rows *sql.Rows
	var SQLQueryArgs []interface{}
	var rowsErr error
	var buildQueryErr error

	if config.SQLQuery == "" {
		var fields []string
		var transformationsList []etypes.Transformation
		transformationsList = config.Transformations
		for i, t := range transformationsList {
			fields = append(fields, t.SourceField)
			if t.Type == "" {
				transformationsList[i].Type = "string"
			}
		}
		config.SQLQuery, SQLQueryArgs, buildQueryErr = utils.BuilExtractdQuery(config, fields)
		if buildQueryErr != nil {
			gl.Log("error", "Failed to build query: "+buildQueryErr.Error())
			return nil, nil, buildQueryErr
		}
	}

	//logz.DebugLog("Running query: "+config.SQLQuery, map[string]interface{}{})

	if len(SQLQueryArgs) > 0 {
		rows, rowsErr = db.Query(config.SQLQuery, SQLQueryArgs...)
	} else {
		rows, rowsErr = db.Query(config.SQLQuery)
	}

	if rowsErr != nil {
		gl.Log("error", "Failed on query execution: "+rowsErr.Error())
		return nil, nil, rowsErr
	}

	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var data []etypes.Data
	columns, columnsErr := rows.Columns()
	if columnsErr != nil {
		gl.Log("error", "Failed to get columns: "+columnsErr.Error())
		return nil, nil, columnsErr
	}

	columnTypes, columnTypesErr := rows.ColumnTypes()
	if columnTypesErr != nil {
		gl.Log("error", "Failed trying to get column types: "+columnTypesErr.Error())
		return nil, nil, columnTypesErr
	}

	columnTypeMap := make(map[string]string)
	for i, colType := range columnTypes {
		dbTypeName := colType.DatabaseTypeName()

		// If DatabaseTypeName is empty, try to infer from ScanType
		if dbTypeName == "" {
			scanType := colType.ScanType()
			if scanType != nil {
				switch scanType.String() {
				case "string":
					dbTypeName = "TEXT"
				case "int64":
					dbTypeName = "INTEGER"
				case "float64":
					dbTypeName = "REAL"
				case "bool":
					dbTypeName = "INTEGER"
				case "[]uint8", "[]byte":
					dbTypeName = "BLOB"
				default:
					dbTypeName = "TEXT" // Default fallback
				}
			} else {
				dbTypeName = "TEXT" // Ultimate fallback
			}
		}

		columnTypeMap[columns[i]] = dbTypeName
	}

	// Track actual data types found in the first rows to improve type inference
	actualTypes := make(map[string]string)
	rowCount := 0

	for rows.Next() {
		rowData := make([]interface{}, len(columns))
		rowPointers := make([]interface{}, len(columns))
		for i := range rowData {
			rowPointers[i] = &rowData[i]
		}

		if scanErr := rows.Scan(rowPointers...); scanErr != nil {
			gl.Log("error", "Failed to scan row data: "+scanErr.Error())
			return nil, nil, scanErr
		}

		row := make(etypes.Data)
		for i, colName := range columns {
			row[colName] = rowData[i]

			// Improve type detection based on actual data (only for first few rows)
			if rowCount < 3 && columnTypeMap[colName] == "TEXT" {
				if actualType := inferTypeFromValue(rowData[i]); actualType != "TEXT" {
					actualTypes[colName] = actualType
				}
			}
		}
		data = append(data, row)
		rowCount++
	}

	// Update column types based on actual data analysis
	for colName, actualType := range actualTypes {
		if columnTypeMap[colName] == "TEXT" {
			columnTypeMap[colName] = actualType
		}
	}

	return data, columnTypeMap, nil
}

// EnsureTableExistsWithTypes ensures the table exists with the correct types
func EnsureTableExistsWithTypes(db *sql.DB, config etypes.Config, fields map[string]string) error {
	if config.DestinationTable == "" {
		gl.Log("error", "nome da tabela não informado")
		return fmt.Errorf("nome da tabela não informado")
	}

	var createTableQuery string
	var fieldsDest = make(map[string]string)
	createTableQuery = gl.Sprintf("CREATE TABLE IF NOT EXISTS %s (", config.DestinationTable)
	for fieldName, fieldType := range fields {
		typeName := etypes.GetVendorSqlType(
			config.DestinationType,
			fieldType,
		)
		if typeName == "" {
			gl.Errorf("tipo de campo não mapeado: %s", fieldType)
			return fmt.Errorf("tipo de campo não mapeado: %s", fieldType)
		}
		if config.UpdateKey == fieldName {
			createTableQuery += gl.Sprintf("%s %s %s, ", fieldName, typeName, "PRIMARY KEY")
		} else {
			createTableQuery += gl.Sprintf("%s %s, ", fieldName, typeName)
		}
		fieldsDest[fieldName] = typeName
	}
	createTableQuery = createTableQuery[:len(createTableQuery)-2] + ")"

	//logz.DebugLog("Campos de destino: "+config.DestinationType+" - "+fmt.Sprintf("%v", fieldsDest), map[string]interface{}{})

	_, createTableQueryErr := db.Exec(createTableQuery)
	if createTableQueryErr != nil {
		gl.Errorf("falha ao criar a tabela: %v", createTableQueryErr)
		return createTableQueryErr
	}

	return nil
}

// ExtractData extracts data from the source
func ExtractData(dbSQL *sql.DB, config etypes.Config) ([]etypes.Data, []string, error) {
	config.SourceType = normalizeDriverName(config.SourceType)
	if config.SQLQuery == "" {
		gl.Log("error", "query SQL não informada")
		return nil, nil, fmt.Errorf("query SQL não informada")
	}

	var db *sql.DB
	var dbErr error
	shouldCloseDB := false
	if dbSQL == nil {
		db, dbErr = sql.Open(config.SourceType, config.SourceConnectionString)
		if dbErr != nil {
			gl.Errorf("falha ao conectar ao banco de dados: %v", dbErr)
			return nil, nil, dbErr
		}
		shouldCloseDB = true
	} else {
		db = dbSQL
	}
	if shouldCloseDB {
		defer func(db *sql.DB) {
			_ = db.Close()
		}(db)
	}

	rows, queryErr := db.Query(config.SQLQuery)
	if queryErr != nil {
		gl.Errorf("falha ao executar a query SQL: %v", queryErr)
		return nil, nil, queryErr
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var data []etypes.Data
	columns, columnsErr := rows.Columns()
	if columnsErr != nil {
		gl.Errorf("falha ao obter colunas: %v", columnsErr)
		return nil, nil, columnsErr
	}

	for rows.Next() {
		rowData := make([]interface{}, len(columns))
		rowPointers := make([]interface{}, len(columns))
		for i := range rowData {
			rowPointers[i] = &rowData[i]
		}

		if scanErr := rows.Scan(rowPointers...); scanErr != nil {
			gl.Errorf("falha ao escanear os dados da linha: %v", scanErr)
			return nil, nil, scanErr
		}

		row := make(etypes.Data)
		for i, colName := range columns {
			row[colName] = rowData[i]
		}
		data = append(data, row)
	}

	if config.OutputPath != "" {
		saveDataErr := SaveData(config.OutputPath, data, config.OutputFormat)
		if saveDataErr != nil {
			gl.Log("error", "Failed to save data: "+saveDataErr.Error())
		}
	}

	return data, columns, nil
}

// SaveData saves data to a file
func SaveData(filePath string, data []etypes.Data, outputFormat string) error {
	if filePath == "" {
		gl.Log("error", "caminho do arquivo não informado")
		return fmt.Errorf("caminho do arquivo não informado")
	}

	if outputFormat == "" {
		outputFormat = "json"
	}

	switch outputFormat {
	case "json":
		if saveDataErr := SaveDataToJSON(filePath, data); saveDataErr != nil {
			gl.Log("error", "Failed to save data to JSON: "+saveDataErr.Error())
			return fmt.Errorf("Failed to save data to JSON: %v", saveDataErr)
		}
	case "yaml":
		if saveDataErr := SaveDataToYAML(filePath, data); saveDataErr != nil {
			gl.Log("error", "Failed to save data to YAML: "+saveDataErr.Error())
			return fmt.Errorf("Failed to save data to YAML: %v", saveDataErr)
		}
	case "xml":
		if saveDataErr := SaveDataToXML(filePath, data); saveDataErr != nil {
			gl.Log("error", "Failed to save data to XML: "+saveDataErr.Error())
			return fmt.Errorf("Failed to save data to XML: %v", saveDataErr)
		}
	default:
		gl.Log("error", "formato de saída inválido")
		return fmt.Errorf("formato de saída inválido")
	}

	return nil
}

// XMLData represents a wrapper for XML serialization
type XMLData struct {
	XMLName xml.Name    `xml:"data"`
	Records []XMLRecord `xml:"record"`
}

// XMLRecord represents a record in the XML file
type XMLRecord struct {
	XMLName xml.Name   `xml:"record"`
	Fields  []XMLField `xml:"field"`
}

// XMLField represents a field in the XML file
type XMLField struct {
	XMLName xml.Name `xml:"field"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",chardata"`
}

// SaveDataToXML saves data to an XML file
func SaveDataToXML(filePath string, data []etypes.Data) error {
	if filePath == "" {
		gl.Log("error", "caminho do arquivo não informado")
		return fmt.Errorf("caminho do arquivo não informado")
	}

	if len(data) == 0 {
		gl.Log("error", "dados não informados")
		return fmt.Errorf("dados não informados")
	}

	if ensureDirErr := os.MkdirAll(filepath.Dir(filePath), 0644); ensureDirErr != nil {
		gl.Log("error", "Failed to ensure file: "+ensureDirErr.Error())
		return fmt.Errorf("Failed to ensure file: %v", ensureDirErr)
	}

	if ensureFileErr := os.WriteFile(filePath, []byte{}, 0644); ensureFileErr != nil {
		gl.Log("error", "Failed to ensure file: "+ensureFileErr.Error())
		return fmt.Errorf("Failed to ensure file: %v", ensureFileErr)
	}

	file, openFileErr := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if openFileErr != nil {
		gl.Log("error", "Failed to open file: "+openFileErr.Error())
		return fmt.Errorf("Failed to open file: %v", openFileErr)
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Convert Data to XML-serializable format
	xmlData := XMLData{
		Records: make([]XMLRecord, len(data)),
	}

	for i, record := range data {
		xmlRecord := XMLRecord{
			Fields: make([]XMLField, 0, len(record)),
		}

		for key, value := range record {
			field := XMLField{
				Name:  key,
				Value: formatValue(value),
			}
			// Remove quotes from formatValue for XML content
			if len(field.Value) >= 2 && field.Value[0] == '\'' && field.Value[len(field.Value)-1] == '\'' {
				field.Value = field.Value[1 : len(field.Value)-1]
			}
			if field.Value == "NULL" {
				field.Value = ""
			}
			xmlRecord.Fields = append(xmlRecord.Fields, field)
		}
		xmlData.Records[i] = xmlRecord
	}

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")

	if encodeErr := encoder.Encode(xmlData); encodeErr != nil {
		gl.Log("error", "Failed to encode data: "+encodeErr.Error())
		return fmt.Errorf("Failed to encode data: %v", encodeErr)
	}

	return nil
}

// SaveDataToYAML saves data to a YAML file
func SaveDataToYAML(filePath string, data []etypes.Data) error {
	if filePath == "" {
		gl.Log("error", "caminho do arquivo não informado")
		return fmt.Errorf("caminho do arquivo não informado")
	}

	if len(data) == 0 {
		gl.Log("error", "dados não informados")
		return fmt.Errorf("dados não informados")
	}

	if ensureDirErr := os.MkdirAll(filepath.Dir(filePath), 0644); ensureDirErr != nil {
		gl.Log("error", "Failed to ensure file: "+ensureDirErr.Error())
		return fmt.Errorf("Failed to ensure file: %v", ensureDirErr)
	}

	if ensureFileErr := os.WriteFile(filePath, []byte{}, 0644); ensureFileErr != nil {
		gl.Log("error", "Failed to ensure file: "+ensureFileErr.Error())
		return fmt.Errorf("Failed to ensure file: %v", ensureFileErr)
	}

	file, openFileErr := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if openFileErr != nil {
		gl.Log("error", "Failed to open file: "+openFileErr.Error())
		return openFileErr
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	encoder := yaml.NewEncoder(file)

	if encodeErr := encoder.Encode(data); encodeErr != nil {
		gl.Log("error", "Failed to encode data: "+encodeErr.Error())
		return encodeErr
	}

	return nil
}

// SaveDataToJSON saves data to a JSON file
func SaveDataToJSON(filePath string, data []etypes.Data) error {
	if filePath == "" {
		gl.Log("error", "caminho do arquivo não informado")
		return fmt.Errorf("caminho do arquivo não informado")
	}

	if len(data) == 0 {
		gl.Log("error", "dados não informados")
		return fmt.Errorf("dados não informados")
	}

	if ensureDirErr := os.MkdirAll(filepath.Dir(filePath), 0644); ensureDirErr != nil {
		gl.Log("error", "Failed to ensure file: "+ensureDirErr.Error())
		return fmt.Errorf("Failed to ensure file: %v", ensureDirErr)
	}

	if ensureFileErr := os.WriteFile(filePath, []byte{}, 0644); ensureFileErr != nil {
		gl.Log("error", "Failed to ensure file: "+ensureFileErr.Error())
		return fmt.Errorf("Failed to ensure file: %v", ensureFileErr)
	}

	file, openFileErr := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if openFileErr != nil {
		gl.Log("error", "Failed to open file: "+openFileErr.Error())
		return fmt.Errorf("Failed to open file: %v", openFileErr)
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	encoder := json.NewEncoder(file)

	if encodeErr := encoder.Encode(data); encodeErr != nil {
		gl.Log("error", "Failed to encode data: "+encodeErr.Error())
		return fmt.Errorf("Failed to encode data: %v", encodeErr)
	}

	return nil
}

// LoadData loads data from a file to the destination database
func LoadData(dbSQL *sql.DB, config etypes.Config) error {
	config.SourceType = normalizeDriverName(config.SourceType)
	config.DestinationType = normalizeDriverName(config.DestinationType)

	var db *sql.DB
	var dbErr error
	shouldCloseDB := false

	if dbSQL == nil {
		db, dbErr = sql.Open(config.DestinationType, config.DestinationConnectionString)
		if dbErr != nil {
			gl.Log("error", "Failed to connect to destination database: "+dbErr.Error())
			return dbErr
		}
		shouldCloseDB = true
	} else {
		db = dbSQL
	}
	if shouldCloseDB {
		defer func(db *sql.DB) {
			_ = db.Close()
		}(db)
	}

	var data []etypes.Data
	var fieldsErr error

	data, sourceFieldTypes, fieldsErr := ExtractDataWithTypes(nil, config)
	if fieldsErr != nil {
		gl.Log("error", "Failed to extract data: "+fieldsErr.Error())
		return fieldsErr
	}

	transformedData, transformedDataErr := utils.ApplyTransformations(data, config.Transformations)
	if transformedDataErr != nil {
		gl.Log("error", "Failed to apply transformations: "+transformedDataErr.Error())
		return transformedDataErr
	}

	destinationFieldTypes, fieldTypeErr := resolveDestinationFieldTypes(sourceFieldTypes, config.Transformations)
	if fieldTypeErr != nil {
		gl.Log("error", "Failed to resolve destination field types: "+fieldTypeErr.Error())
		return fieldTypeErr
	}

	if ensureTableExistsWithTypesErr := EnsureTableExistsWithTypes(db, config, destinationFieldTypes); ensureTableExistsWithTypesErr != nil {
		gl.Log("error", "Failed to ensure table exists: "+ensureTableExistsWithTypesErr.Error())
		return ensureTableExistsWithTypesErr
	}

	if config.OutputPath != "" {
		if saveDataErr := SaveData(config.OutputPath, transformedData, config.OutputFormat); saveDataErr != nil {
			gl.Log("error", "Failed to save data: "+saveDataErr.Error())
			return saveDataErr
		}
	}

	tx, txErr := db.Begin()
	if txErr != nil {
		gl.Errorf("Failed to start transaction: %v", txErr)
		return fmt.Errorf("Failed to start transaction: %v", txErr)
	}
	if err := executeInsertBatch(tx, config, transformedData); err != nil {
		_ = tx.Rollback()
		gl.Log("error", "Failed to execute insert query: "+err.Error())
		return err
	}

	if commitErr := tx.Commit(); commitErr != nil {
		gl.Log("error", "Failed to commit transaction: "+commitErr.Error())
		return fmt.Errorf("Failed to commit transaction: %v", commitErr)
	}

	gl.Log("info", "Dados carregados no banco de destino com sucesso")

	return nil
}

// ExecuteETL executes the ETL process
func ExecuteETL(configPath, outputPath, outputFormat string, needCheck bool, checkMethod string) error {
	gl.Log("info", "Iniciando o processo de GETl")

	// Carregar a configuração
	config, loadConfigErr := utils.LoadConfigFile(configPath)
	if loadConfigErr != nil {
		gl.Errorf("falha ao carregar a configuração: %v", loadConfigErr)
		return loadConfigErr
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
			gl.Log("error", "método de verificação não informado")
			return fmt.Errorf("método de verificação não informado")
		}
	}

	// Check if incremental sync is enabled
	if config.IncrementalSync.Enabled {
		return ExecuteIncrementalETL(config)
	}

	// Extrair os dados, transformar e carregar no destino
	loadDataErr := LoadData(nil, config)
	if loadDataErr != nil {
		gl.Errorf("falha ao carregar os dados no destino: %v", loadDataErr)
		return loadDataErr
	}

	gl.Log("info", "Processo de GETl finalizado com sucesso")

	return nil
}

// ExecuteIncrementalETL performs incremental ETL using smart strategies
func ExecuteIncrementalETL(config etypes.Config) error {
	gl.Log("info", "Iniciando processo de GETl incremental")

	// Set default state file if not provided
	if config.IncrementalSync.StateFile == "" {
		config.IncrementalSync.StateFile = gl.Sprintf("/tmp/getl-state-%s-%s.json",
			config.SourceTable, config.DestinationTable)
	}

	// Execute based on strategy
	switch config.IncrementalSync.Strategy {
	case etypes.TimestampBased:
		return executeTimestampIncrementalETL(config)
	case etypes.PrimaryKeyBased:
		return executePrimaryKeyIncrementalETL(config)
	default:
		gl.Log("info", "Unknown incremental strategy, falling back to full sync")
		return LoadData(nil, config)
	}
}

// executeTimestampIncrementalETL performs timestamp-based incremental sync
func executeTimestampIncrementalETL(config etypes.Config) error {
	gl.Infof("Executing timestamp-based incremental sync on field: %s", config.IncrementalSync.TimestampField)

	// Load last sync state
	lastSyncValue, err := loadLastSyncValue(config.IncrementalSync.StateFile)
	if err != nil {
		gl.Log("info", "No previous sync state found, starting full sync")
		lastSyncValue = nil
	}

	// Modify the SQL query to include timestamp filter
	originalQuery := config.SQLQuery
	if originalQuery == "" {
		originalQuery = gl.Sprintf("SELECT * FROM %s", config.SourceTable)
	}

	if lastSyncValue != nil {
		whereClause := gl.Sprintf("%s > '%v'", config.IncrementalSync.TimestampField, lastSyncValue)
		if strings.Contains(strings.ToUpper(originalQuery), "WHERE") {
			config.SQLQuery = originalQuery + " AND " + whereClause
		} else {
			config.SQLQuery = originalQuery + " WHERE " + whereClause
		}
		gl.Infof("Resuming from last sync: %v", lastSyncValue)
	} else {
		config.SQLQuery = originalQuery
		gl.Log("info", "First time sync - processing all records")
	}

	// Add ORDER BY to ensure consistent results
	if !strings.Contains(strings.ToUpper(config.SQLQuery), "ORDER BY") {
		config.SQLQuery += gl.Sprintf(" ORDER BY %s", config.IncrementalSync.TimestampField)
	}

	gl.Infof("Incremental query: %s", config.SQLQuery)

	// Execute the ETL with modified query
	loadDataErr := LoadData(nil, config)
	if loadDataErr != nil {
		return loadDataErr
	}

	// Update sync state with current timestamp
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	saveErr := saveLastSyncValue(config.IncrementalSync.StateFile, currentTime)
	if saveErr != nil {
		gl.Errorf("Failed to save sync state: %v", saveErr)
	} else {
		gl.Infof("Saved sync state: %s", currentTime)
	}

	gl.Log("info", "Timestamp-based incremental sync completed successfully")
	return nil
}

// executePrimaryKeyIncrementalETL performs primary key-based incremental sync
func executePrimaryKeyIncrementalETL(config etypes.Config) error {
	gl.Infof("Executing primary key-based incremental sync on field: %s", config.PrimaryKey)

	// Load last sync state
	lastSyncValue, err := loadLastSyncValue(config.IncrementalSync.StateFile)
	if err != nil {
		gl.Log("info", "No previous sync state found, starting full sync")
		lastSyncValue = nil
	}

	// Modify the SQL query to include primary key filter
	originalQuery := config.SQLQuery
	if originalQuery == "" {
		originalQuery = gl.Sprintf("SELECT * FROM %s", config.SourceTable)
	}

	if lastSyncValue != nil {
		whereClause := gl.Sprintf("%s > %v", config.PrimaryKey, lastSyncValue)
		if strings.Contains(strings.ToUpper(originalQuery), "WHERE") {
			config.SQLQuery = originalQuery + " AND " + whereClause
		} else {
			config.SQLQuery = originalQuery + " WHERE " + whereClause
		}
		gl.Infof("Resuming from last primary key: %v", lastSyncValue)
	} else {
		config.SQLQuery = originalQuery
		gl.Log("info", "First time sync - processing all records")
	}

	// Add ORDER BY to ensure consistent results
	if !strings.Contains(strings.ToUpper(config.SQLQuery), "ORDER BY") {
		config.SQLQuery += gl.Sprintf(" ORDER BY %s", config.PrimaryKey)
	}

	gl.Infof("Incremental query: %s", config.SQLQuery)

	// Execute the ETL with modified query
	loadDataErr := LoadData(nil, config)
	if loadDataErr != nil {
		return loadDataErr
	}

	// For primary key sync, we'll use a simple increment as placeholder
	// In a real implementation, we'd query for the actual max value
	newSyncValue := 1
	if lastSyncValue != nil {
		if val, ok := lastSyncValue.(float64); ok {
			newSyncValue = int(val) + 100 // Increment by batch size
		}
	}

	saveErr := saveLastSyncValue(config.IncrementalSync.StateFile, newSyncValue)
	if saveErr != nil {
		gl.Errorf("Failed to save sync state: %v", saveErr)
	} else {
		gl.Infof("Saved sync state: %v", newSyncValue)
	}

	gl.Log("info", "Primary key-based incremental sync completed successfully")
	return nil
}

// Helper functions for state management
func loadLastSyncValue(stateFile string) (interface{}, error) {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, err
	}

	var state etypes.SyncState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return state.LastSyncValue, nil
}

func saveLastSyncValue(stateFile string, value interface{}) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	state := etypes.SyncState{
		LastSyncValue: value,
		LastSyncTime:  time.Now().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(stateFile, data, 0644)
}

// VacuumDatabase performs vacuum on the database
func VacuumDatabase(dbPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("falha ao abrir o banco de dados: %v", err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	_, err = db.Exec("VACUUM")
	if err != nil {
		return fmt.Errorf("falha ao executar VACUUM: %v", err)
	}

	gl.Log("info", "VACUUM executado com sucesso")
	return nil
}

// ExecuteETLJobs executes all ETL jobs
func ExecuteETLJobs() error {
	gl.Log("info", "Iniciando os trabalhos de GETl")

	jobsObj, jobsListErr := utils.GetETLJobs()
	if jobsListErr != nil {
		gl.Errorf("falha ao buscar os trabalhos de GETl: %v", jobsListErr)
		return jobsListErr
	}

	jobsList := jobsObj.GetJobs()
	for _, job := range jobsList {
		executeErr := ExecuteETL(job.Path(), job.OutputPath(), job.OutputFormat(), job.NeedCheck(), job.CheckMethod())
		if executeErr != nil {
			gl.Errorf("falha ao executar o trabalho de GETl: %v", executeErr)
			return executeErr
		}
	}

	gl.Log("info", "Trabalhos de GETl finalizados com sucesso")

	return nil
}

// formatValue formats a value for SQL
func formatValue(val interface{}) string {
	if val == nil {
		return "NULL"
	}
	switch v := val.(type) {
	case string:
		return gl.Sprintf("'%s'", v)
	case int, int8, int16, int32, int64:
		return gl.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return gl.Sprintf("%d", v)
	case float32, float64:
		return gl.Sprintf("%f", v)
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	case time.Time:
		return gl.Sprintf("'%s'", v.Format("2006-01-02 15:04:05"))
	default:
	}
	return gl.Sprintf("'%v'", val)
}
