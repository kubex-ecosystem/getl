package etl

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/kubex-ecosystem/getl/etypes"
	gl "github.com/kubex-ecosystem/getl/internal/module/logger"
)

// IncrementalSyncManager handles incremental synchronization between databases
type IncrementalSyncManager struct {
	config Config
	state  SyncState
}

// NewIncrementalSyncManager creates a new incremental sync manager
func NewIncrementalSyncManager(config Config) *IncrementalSyncManager {
	manager := &IncrementalSyncManager{
		config: config,
		state: SyncState{
			SourceTable:      config.SourceTable,
			DestinationTable: config.DestinationTable,
			Strategy:         string(config.IncrementalSync.Strategy),
			LastSyncTime:     time.Now().Format(time.RFC3339),
		},
	}

	// Load existing state if available
	if config.IncrementalSync.StateFile != "" {
		manager.loadState()
	}

	return manager
}

// ExecuteIncrementalSync performs incremental synchronization based on the configured strategy
func (ism *IncrementalSyncManager) ExecuteIncrementalSync() error {
	if !ism.config.IncrementalSync.Enabled {
		gl.Log("info", "Incremental sync is disabled, performing full sync")
		return ism.executeFullSync()
	}

	gl.Log("info", fmt.Sprintf("Starting incremental sync with strategy: %s", ism.config.IncrementalSync.Strategy))

	switch ism.config.IncrementalSync.Strategy {
	case TimestampBased:
		return ism.executeTimestampBasedSync()
	case PrimaryKeyBased:
		return ism.executePrimaryKeyBasedSync()
	case HashBased:
		return ism.executeHashBasedSync()
	default:
		return ism.executeFullSync()
	}
}

// executeTimestampBasedSync synchronizes based on timestamp columns
func (ism *IncrementalSyncManager) executeTimestampBasedSync() error {
	gl.Log("info", "Executing timestamp-based incremental sync")

	if ism.config.IncrementalSync.TimestampField == "" {
		return fmt.Errorf("timestamp field is required for timestamp-based sync")
	}

	// Build incremental query
	var whereClause string
	if ism.state.LastSyncValue != nil {
		whereClause = fmt.Sprintf("WHERE %s > '%v'", ism.config.IncrementalSync.TimestampField, ism.state.LastSyncValue)
	}

	// Update the SQL query to include the WHERE clause
	originalQuery := ism.config.SQLQuery
	if originalQuery == "" {
		originalQuery = fmt.Sprintf("SELECT * FROM %s", ism.config.SourceTable)
	}

	if whereClause != "" {
		if strings.Contains(strings.ToUpper(originalQuery), "WHERE") {
			ism.config.SQLQuery = originalQuery + " AND " + strings.TrimPrefix(whereClause, "WHERE ")
		} else {
			ism.config.SQLQuery = originalQuery + " " + whereClause
		}
	}

	gl.Log("info", fmt.Sprintf("Incremental query: %s", ism.config.SQLQuery))

	// Execute the sync
	err := ism.executeSyncWithQuery()
	if err != nil {
		return err
	}

	// Update state with the latest timestamp
	return ism.updateTimestampState()
}

// executePrimaryKeyBasedSync synchronizes based on primary key ranges
func (ism *IncrementalSyncManager) executePrimaryKeyBasedSync() error {
	gl.Log("info", "Executing primary key-based incremental sync")

	if ism.config.PrimaryKey == "" {
		return fmt.Errorf("primary key is required for primary key-based sync")
	}

	// Build incremental query based on last primary key value
	var whereClause string
	if ism.state.LastSyncValue != nil {
		whereClause = fmt.Sprintf("WHERE %s > %v", ism.config.PrimaryKey, ism.state.LastSyncValue)
	}

	// Update the SQL query
	originalQuery := ism.config.SQLQuery
	if originalQuery == "" {
		originalQuery = fmt.Sprintf("SELECT * FROM %s", ism.config.SourceTable)
	}

	if whereClause != "" {
		if strings.Contains(strings.ToUpper(originalQuery), "WHERE") {
			ism.config.SQLQuery = originalQuery + " AND " + strings.TrimPrefix(whereClause, "WHERE ")
		} else {
			ism.config.SQLQuery = originalQuery + " " + whereClause
		}
	}

	// Add ORDER BY to ensure consistent results
	if !strings.Contains(strings.ToUpper(ism.config.SQLQuery), "ORDER BY") {
		ism.config.SQLQuery += fmt.Sprintf(" ORDER BY %s", ism.config.PrimaryKey)
	}

	gl.Log("info", fmt.Sprintf("Incremental query: %s", ism.config.SQLQuery))

	// Execute the sync
	err := ism.executeSyncWithQuery()
	if err != nil {
		return err
	}

	// Update state with the latest primary key value
	return ism.updatePrimaryKeyState()
}

