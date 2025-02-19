package templates

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type ETLConfig struct {
	SourceType                  string           `json:"sourceType"`
	SourceConnectionString      string           `json:"sourceConnectionString"`
	DestinationType             string           `json:"destinationType"`
	DestinationConnectionString string           `json:"destinationConnectionString"`
	DestinationTable            string           `json:"destinationTable,omitempty"`
	SQLQuery                    string           `json:"sqlQuery"`
	KafkaURL                    string           `json:"kafkaURL,omitempty"`
	KafkaTopic                  string           `json:"kafkaTopic,omitempty"`
	KafkaGroupID                string           `json:"kafkaGroupID,omitempty"`
	Transformations             []Transformation `json:"transformations,omitempty"`
}

type Transformation struct {
	SourceField      string `json:"sourceField"`
	DestinationField string `json:"destinationField"`
	Operation        string `json:"operation"`
	SPath            string `json:"sPath"`
	DPath            string `json:"dPath"`
}

func main() {
	var configType string
	flag.StringVar(&configType, "type", "", "Type of ETL config to generate (oracle-to-sqlite or sqlite-to-sqlite)")
	flag.Parse()

	var config ETLConfig

	switch configType {
	case "oracle-to-sqlite":
		config = ETLConfig{
			SourceType:                  "godror",
			SourceConnectionString:      "sankhya/Bioextratus2000@127.0.0.1:1521/orcl",
			DestinationType:             "sqlite3",
			DestinationConnectionString: "/home/user/.kubex/web/gorm.db",
			DestinationTable:            "erp_products",
			SQLQuery:                    "SELECT P.CODPARC, P.NOMEPARC FROM TGFPAR P",
		}
	case "sqlite-to-sqlite":
		config = ETLConfig{
			SourceType:                  "sqlite3",
			SourceConnectionString:      "/home/user/.kubex/web/gorm.db",
			DestinationType:             "sqlite3",
			DestinationConnectionString: "/home/user/.kubex/web/gorm.db",
			Transformations: []Transformation{
				{SourceField: "CODPROD", DestinationField: "id_v", Operation: "none", SPath: "erp_products", DPath: "erp_products_teste"},
				{SourceField: "DESCRPROD", DestinationField: "name_v", Operation: "none", SPath: "erp_products", DPath: "erp_products_teste"},
				{SourceField: "DESCRGRUPOPROD", DestinationField: "depart_v", Operation: "none", SPath: "erp_products", DPath: "erp_products_teste"},
				{SourceField: "ATIVO", DestinationField: "active", Operation: "none", SPath: "erp_products", DPath: "erp_products_teste"},
				{SourceField: "ESTOQUE", DestinationField: "stock", Operation: "none", SPath: "erp_products", DPath: "erp_products_teste"},
				{SourceField: "RESERVADO", DestinationField: "reserved", Operation: "none", SPath: "erp_products", DPath: "erp_products_teste"},
				{SourceField: "SALDO", DestinationField: "balance", Operation: "none", SPath: "erp_products", DPath: "erp_products_teste"},
				{SourceField: "PRECO", DestinationField: "price", Operation: "none", SPath: "erp_products", DPath: "erp_products_teste"},
			},
		}
	default:
		fmt.Println("Invalid type. Use 'oracle-to-sqlite' or 'sqlite-to-sqlite'.")
		return
	}

	file, err := os.Create("etl_config.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	fmt.Println("ETL config file generated successfully.")
}
