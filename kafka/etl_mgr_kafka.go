package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	. "github.com/kubex-ecosystem/getl/etypes"
	s "github.com/kubex-ecosystem/getl/sql"
	gl "github.com/kubex-ecosystem/getl/internal/module/logger"
	"github.com/segmentio/kafka-go"
)

type IKafka interface {
	GetKafkaReader() *kafka.Reader
	GetKafkaWriter() *kafka.Writer
	GetConfig() Config
	GetSourceType() string
	GetSourceConnectionString() string
	GetDestinationType() string
	GetDestinationConnectionString() string
	GetTopic() string
	GetGroupID() string
	GetKafkaURL() string

	SetConfig(config Config)
	SetSourceType(sourceType string)
	SetDestinationType(destinationType string)
	SetTopic(topic string)
	SetGroupID(groupID string)
	SetKafkaURL(kafkaURL string)
	SetSourceConnectionString(sourceConnectionString string)
	SetDestinationConnectionString(destinationConnectionString string)

	Close()
	SyncData()
	RunETL() error
}
type Kafka struct {
	// mu is the mutex for the Kafka struct
	mu *sync.Mutex
	// Reader is the Kafka reader
	Reader *kafka.Reader `json:"reader"`
	// Writer is the Kafka writer
	Writer *kafka.Writer `json:"writer"`
	// KafkaURL is the URL of the Kafka broker
	KafkaURL string `json:"kafka_url"`
	// Topic is the Kafka topic to read from
	Topic string `json:"topic"`
	// GroupID is the Kafka group ID to read from
	GroupID string `json:"group_id"`
	// SourceType is the type of the source database
	SourceType string `json:"source_type"`
	// SourceConnectionString is the connection string for the source database
	SourceConnectionString string `json:"source_connection_string"`
	// DestinationType is the type of the destination database
	DestinationType string `json:"destination_type"`
	// DestinationConnectionString is the connection string for the destination database
	DestinationConnectionString string `json:"destination_connection_string"`
	// Config is the configuration for the Kafka struct
	KafkaConfig Config `json:"config"`
}

func NewSimpleKafka(kafkaURL, topic, groupID string) IKafka {
	return &Kafka{
		mu:       &sync.Mutex{},
		KafkaURL: kafkaURL,
		Topic:    topic,
		GroupID:  groupID,
	}
}
func NewKafka(kafkaURL, topic, groupID, sourceType, sourceConnectionString, destinationType, destinationConnectionString string) IKafka {
	return &Kafka{
		mu:                          &sync.Mutex{},
		KafkaURL:                    kafkaURL,
		Topic:                       topic,
		GroupID:                     groupID,
		SourceType:                  sourceType,
		SourceConnectionString:      sourceConnectionString,
		DestinationType:             destinationType,
		DestinationConnectionString: destinationConnectionString,
	}
}
func NewKafkaWithConfig(kafkaURL, topic, groupID, sourceType, sourceConnectionString, destinationType, destinationConnectionString string, config Config) IKafka {
	return &Kafka{
		mu:                          &sync.Mutex{},
		KafkaURL:                    kafkaURL,
		Topic:                       topic,
		GroupID:                     groupID,
		SourceType:                  sourceType,
		SourceConnectionString:      sourceConnectionString,
		DestinationType:             destinationType,
		DestinationConnectionString: destinationConnectionString,
		KafkaConfig:                 config,
	}
}
func NewKafkaWithConfigAndSQL(kafkaURL, topic, groupID, sourceType, sourceConnectionString, destinationType, destinationConnectionString string, config Config) IKafka {
	return &Kafka{
		mu:                          &sync.Mutex{},
		KafkaURL:                    kafkaURL,
		Topic:                       topic,
		GroupID:                     groupID,
		SourceType:                  sourceType,
		SourceConnectionString:      sourceConnectionString,
		DestinationType:             destinationType,
		DestinationConnectionString: destinationConnectionString,
		KafkaConfig:                 config,
	}
}

