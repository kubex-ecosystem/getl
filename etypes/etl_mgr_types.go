package etypes

import (
	"fmt"

	gl "github.com/kubex-ecosystem/getl/internal/module/logger"
)

const batchSize = 1000

type TableHandler struct {
	Columns []string
	Data    [][]string
}

func (h *TableHandler) GetHeaders() []string { return h.Columns }
func (h *TableHandler) GetRows() [][]string  { return h.Data }
func (h *TableHandler) GetArrayMap() map[string][]string {
	var result = make(map[string][]string)
	for _, column := range h.Columns {
		var columnData []string
		for _, row := range h.Data {
			columnData = append(columnData, row[0])
		}
		result[column] = columnData
	}
	return result
}

func (h *TableHandler) GetHashMap() map[string]string {
	var result = make(map[string]string)
	for _, column := range h.Columns {
		result[column] = h.Data[0][0]
	}
	return result
}
func (h *TableHandler) GetObjectMap() []map[string]string {
	var result []map[string]string
	for _, row := range h.Data {
		rowMap := make(map[string]string)
		for i, column := range h.Columns {
			rowMap[column] = row[i]
		}
		result = append(result, rowMap)
	}
	return result
}
func (h *TableHandler) GetByteMap() map[string][]byte {
	var result = make(map[string][]byte)
	for _, column := range h.Columns {
		result[column] = []byte(h.Data[0][0])
	}
	return result
}

type Field map[string]string
type Fields map[string][]Field
type Data map[string]interface{}
type Config struct {
	SourceType                  string           `json:"sourceType"`
	SourceConnectionString      string           `json:"sourceConnectionString"`
	SourceTable                 string           `json:"sourceTable"`
	DestinationType             string           `json:"destinationType"`
	DestinationConnectionString string           `json:"destinationConnectionString"`
	DestinationTable            string           `json:"destinationTable"`
	SQLQuery                    string           `json:"sqlQuery"`
	OutputPath                  string           `json:"outputPath"`
	OutputFormat                string           `json:"outputFormat"`
	Transformations             []Transformation `json:"transformations"`
	NeedCheck                   bool             `json:"needCheck"`
	CheckMethod                 string           `json:"checkMethod"`
	Joins                       []Join           `json:"joins"`
	Where                       string           `json:"where"`
	OrderBy                     string           `json:"orderBy"`
	Triggers                    []Trigger        `json:"triggers"`
	LogTable                    string           `json:"logTable"`
	SyncInterval                string           `json:"syncInterval"`
	KafkaURL                    string           `json:"kafkaURL"`
	KafkaTopic                  string           `json:"kafkaTopic"`
	KafkaGroupID                string           `json:"kafkaGroupID"`
	PrimaryKey                  string           `json:"primaryKey"`
	UpdateKey                   string           `json:"updateKey"`

	// Incremental Sync Configuration
	IncrementalSync IncrementalSyncConfig `json:"incrementalSync"`
}

// IncrementalSyncConfig defines the configuration for incremental synchronization
type IncrementalSyncConfig struct {
	Enabled        bool                    `json:"enabled"`
	Strategy       IncrementalSyncStrategy `json:"strategy"`
	TimestampField string                  `json:"timestampField,omitempty"`
	LastSyncValue  interface{}             `json:"lastSyncValue,omitempty"`
	StateFile      string                  `json:"stateFile,omitempty"`
	BatchSize      int                     `json:"batchSize,omitempty"`
}

// IncrementalSyncStrategy defines the type of incremental sync strategy
type IncrementalSyncStrategy string

const (
	// TimestampBased uses a timestamp column to detect changes
	TimestampBased IncrementalSyncStrategy = "timestamp"
	// PrimaryKeyBased uses primary key ranges for insert-only tables
	PrimaryKeyBased IncrementalSyncStrategy = "primary_key"
	// HashBased uses row hashing to detect any changes
	HashBased IncrementalSyncStrategy = "hash"
	// FullSync performs a complete synchronization (default)
	FullSync IncrementalSyncStrategy = "full"
)

