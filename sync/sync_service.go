package etl

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/kubex-ecosystem/getl/meta"
)

type SyncService struct {
	db        *sql.DB
	interval  time.Duration
	tableName string
}

func NewSyncService(db *sql.DB, interval time.Duration, tableName string) *SyncService {
	return &SyncService{
		db:        db,
		interval:  interval,
		tableName: tableName,
	}
}

func (s *SyncService) Start() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			changed, err := meta.CheckAndUpdateHashes(s.db, s.tableName)
			if err != nil {
				fmt.Printf("Erro ao verificar e atualizar hashes: %v\n", err)
				continue
			}
			if changed {
				fmt.Println("Dados alterados, sincronizando...")
				// Adicione a lógica de sincronização aqui
			} else {
				fmt.Println("Nenhuma alteração detectada.")
			}
		}
	}
}
