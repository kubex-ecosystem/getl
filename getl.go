package getl

import (
	"database/sql"
	"fmt"
	e "github.com/faelmori/getl/edi"
	t "github.com/faelmori/getl/etypes"
	x "github.com/faelmori/getl/genx"
	k "github.com/faelmori/getl/kafka"
	p "github.com/faelmori/getl/protoextr"
	s "github.com/faelmori/getl/sql"
	y "github.com/faelmori/getl/sync"
	l "github.com/faelmori/logz"
)

// Getl is the main struct for the getl package
type Getl struct {
	// ConfigPath is the path to the config file
	ConfigPath string `json:"config_path"`

	// Config is the configuration for the getl package
	Config *t.Config `json:"config"`

	// Kafka is the kafka client for the getl package
	Kafka k.IKafka `json:"kafka"`

	// SQL is the sql client for the getl package
	SQLEvent *s.Event `json:"sql_event"`

	// Sync is the sync client for the getl package
	Sync *y.SyncService `json:"sync"`

	// Protoextr is the protoextr client for the getl package
	Protoextr *p.ProtobufDataTable `json:"protoextr"`

	// EDI is the edi client for the getl package
	EdiOCOREN *e.OCOREN `json:"OCOREN"`
	EdiCONEMB *e.CONEMB `json:"CONEMB"`

	// Genx is the genx client for the getl package
	//Genx *x.Genx `json:"genx"`

}

func (g *Getl) SerializeYAMLToJSON(v interface{}) string  { return x.SerializeYAMLToJSON(v) }
func (g *Getl) SerializeJSONToYAML(v interface{}) string  { return x.SerializeJSONToYAML(v) }
func (g *Getl) SerializeSQLInsert(v interface{}) string   { return x.SerializeSQLInsert(v) }
func (g *Getl) SerializeSQLUpdate(v interface{}) string   { return x.SerializeSQLInsert(v) } //{ return x.SerializeSQLUpdate(v) }
func (g *Getl) SerializeSQLDelete(v interface{}) string   { return x.SerializeSQLInsert(v) } //{ return x.SerializeSQLDelete(v) }
func (g *Getl) SerializeSQLSelect(v interface{}) string   { return x.SerializeSQLInsert(v) } //{ return x.SerializeSQLSelect(v)}
func (g *Getl) SerializeSQLCreate(v interface{}) string   { return x.SerializeSQLInsert(v) } //{ return x.SerializeSQLCreate(v) }
func (g *Getl) SerializeSQLTruncate(v interface{}) string { return x.SerializeSQLInsert(v) } //{ return x.SerializeSQLTruncate(v) }

func (g *Getl) ExecuteETL() error {
	if err := s.ExecuteETL(g.ConfigPath, g.Config.OutputPath, g.Config.OutputFormat, g.Config.NeedCheck, g.Config.CheckMethod); err != nil {
		l.Error("Error executing ETL", map[string]interface{}{})
		return err
	}
	l.Info("ETL executed successfully", map[string]interface{}{})
	return nil
}

// ExtractData extracts data from the source
func (g *Getl) ExtractData(dbSQL *sql.DB, config t.Config) ([]t.Data, []string, error) {
	data, dataArr, err := s.ExtractData(dbSQL, config)
	if err != nil {
		l.Error("Error extracting data", map[string]interface{}{})
		return nil, nil, err
	}
	l.Info("Data extracted successfully", map[string]interface{}{})
	return data, dataArr, nil
}

// LoadData loads data into the target
func (g *Getl) LoadData(dbSQL *sql.DB, config t.Config) error {
	if err := s.LoadData(dbSQL, config); err != nil {
		l.Error("Error loading data", map[string]interface{}{})
		return err
	}
	l.Info("Data loaded successfully", map[string]interface{}{})
	return nil
}

// VacuumDatabase cleans up the database
func (g *Getl) VacuumDatabase(config t.Config) error {
	if err := s.VacuumDatabase(config.DestinationConnectionString); err != nil {
		l.Error("Error vacuuming database", map[string]interface{}{})
		return err
	}
	return nil
}

// SaveData saves data to the database
func (g *Getl) SaveData(config t.Config) error {
	data, _, err := s.ExtractDataWithTypes(nil, config)
	if err != nil {
		l.Error("Error extracting data", map[string]interface{}{})
		return err
	}
	if saveErr := s.SaveData(config.OutputPath, data, config.OutputFormat); saveErr != nil {
		l.Error("Error saving data", map[string]interface{}{})
		return saveErr
	}
	l.Info(fmt.Sprintf("Data saved to %s", config.OutputPath), map[string]interface{}{})
	return nil
}

// DataTable views data from the database
func (g *Getl) DataTable(configPath, outputPath, outputFormat string, export bool) error {
	return s.ShowDataTableFromConfig(configPath, export, outputPath, outputFormat)
}