// SyncState tracks the state of incremental synchronization
type SyncState struct {
	SourceTable      string      `json:"sourceTable"`
	DestinationTable string      `json:"destinationTable"`
	Strategy         string      `json:"strategy"`
	LastSyncValue    interface{} `json:"lastSyncValue"`
	LastSyncTime     string      `json:"lastSyncTime"`
	RecordsProcessed int64       `json:"recordsProcessed"`
	TotalRecords     int64       `json:"totalRecords"`
}

type Transformation struct {
	SourceField      string `json:"sourceField"`
	DestinationField string `json:"destinationField"`
	Operation        string `json:"operation"`
	SPath            string `json:"sPath"`
	DPath            string `json:"dPath"`
	Type             string `json:"type"`
}
type Join struct {
	Table     string `json:"table"`
	Condition string `json:"condition"`
	JoinType  string `json:"joinType"`
}
type Trigger struct {
	Name      string `json:"name"`
	Table     string `json:"table"`
	Event     string `json:"event"`
	Statement string `json:"statement"`
}
type VendorSqlTypeMap struct {
	sourceType string
	targetType string
	fallback   string
}
type VendorSqlTypeMapList []VendorSqlTypeMap
type VendorSqlMapping struct {
	driver  string
	mapping VendorSqlTypeMapList
}
type VendorSqlMappingList []VendorSqlMapping