// executeHashBasedSync synchronizes based on row hashing
func (ism *IncrementalSyncManager) executeHashBasedSync() error {
	gl.Log("info", "Executing hash-based incremental sync")

	// For hash-based sync, we need to compare source and destination
	// This is more complex and requires reading both sides
	gl.Log("info", "Hash-based sync implementation pending - falling back to full sync")
	return ism.executeFullSync()
}

// executeFullSync performs a complete synchronization
func (ism *IncrementalSyncManager) executeFullSync() error {
	gl.Log("info", "Executing full synchronization")

	// Reset the SQL query to original or default
	if ism.config.SQLQuery == "" {
		ism.config.SQLQuery = fmt.Sprintf("SELECT * FROM %s", ism.config.SourceTable)
	}

	return ism.executeSyncWithQuery()
}

// executeSyncWithQuery executes the actual sync operation
func (ism *IncrementalSyncManager) executeSyncWithQuery() error {
	// Import the ETL execution function from sql package
	// We'll need to modify this to use the existing ETL functions
	gl.Log("info", "Executing sync with configured query")

	// For now, we'll use a placeholder - we need to integrate with existing ExecuteETL
	// This will be implemented by calling the existing ETL functions with our modified config

	// TODO: Call existing ETL functions here
	// err := ExecuteETL(ism.config.ConfigPath, ism.config.OutputPath, ism.config.OutputFormat, ism.config.NeedCheck, ism.config.CheckMethod)

	ism.state.RecordsProcessed++
	ism.state.LastSyncTime = time.Now().Format(time.RFC3339)

	return ism.saveState()
}

// updateTimestampState updates the sync state with the latest timestamp
func (ism *IncrementalSyncManager) updateTimestampState() error {
	// Query the source database for the maximum timestamp value
	// This will be the starting point for the next sync

	// For now, use current time as placeholder
	ism.state.LastSyncValue = time.Now().Format(time.RFC3339)
	gl.Log("info", fmt.Sprintf("Updated timestamp state to: %v", ism.state.LastSyncValue))

	return ism.saveState()
}

// updatePrimaryKeyState updates the sync state with the latest primary key value
func (ism *IncrementalSyncManager) updatePrimaryKeyState() error {
	// Query the source database for the maximum primary key value
	// This will be the starting point for the next sync

	// For now, increment by 1 as placeholder
	if ism.state.LastSyncValue == nil {
		ism.state.LastSyncValue = 1
	} else {
		if pkValue, ok := ism.state.LastSyncValue.(float64); ok {
			ism.state.LastSyncValue = int(pkValue) + 1
		}
	}

	gl.Log("info", fmt.Sprintf("Updated primary key state to: %v", ism.state.LastSyncValue))

	return ism.saveState()
}

// loadState loads the sync state from file
func (ism *IncrementalSyncManager) loadState() error {
	if ism.config.IncrementalSync.StateFile == "" {
		return nil
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(ism.config.IncrementalSync.StateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Check if state file exists
	if _, err := os.Stat(ism.config.IncrementalSync.StateFile); os.IsNotExist(err) {
		gl.Log("info", "No existing state file found, starting fresh")
		return nil
	}

	// Read and parse state file
	data, err := os.ReadFile(ism.config.IncrementalSync.StateFile)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	if err := json.Unmarshal(data, &ism.state); err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}

	gl.Log("info", fmt.Sprintf("Loaded sync state from: %s", ism.config.IncrementalSync.StateFile))
	gl.Log("info", fmt.Sprintf("Last sync value: %v", ism.state.LastSyncValue))

	return nil
}

// saveState saves the current sync state to file
func (ism *IncrementalSyncManager) saveState() error {
	if ism.config.IncrementalSync.StateFile == "" {
		return nil
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(ism.config.IncrementalSync.StateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Marshal state to JSON
	data, err := json.MarshalIndent(ism.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to file
	if err := os.WriteFile(ism.config.IncrementalSync.StateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	gl.Log("info", fmt.Sprintf("Saved sync state to: %s", ism.config.IncrementalSync.StateFile))

	return nil
}

// generateRowHash generates a hash for a data row to detect changes
func generateRowHash(data Data) string {
	// Convert data to JSON and hash it
	jsonData, _ := json.Marshal(data)
	hash := md5.Sum(jsonData)
	return fmt.Sprintf("%x", hash)
}

// GetSyncStatistics returns current sync statistics
func (ism *IncrementalSyncManager) GetSyncStatistics() SyncState {
	return ism.state
}
