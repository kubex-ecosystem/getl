package etl

import (
	"database/sql"
	"time"

	gl "github.com/kubex-ecosystem/getl/internal/module/logger"
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
				gl.Log("error", "Erro ao verificar e atualizar hashes:", err)
				continue
			}
			if changed {
				gl.Log("info", "Dados alterados, sincronizando...")
				// Adicione a lógica de sincronização aqui
			} else {
				gl.Log("info", "Nenhuma alteração detectada.")
			}
		}
	}
}