func (k *Kafka) GetKafkaReader() *kafka.Reader {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.Reader == nil {
		k.Reader = kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{k.KafkaURL},
			Topic:   k.Topic,
			GroupID: k.GroupID,
		})
	}
	return k.Reader
}
func (k *Kafka) GetKafkaWriter() *kafka.Writer {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.Writer == nil {
		k.Writer = &kafka.Writer{
			Addr:     kafka.TCP(k.KafkaURL),
			Topic:    k.Topic,
			Balancer: &kafka.LeastBytes{},
		}
	}
	return k.Writer
}
func (k *Kafka) GetConfig() Config                         { return k.KafkaConfig }
func (k *Kafka) GetSourceType() string                     { return k.SourceType }
func (k *Kafka) GetSourceConnectionString() string         { return k.SourceConnectionString }
func (k *Kafka) GetDestinationType() string                { return k.DestinationType }
func (k *Kafka) GetDestinationConnectionString() string    { return k.DestinationConnectionString }
func (k *Kafka) GetTopic() string                          { return k.Topic }
func (k *Kafka) GetGroupID() string                        { return k.GroupID }
func (k *Kafka) GetKafkaURL() string                       { return k.KafkaURL }
func (k *Kafka) SetConfig(config Config)                   { k.KafkaConfig = config }
func (k *Kafka) SetSourceType(sourceType string)           { k.SourceType = sourceType }
func (k *Kafka) SetDestinationType(destinationType string) { k.DestinationType = destinationType }
func (k *Kafka) SetTopic(topic string)                     { k.Topic = topic }
func (k *Kafka) SetGroupID(groupID string)                 { k.GroupID = groupID }
func (k *Kafka) SetKafkaURL(kafkaURL string)               { k.KafkaURL = kafkaURL }
func (k *Kafka) SetSourceConnectionString(sourceConnectionString string) {
	k.SourceConnectionString = sourceConnectionString
}
func (k *Kafka) SetDestinationConnectionString(destinationConnectionString string) {
	k.DestinationConnectionString = destinationConnectionString
}
func (k *Kafka) Close() {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.Reader != nil {
		if err := k.Reader.Close(); err != nil {
			gl.Log("error", "erro ao fechar o leitor Kafka: "+err.Error())
		}
	}
	if k.Writer != nil {
		if err := k.Writer.Close(); err != nil {
			gl.Log("error", "erro ao fechar o escritor Kafka: "+err.Error())
		}
	}
}
func (k *Kafka) SyncData() {
	db, openErr := sql.Open(k.DestinationType, k.DestinationConnectionString)
	if openErr != nil {
		gl.Log("error", "erro ao conectar ao banco de dados de destino: "+openErr.Error())
		return
	}
	defer db.Close()

	kafkaReader := k.GetKafkaReader()
	for {
		msg, kafkaReaderErr := kafkaReader.ReadMessage(context.Background())
		if kafkaReaderErr != nil {
			gl.Log("error", "erro ao ler mensagem do Kafka: "+kafkaReaderErr.Error())
			continue
		}

		var row = make(map[string]interface{})
		if unmarshalErr := json.Unmarshal(msg.Value, &row); unmarshalErr != nil {
			gl.Log("error", "erro ao decodificar mensagem do Kafka: "+unmarshalErr.Error())
			continue
		}

		if loadDataErr := s.LoadData(db, k.KafkaConfig); loadDataErr != nil {
			gl.Log("error", "erro ao carregar dados no banco de destino: "+loadDataErr.Error())
		}
	}
}
func (k *Kafka) RunETL() error {
	db, dbErr := sql.Open(k.SourceType, k.SourceConnectionString)
	if dbErr != nil {
		return fmt.Errorf("falha ao conectar ao banco de dados de origem: %w", dbErr)
	}
	defer db.Close()

	rows, rowsErr := db.Query(k.KafkaConfig.SQLQuery)
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

		err = k.GetKafkaWriter().WriteMessages(context.Background(), kafka.Message{
			Value: message,
		})
		if err != nil {
			return fmt.Errorf("falha ao escrever mensagem no Kafka: %w", err)
		}
	}

	return nil
}

func CreateKafkaConfig(kafkaURL, topic, groupID, sourceType, sourceConnectionString, destinationType, destinationConnectionString string) Config {
	return Config{
		KafkaURL:                    kafkaURL,
		SourceType:                  sourceType,
		SourceConnectionString:      sourceConnectionString,
		DestinationType:             destinationType,
		DestinationConnectionString: destinationConnectionString,
	}
}
func CreateKafkaConfigWithSQL(kafkaURL, topic, groupID, sourceType, sourceConnectionString, destinationType, destinationConnectionString, sqlQuery string) Config {
	return Config{
		KafkaURL: kafkaURL,
		//Topic:                       topic,
		//GroupID:                     groupID,
		SourceType:                  sourceType,
		SourceConnectionString:      sourceConnectionString,
		DestinationType:             destinationType,
		DestinationConnectionString: destinationConnectionString,
		SQLQuery:                    sqlQuery,
	}
}
func CreateKafkaConfigWithSQLAndTopic(kafkaURL, topic, groupID, sourceType, sourceConnectionString, destinationType, destinationConnectionString, sqlQuery string) Config {
	return Config{
		KafkaURL: kafkaURL,
		//Topic:                       topic,
		//GroupID:                     groupID,
		SourceType:                  sourceType,
		SourceConnectionString:      sourceConnectionString,
		DestinationType:             destinationType,
		DestinationConnectionString: destinationConnectionString,
		SQLQuery:                    sqlQuery,
	}
}
func CreateKafkaConfigWithSQLAndTopicAndGroupID(kafkaURL, topic, groupID, sourceType, sourceConnectionString, destinationType, destinationConnectionString, sqlQuery string) Config {
	return Config{
		KafkaURL: kafkaURL,
		//Topic:                       topic,
		//GroupID:                     groupID,
		SourceType:                  sourceType,
		SourceConnectionString:      sourceConnectionString,
		DestinationType:             destinationType,
		DestinationConnectionString: destinationConnectionString,
		SQLQuery:                    sqlQuery,
	}
}

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

func CreateKafkaReaderWithConfig(kafkaURL, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaURL},
		Topic:   topic,
		GroupID: groupID,
	})
}
func CreateKafkaWriterWithConfig(kafkaURL, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(kafkaURL),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func SyncData(config Config, kafkaReader *kafka.Reader) {
	db, openErr := sql.Open(config.DestinationType, config.DestinationConnectionString)
	if openErr != nil {
		gl.Log("error", "erro ao conectar ao banco de dados de destino: "+openErr.Error())
		return
	}
	defer db.Close()

	for {
		msg, kafkaReaderErr := kafkaReader.ReadMessage(context.Background())
		if kafkaReaderErr != nil {
			gl.Log("error", "erro ao ler mensagem do Kafka: "+kafkaReaderErr.Error())
			continue
		}

		var row = make(map[string]interface{})
		if unmarshalErr := json.Unmarshal(msg.Value, &row); unmarshalErr != nil {
			gl.Log("error", "erro ao decodificar mensagem do Kafka: "+unmarshalErr.Error())
			continue
		}

		if loadDataErr := s.LoadData(db, config); loadDataErr != nil {
			gl.Log("error", "erro ao carregar dados no banco de destino: "+loadDataErr.Error())
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
