package sql

import (
	"database/sql"
	"fmt"
	"time"

	. "github.com/kubex-ecosystem/getl/etypes"
)

// connectDB estabelece uma conexão com o banco de dados especificado na configuração.
// config: a configuração contendo o tipo de banco de dados e a string de conexão.
// Retorna um ponteiro para a conexão com o banco de dados e um erro, se houver.
func connectDB(config Config) (*sql.DB, error) {
	db, err := sql.Open(config.DestinationType, config.DestinationConnectionString)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao banco de dados: %w", err)
	}

	// Verifica a conexão
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("falha ao verificar a conexão com o banco de dados: %w", err)
	}

	return db, nil
}

// reconnectDB tenta restabelecer a conexão com o banco de dados especificado na configuração.
// config: a configuração contendo o tipo de banco de dados e a string de conexão.
// Tenta reconectar até 5 vezes com um intervalo de 10 segundos entre as tentativas.
// Retorna um ponteiro para a conexão com o banco de dados e um erro, se houver.
func reconnectDB(config Config) (*sql.DB, error) {
	var db *sql.DB
	var err error
	maxRetries := 5
	retryInterval := 10 * time.Second

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open(config.DestinationType, config.DestinationConnectionString)
		if err == nil {
			if pingErr := db.Ping(); pingErr == nil {
				return db, nil
			}
		}
		time.Sleep(retryInterval)
	}

	return nil, fmt.Errorf("falha ao reconectar ao banco de dados após %d tentativas: %w", maxRetries, err)
}
