package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	. "github.com/faelmori/kbx/mods/getl/etypes"
	_ "github.com/go-sql-driver/mysql"
	"github.com/segmentio/kafka-go"
	"testing"
)

func TestRunETL(t *testing.T) {
	// Configuração do banco de dados de teste
	db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/testdb")
	if err != nil {
		t.Fatalf("falha ao conectar ao banco de dados de teste: %v", err)
	}
	defer db.Close()

	// Configuração do Kafka de teste
	kafkaURL := "localhost:9092"
	topic := "test_topic"
	writer := &kafka.Writer{
		Addr:  kafka.TCP(kafkaURL),
		Topic: topic,
	}
	defer writer.Close()

	// Configuração do ETL
	config := Config{
		SourceType:             "mysql",
		SourceConnectionString: "user:password@tcp(localhost:3306)/testdb",
		SQLQuery: `WITH PRO AS (
		            SELECT P.CODPROD, P.DESCRPROD, P.CODGRUPOPROD, G.DESCRGRUPOPROD, P.IMAGEM, P.ATIVO
		            FROM TGFPRO P
		            INNER JOIN TGFGRU G ON P.CODGRUPOPROD = G.CODGRUPOPROD
		            WHERE P.USOPROD = 'R'
		        ), EST AS (
		            SELECT E.CODPROD, SUM(E.ESTOQUE) AS ESTOQUE, SUM(E.RESERVADO) AS RESERVADO, (SUM(E.ESTOQUE) - SUM(E.RESERVADO)) AS SALDO
		            FROM PRO P
		            LEFT JOIN TGFEST E ON E.CODPROD = P.CODPROD
		            GROUP BY E.CODPROD
		        ), PRECO AS (
		            SELECT P.CODPROD, ROUND(NVL(SNK_PRECO(0, P.CODPROD), 0), 2) AS PRECO
		            FROM PRO P
		            LEFT JOIN TGFPRC PR ON PR.CODPROD = P.CODPROD
		            WHERE PR.CODTABELA = 1
		            GROUP BY P.CODPROD
		        ), CONSOLIDA AS (
		            SELECT P.CODPROD, P.DESCRPROD, P.DESCRGRUPOPROD, P.IMAGEM, P.ATIVO, E.ESTOQUE, E.RESERVADO, E.SALDO, R.PRECO
		            FROM PRO P
		            LEFT JOIN EST E ON E.CODPROD = P.CODPROD
		            LEFT JOIN PRECO R ON R.CODPROD = P.CODPROD
		        )
		        SELECT * FROM CONSOLIDA`,
	}

	// Executar o ETL
	if err := RunETL(config, writer); err != nil {
		t.Fatalf("falha ao executar o GETl: %v", err)
	}

	// Verificar os resultados no Kafka
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaURL},
		Topic:   topic,
		GroupID: "test_group",
	})
	defer reader.Close()

	msg, err := reader.ReadMessage(context.Background())
	if err != nil {
		t.Fatalf("falha ao ler mensagem do Kafka: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(msg.Value, &result); err != nil {
		t.Fatalf("falha ao deserializar mensagem: %v", err)
	}

	// Verificar se os resultados estão corretos
	expected := map[string]interface{}{
		"CODPROD":        1,
		"DESCRPROD":      "Produto 1",
		"CODGRUPOPROD":   1,
		"DESCRGRUPOPROD": "Grupo 1",
		"IMAGEM":         "imagem1.jpg",
		"ATIVO":          true,
		"ESTOQUE":        100,
		"RESERVADO":      10,
		"SALDO":          90,
		"PRECO":          10.0,
	}

	for key, value := range expected {
		if result[key] != value {
			t.Errorf("esperado %v para %s, mas obteve %v", value, key, result[key])
		}
	}
}