var vAendorMappingList = VendorSqlMappingList{
	{
		driver: "sqlite3",
		mapping: VendorSqlTypeMapList{
			{"NUMBER", "REAL", "REAL"},
			{"VARCHAR", "TEXT", "TEXT"},
			{"TEXT", "TEXT", "TEXT"},
			{"INT", "INTEGER", "INTEGER"},
			{"DECIMAL", "REAL", "REAL"},
			{"VARCHAR2", "TEXT", "TEXT"},
			{"DATE", "TEXT", "TEXT"},
			{"DATETIME", "TEXT", "TEXT"},
			{"TIMESTAMP", "TEXT", "TEXT"},
			{"BOOLEAN", "INTEGER", "INTEGER"},
			{"BLOB", "BLOB", "BLOB"},
			{"CLOB", "CLOB", "CLOB"},
		},
	},
	{
		driver: "sqlite",
		mapping: VendorSqlTypeMapList{
			{"NUMBER", "REAL", "REAL"},
			{"VARCHAR", "TEXT", "TEXT"},
			{"TEXT", "TEXT", "TEXT"},
			{"INT", "REAL", "REAL"},
			{"DECIMAL", "REAL", "REAL"},
			{"VARCHAR2", "TEXT", "TEXT"},
			{"DATE", "TEXT", "TEXT"},
			{"DATETIME", "TEXT", "TEXT"},
			{"TIMESTAMP", "TEXT", "TEXT"},
			{"BOOLEAN", "INTEGER", "INTEGER"},
			{"BLOB", "BLOB", "BLOB"},
			{"CLOB", "CLOB", "CLOB"},
			{"REAL", "REAL", "REAL"},
			{"FLOAT", "REAL", "REAL"},
		},
	},
	{
		driver: "postgres",
		mapping: VendorSqlTypeMapList{
			{"NUMBER", "NUMERIC", "NUMERIC"},
			{"VARCHAR", "VARCHAR", "VARCHAR"},
			{"NVARCHAR", "VARCHAR", "VARCHAR"},
			{"CHAR", "TEXT", "TEXT"},
			{"TEXT", "TEXT", "TEXT"},
			{"INT", "INTEGER", "INTEGER"},
			{"SMALLINT", "INTEGER", "INTEGER"},
			{"DECIMAL", "NUMERIC", "NUMERIC"},
			{"VARCHAR2", "VARCHAR", "VARCHAR"},
			{"DATE", "DATE", "DATE"},
			{"DATETIME", "TIMESTAMP", "TIMESTAMP"},
			{"TIMESTAMP", "TIMESTAMP", "TIMESTAMP"},
			{"BOOLEAN", "BOOLEAN", "BOOLEAN"},
			{"BLOB", "BYTEA", "BYTEA"},
			{"CLOB", "TEXT", "TEXT"},
			{"REAL", "REAL", "REAL"},
			{"FLOAT", "REAL", "REAL"},
		},
	},
	{
		driver: "mysql",
		mapping: VendorSqlTypeMapList{
			{"NUMBER", "INT", "INT"},
			{"VARCHAR", "VARCHAR", "VARCHAR"},
			{"TEXT", "TEXT", "TEXT"},
			{"INT", "INT", "INT"},
			{"DECIMAL", "DECIMAL", "DECIMAL"},
			{"VARCHAR2", "VARCHAR", "VARCHAR"},
			{"DATE", "DATE", "DATE"},
			{"DATETIME", "DATETIME", "DATETIME"},
			{"TIMESTAMP", "TIMESTAMP", "TIMESTAMP"},
			{"BOOLEAN", "TINYINT", "TINYINT"},
			{"BLOB", "BLOB", "BLOB"},
			{"CLOB", "TEXT", "TEXT"},
			{"REAL", "REAL", "REAL"},
			{"FLOAT", "FLOAT", "FLOAT"},
		},
	},
	{
		driver: "oracle",
		mapping: VendorSqlTypeMapList{
			{"NUMBER", "NUMBER", "NUMBER"},
			{"VARCHAR", "VARCHAR2", "VARCHAR2"},
			{"TEXT", "CLOB", "CLOB"},
			{"INT", "NUMBER", "NUMBER"},
			{"DECIMAL", "NUMBER", "NUMBER"},
			{"VARCHAR2", "VARCHAR2", "VARCHAR2"},
			{"DATE", "DATE", "DATE"},
			{"DATETIME", "TIMESTAMP", "TIMESTAMP"},
			{"TIMESTAMP", "TIMESTAMP", "TIMESTAMP"},
			{"BOOLEAN", "NUMBER", "NUMBER"},
			{"BLOB", "BLOB", "BLOB"},
			{"CLOB", "CLOB", "CLOB"},
			{"REAL", "REAL", "REAL"},
		},
	},
	{
		driver: "sqlserver",
		mapping: VendorSqlTypeMapList{
			{"NUMBER", "DECIMAL", "DECIMAL"},
			{"VARCHAR", "VARCHAR", "VARCHAR"},
			{"TEXT", "TEXT", "TEXT"},
			{"INT", "INT", "INT"},
			{"DECIMAL", "DECIMAL", "DECIMAL"},
			{"VARCHAR2", "VARCHAR", "VARCHAR"},
			{"DATE", "DATE", "DATE"},
			{"DATETIME", "DATETIME", "DATETIME"},
			{"TIMESTAMP", "DATETIME", "DATETIME"},
			{"BOOLEAN", "BIT", "BIT"},
			{"BLOB", "VARBINARY", "VARBINARY"},
			{"CLOB", "TEXT", "TEXT"},
			{"REAL", "REAL", "REAL"},
			{"FLOAT", "FLOAT", "FLOAT"},
		},
	},
	{
		driver: "mssql",
		mapping: VendorSqlTypeMapList{
			{"NUMBER", "DECIMAL", "DECIMAL"},
			{"VARCHAR", "VARCHAR", "VARCHAR"},
			{"TEXT", "TEXT", "TEXT"},
			{"INT", "INT", "INT"},
			{"DECIMAL", "DECIMAL", "DECIMAL"},
			{"VARCHAR2", "VARCHAR", "VARCHAR"},
			{"DATE", "DATE", "DATE"},
			{"DATETIME", "DATETIME", "DATETIME"},
			{"TIMESTAMP", "DATETIME", "DATETIME"},
			{"BOOLEAN", "BIT", "BIT"},
			{"BLOB", "VARBINARY", "VARBINARY"},
			{"CLOB", "TEXT", "TEXT"},
			{"REAL", "REAL", "REAL"},
			{"FLOAT", "FLOAT", "FLOAT"},
		},
	},
	{
		driver: "godror",
		mapping: VendorSqlTypeMapList{
			{"NUMBER", "NUMBER", "NUMBER"},
			{"VARCHAR", "VARCHAR2", "VARCHAR2"},
			{"TEXT", "CLOB", "CLOB"},
			{"INT", "NUMBER", "NUMBER"},
			{"DECIMAL", "NUMBER", "NUMBER"},
			{"VARCHAR2", "VARCHAR2", "VARCHAR2"},
			{"DATE", "DATE", "DATE"},
			{"DATETIME", "TIMESTAMP", "TIMESTAMP"},
			{"TIMESTAMP", "TIMESTAMP", "TIMESTAMP"},
			{"BOOLEAN", "NUMBER", "NUMBER"},
			{"BLOB", "BLOB", "BLOB"},
			{"CLOB", "CLOB", "CLOB"},
			{"REAL", "REAL", "REAL"},
		},
	},
}

