package sql

import (
	"database/sql"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	. "github.com/faelmori/kbx/mods/getl/etypes"
	. "github.com/faelmori/kbx/mods/getl/utils"
	"github.com/faelmori/kbx/mods/logz"
	ui "github.com/faelmori/kbx/mods/ui/components"
	_ "github.com/godror/godror"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"strings"
)

// ShowDataTableFromConfig exibe uma tabela de dados com base na configuração fornecida.
// fileConfigPath: o caminho do arquivo de configuração.
// Retorna um erro, se houver.
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

// ExtractDataWithTypes extrai dados de um banco de dados com tipos de coluna.
// dbSQL: conexão de banco de dados existente (pode ser nil).
// config: configuração do ETL.
// Retorna os dados extraídos, um mapa de tipos de coluna e um erro, se houver.
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
	defer db.Close()

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

	defer rows.Close()

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

// ensureTableExistsWithTypes garante que uma tabela exista com os tipos de coluna especificados.
// db: conexão de banco de dados.
// tableName: nome da tabela.
// fields: mapa de campos e seus tipos.
// dbVendor: fornecedor do banco de dados.
// Retorna um erro, se houver.
func EnsureTableExistsWithTypes(db *sql.DB, tableName string, fields map[string]string, dbVendor string) error {
	if tableName == "" {
		return logz.ErrorLog("nome da tabela não informado", "etl", logz.QUIET)
	}

	var createTableQuery string
	var fieldsDest = make(map[string]string)
	createTableQuery = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", tableName)
	for fieldName, fieldType := range fields {
		typeName := GetVendorSqlType(dbVendor, fieldType)
		if typeName == "" {
			return logz.ErrorLog(fmt.Sprintf("tipo de campo não mapeado: %s", fieldType), "etl", logz.QUIET)
		}
		createTableQuery += fmt.Sprintf("%s %s, ", fieldName, typeName)
		fieldsDest[fieldName] = typeName
	}
	createTableQuery = createTableQuery[:len(createTableQuery)-2] + ")"

	_ = logz.DebugLog("Campos de destino: "+dbVendor+" - "+fmt.Sprintf("%v", fieldsDest), "etl", logz.QUIET)

	_, createTableQueryErr := db.Exec(createTableQuery)
	if createTableQueryErr != nil {
		return logz.ErrorLog(fmt.Sprintf("falha ao criar a tabela: %v", createTableQueryErr), "etl", logz.QUIET)
	}

	return nil
}

// extractData extrai dados de um banco de dados.
// dbSQL: conexão de banco de dados existente (pode ser nil).
// config: configuração do ETL.
// Retorna os dados extraídos, uma lista de colunas e um erro, se houver.
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
	defer db.Close()

	rows, queryErr := db.Query(config.SQLQuery)
	if queryErr != nil {
		return nil, nil, logz.ErrorLog(fmt.Sprintf("falha ao executar a query SQL: %v", queryErr), "etl", logz.QUIET)
	}
	defer rows.Close()

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

// saveData salva os dados extraídos em um arquivo JSON.
// filePath: caminho do arquivo.
// data: dados a serem salvos.
// Retorna um erro, se houver.
func SaveData(filePath string, data []Data, outputFormat string) error {

	return nil
}

// loadData carrega dados no banco de dados de destino.
// dbSQL: conexão de banco de dados existente (pode ser nil).
// config: configuração do ETL.
// Retorna um erro, se houver.
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
	defer db.Close()

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

	if ensureTableExistsWithTypesErr := EnsureTableExistsWithTypes(db, config.DestinationTable, fieldsDest, config.DestinationType); ensureTableExistsWithTypesErr != nil {
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

	for _, row := range transformedData {
		var columns, values strings.Builder
		columns.WriteString(fmt.Sprintf("INSERT INTO %s (", config.DestinationTable))
		values.WriteString("VALUES (")

		i := 0
		for col, val := range row {
			if i > 0 {
				columns.WriteString(", ")
				values.WriteString(", ")
			}
			columns.WriteString(col)
			values.WriteString(fmt.Sprintf("'%v'", val))
			i++
		}

		columns.WriteString(") ")
		values.WriteString(")")

		values.WriteString(";")

		insertQuery := columns.String() + values.String()

		_, err := db.Exec(insertQuery)
		if err != nil {
			_ = tx.Rollback()
			return logz.ErrorLog("Failed to execute insert query: "+err.Error(), "etl", logz.QUIET)
		}
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return logz.ErrorLog("Failed to commit transaction: "+commitErr.Error(), "etl", logz.QUIET)
	}

	_ = logz.InfoLog("Dados carregados no banco de destino com sucesso", "etl", logz.QUIET)

	return nil
}

// ExecuteETL executa o processo de ETL.
// configPath: caminho do arquivo de configuração.
// outputPath: caminho do arquivo de saída (opcional).
// needCheck: indica se é necessário verificar alterações.
// checkMethod: método de verificação (opcional).
// Retorna um erro, se houver.
func ExecuteETL(configPath string, outputPath string, needCheck bool, checkMethod string) error {
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

// vacuumDatabase executa o comando VACUUM no banco de dados SQLite.
// dbPath: o caminho do banco de dados.
// Retorna um erro, se houver.
func VacuumDatabase(dbPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("falha ao abrir o banco de dados: %w", err)
	}
	defer db.Close()

	_, err = db.Exec("VACUUM")
	if err != nil {
		return fmt.Errorf("falha ao executar VACUUM: %w", err)
	}

	_ = logz.InfoLog("VACUUM executado com sucesso", "etl", logz.QUIET)
	return nil
}

// ExecuteETLJobs executa os trabalhos de ETL.
// Retorna um erro, se houver.
func ExecuteETLJobs() error {
	_ = logz.InfoLog("Iniciando os trabalhos de GETl", "etl", logz.QUIET)

	jobsObj, jobsListErr := GetETLJobs()
	if jobsListErr != nil {
		return logz.ErrorLog(fmt.Sprintf("falha ao buscar os trabalhos de GETl: %v", jobsListErr), "etl", logz.QUIET)
	}

	jobsList := jobsObj.GetJobs()
	for _, job := range jobsList {
		executeErr := ExecuteETL(job.Path(), job.OutputPath(), job.NeedCheck(), job.CheckMethod())
		if executeErr != nil {
			return logz.ErrorLog(fmt.Sprintf("falha ao executar o trabalho de GETl: %v", executeErr), "etl", logz.QUIET)
		}
	}

	_ = logz.InfoLog("Trabalhos de GETl finalizados com sucesso", "etl", logz.QUIET)

	return nil
}
