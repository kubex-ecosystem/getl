package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	. "github.com/faelmori/getl/etypes"
	. "github.com/faelmori/getl/sql"
	"github.com/faelmori/logz"
	"github.com/segmentio/kafka-go"
)

func CreateKafkaWriter(kafkaURL, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(kafkaURL),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func CreateKafkaReader(kafkaURL, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaURL},
		Topic:   topic,
		GroupID: groupID,
	})
}

func SyncData(config Config, kafkaReader *kafka.Reader) {
	db, openErr := sql.Open(config.DestinationType, config.DestinationConnectionString)
	if openErr != nil {
		_ = logz.ErrorLog("erro ao conectar ao banco de dados de destino: "+openErr.Error(), "etl")
		return
	}
	defer db.Close()

	for {
		msg, kafkaReaderErr := kafkaReader.ReadMessage(context.Background())
		if kafkaReaderErr != nil {
			_ = logz.ErrorLog("erro ao ler mensagem do Kafka: "+kafkaReaderErr.Error(), "etl")
			continue
		}

		var row = make(map[string]interface{})
		if unmarshalErr := json.Unmarshal(msg.Value, &row); unmarshalErr != nil {
			_ = logz.ErrorLog("erro ao decodificar mensagem do Kafka: "+unmarshalErr.Error(), "etl")
			continue
		}

		if loadDataErr := LoadData(db, config); loadDataErr != nil {
			_ = logz.ErrorLog("erro ao carregar dados no banco de destino: "+loadDataErr.Error(), "etl")
		}
	}
}

func RunETL(config Config, kafkaWriter *kafka.Writer) error {
	db, dbErr := sql.Open(config.SourceType, config.SourceConnectionString)
	if dbErr != nil {
		return fmt.Errorf("falha ao conectar ao banco de dados de origem: %w", dbErr)
	}
	defer db.Close()

	rows, rowsErr := db.Query(config.SQLQuery)
	if rowsErr != nil {
		return fmt.Errorf("falha ao executar a consulta SQL: %w", rowsErr)
	}
	defer rows.Close()

	columns, columnsErr := rows.Columns()
	if columnsErr != nil {
		return fmt.Errorf("falha ao obter colunas: %w", columnsErr)
	}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("falha ao escanear linha: %w", err)
		}

		rowMap := make(map[string]interface{})
		for i, col := range columns {
			rowMap[col] = values[i]
		}

		message, err := json.Marshal(rowMap)
		if err != nil {
			return fmt.Errorf("falha ao serializar linha: %w", err)
		}

		err = kafkaWriter.WriteMessages(context.Background(), kafka.Message{
			Value: message,
		})
		if err != nil {
			return fmt.Errorf("falha ao escrever mensagem no Kafka: %w", err)
		}
	}

	return nil
}