func GetVendorSqlTypeMap(driver string) VendorSqlTypeMapList {
	for _, mapping := range vAendorMappingList {
		if mapping.driver == driver {
			return mapping.mapping
		}
	}
	gl.Log("error", fmt.Sprintf("No mapping found for driver %s", driver))
	return nil
}
func GetVendorSqlType(driver, sourceType string) string {
	mapping := GetVendorSqlTypeMap(driver)
	if mapping == nil {
		gl.Log("error", fmt.Sprintf("No mapping found for driver %s", driver))
		return ""
	}

	// Direct mapping lookup
	for _, mapItem := range mapping {
		if mapItem.sourceType == sourceType {
			return mapItem.targetType
		}
	}

	// Intelligent fallbacks for common normalized types
	switch sourceType {
	case "INTEGER":
		return GetVendorSqlType(driver, "INT")
	case "REAL":
		return GetVendorSqlType(driver, "FLOAT")
	case "TEXT":
		return GetVendorSqlType(driver, "VARCHAR")
	case "BLOB":
		return GetVendorSqlType(driver, "BLOB")
	}

	// Ultimate fallback based on driver
	switch driver {
	case "sqlite3", "sqlite":
		switch sourceType {
		case "INTEGER":
			return "INTEGER"
		case "REAL":
			return "REAL"
		case "TEXT":
			return "TEXT"
		case "BLOB":
			return "BLOB"
		default:
			return "TEXT"
		}
	default:
		gl.Log("error", fmt.Sprintf("No mapping found for source type %s in driver %s", sourceType, driver))
		return ""
	}
}

type VJobList struct {
	VJobs []VJob `json:"jobs"`
}

func (jl *VJobList) GetJobs() []Job {
	var jobs []Job
	for _, j := range jl.VJobs {
		jobs = append(jobs, &j)
	}
	return jobs
}
func (jl *VJobList) AddJob(job Job) {
	var j *VJob
	j = job.(*VJob)
	jl.VJobs = append(jl.VJobs, *j)
}
func (jl *VJobList) RemoveJob(job Job) {
	var j *VJob
	j = job.(*VJob)
	for i, v := range jl.VJobs {
		if v.VID == j.VID {
			jl.VJobs = append(jl.VJobs[:i], jl.VJobs[i+1:]...)
		}
	}
}

type VJob struct {
	VID           string `json:"id"`
	VName         string `json:"name"`
	VDescription  string `json:"description"`
	VConfig       Config `json:"config"`
	VSchedule     string `json:"schedule"`
	VLastRun      string `json:"lastRun"`
	VNextRun      string `json:"nextRun"`
	VOutputPath   string `json:"outputPath"`
	VOutputFormat string `json:"outputFormat"`
	VNeedCheck    bool   `json:"needCheck"`
	VCheckMethod  string `json:"checkMethod"`
	VPath         string `json:"path"`
}

func (j *VJob) Execute() error       { return nil }
func (j *VJob) ID() string           { return j.VID }
func (j *VJob) Name() string         { return j.VName }
func (j *VJob) Description() string  { return j.VDescription }
func (j *VJob) Config() Config       { return j.VConfig }
func (j *VJob) Schedule() string     { return j.VSchedule }
func (j *VJob) LastRun() string      { return j.VLastRun }
func (j *VJob) NextRun() string      { return j.VNextRun }
func (j *VJob) OutputPath() string   { return j.VOutputPath }
func (j *VJob) OutputFormat() string { return j.VOutputFormat }
func (j *VJob) NeedCheck() bool      { return j.VNeedCheck }
func (j *VJob) CheckMethod() string  { return j.VCheckMethod }
func (j *VJob) Path() string         { return j.VPath }

type Job interface {
	Execute() error
	ID() string
	Name() string
	Description() string
	Config() Config
	Schedule() string
	LastRun() string
	NextRun() string
	OutputPath() string
	OutputFormat() string
	NeedCheck() bool
	CheckMethod() string
	Path() string
}
type JobList interface {
	GetJobs() []Job
	AddJob(job Job)
	RemoveJob(job Job)
}
