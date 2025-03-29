package meta

import (
	"database/sql"
	"errors"
	"fmt"
	. "github.com/faelmori/getl/etypes"
	//"github.com/faelmori/kbx/mods/utils"
	"github.com/faelmori/gkbxsrv/utils"
)

func CreateInternalSchema(db *sql.DB) error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS etl_meta_info (
		table_name TEXT PRIMARY KEY,
		hash TEXT
	)`
	_, err := db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("falha ao criar esquema interno: %w", err)
	}
	return nil
}

func CheckAndUpdateHashes(db *sql.DB, tableName string) (bool, error) {
	// Simulação de dados para exemplo
	data := []map[string]string{
		{"DESCRGRUPOPROD": "Grupo1", "ATIVO": "1", "ESTOQUE": "100", "CODPROD": "001"},
		{"DESCRPROD": "Produto1", "RESERVADO": "10", "SALDO": "90", "PRECO": "10.00"},
	}

	hash := utils.NewHash()
	for _, row := range data {
		hash.Write([]byte(fmt.Sprintf("%s%s%s%s", row["DESCRGRUPOPROD"], row["ATIVO"], row["ESTOQUE"], row["CODPROD"])))
		hash.Write([]byte(fmt.Sprintf("%s%s%s%s", row["DESCRPROD"], row["RESERVADO"], row["SALDO"], row["PRECO"])))
	}
	newHash := hash.Sum(nil)

	var existingHash string
	err := db.QueryRow("SELECT hash FROM etl_meta_info WHERE table_name = ?", tableName).Scan(&existingHash)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("falha ao obter hash existente: %w", err)
	}

	if string(newHash) == existingHash {
		return false, nil // Nenhuma alteração
	}

	if existingHash == "" {
		_, err = db.Exec("INSERT INTO etl_meta_info (table_name, hash) VALUES (?, ?)", tableName, newHash)
	} else {
		_, err = db.Exec("UPDATE etl_meta_info SET hash = ? WHERE table_name = ?", newHash, tableName)
	}
	if err != nil {
		return false, fmt.Errorf("falha ao atualizar hash: %w", err)
	}

	return true, nil // Dados alterados
}

// Função para criar o esquema interno no banco de dados
func createInternalSchema(db *sql.DB) error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS etl_meta_info (
		table_name TEXT PRIMARY KEY,
		hash TEXT
	)`
	_, err := db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("falha ao criar esquema interno: %w", err)
	}
	return nil
}

// Função para verificar e atualizar hashes
func checkAndUpdateHashes(db *sql.DB, tableName string, data []Data) (bool, error) {
	// Gerar um hash incremental para o conjunto de dados
	hash := utils.NewHash()
	for _, row := range data {
		hash.Write([]byte(fmt.Sprintf("%s%s%s%s", row["DESCRGRUPOPROD"], row["ATIVO"], row["ESTOQUE"], row["CODPROD"])))
		hash.Write([]byte(fmt.Sprintf("%s%s%s%s", row["DESCRPROD"], row["RESERVADO"], row["SALDO"], row["PRECO"])))
	}
	newHash := hash.Sum(nil)

	// Verificar se o hash mudou
	var existingHash string
	err := db.QueryRow("SELECT hash FROM etl_meta_info WHERE table_name = ?", tableName).Scan(&existingHash)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("falha ao obter hash existente: %w", err)
	}

	if string(newHash) == existingHash {
		return false, nil // Nenhuma alteração
	}

	// Atualizar o hash no banco de dados
	if existingHash == "" {
		_, err = db.Exec("INSERT INTO etl_meta_info (table_name, hash) VALUES (?, ?)", tableName, newHash)
	} else {
		_, err = db.Exec("UPDATE etl_meta_info SET hash = ? WHERE table_name = ?", newHash, tableName)
	}
	if err != nil {
		return false, fmt.Errorf("falha ao atualizar hash: %w", err)
	}

	return true, nil // Dados alterados
}
