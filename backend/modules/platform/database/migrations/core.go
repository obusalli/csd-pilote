package migrations

import (
	"fmt"
	"reflect"
	"strings"

	"csd-pilote/backend/modules/pilot/clusters"
	"csd-pilote/backend/modules/pilot/containers"
	"csd-pilote/backend/modules/pilot/hypervisors"

	"gorm.io/gorm"
)

// DB is the GORM database instance used for migrations
// Must be set before calling AutoMigrateWithResult
var DB *gorm.DB

// Verbose controls whether detailed migration output is shown
var Verbose bool

// SchemaName is the database schema name
const SchemaName = "csd_pilote"

// MigrationResult contains statistics about the migration
type MigrationResult struct {
	TotalTables    int
	TotalCreated   int
	TotalUpdated   int
	TotalNoChange  int
	IndexesCreated int
	Groups         []GroupResult
}

// GroupResult contains statistics for a group of tables
type GroupResult struct {
	Name     string
	Tables   []TableResult
	Created  int
	Updated  int
	NoChange int
}

// TableResult contains the status of a single table
type TableResult struct {
	Name   string
	Status TableStatus
}

// TableStatus represents the migration status of a table
type TableStatus int

const (
	TableUnchanged TableStatus = iota
	TableCreated
	TableUpdated
)

// NewMigrationResult creates a new MigrationResult
func NewMigrationResult() *MigrationResult {
	return &MigrationResult{
		Groups: make([]GroupResult, 0),
	}
}

// AddGroup adds a group result to the migration result
func (r *MigrationResult) AddGroup(group GroupResult) {
	r.Groups = append(r.Groups, group)
	r.TotalTables += len(group.Tables)
	r.TotalCreated += group.Created
	r.TotalUpdated += group.Updated
	r.TotalNoChange += group.NoChange
}

// getTableName extracts the table name from a GORM model
func getTableName(db *gorm.DB, model interface{}) string {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(model); err != nil {
		t := reflect.TypeOf(model)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		return strings.ToLower(t.Name())
	}
	return stmt.Schema.Table
}

// tableExists checks if a table exists in the database
func tableExists(db *gorm.DB, tableName string) bool {
	schemaName := SchemaName
	table := tableName
	if strings.Contains(tableName, ".") {
		parts := strings.SplitN(tableName, ".", 2)
		schemaName = parts[0]
		table = parts[1]
	}

	var count int64
	db.Raw(`
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = ? AND table_name = ?
	`, schemaName, table).Scan(&count)
	return count > 0
}

// getColumnInfo returns a map of column names for a table
func getColumnInfo(db *gorm.DB, tableName string) map[string]bool {
	schemaName := SchemaName
	table := tableName
	if strings.Contains(tableName, ".") {
		parts := strings.SplitN(tableName, ".", 2)
		schemaName = parts[0]
		table = parts[1]
	}

	var columns []string
	db.Raw(`
		SELECT column_name FROM information_schema.columns
		WHERE table_schema = ? AND table_name = ?
	`, schemaName, table).Scan(&columns)

	result := make(map[string]bool)
	for _, col := range columns {
		result[col] = true
	}
	return result
}

// migrateGroup migrates a group of models and returns results
func migrateGroup(db *gorm.DB, groupName string, models []interface{}) (GroupResult, error) {
	result := GroupResult{
		Name:   groupName,
		Tables: make([]TableResult, 0, len(models)),
	}

	// Check which tables exist before migration
	existingTables := make(map[string]map[string]bool)
	for _, model := range models {
		tableName := getTableName(db, model)
		if tableExists(db, tableName) {
			existingTables[tableName] = getColumnInfo(db, tableName)
		}
	}

	// Run migration
	if err := db.AutoMigrate(models...); err != nil {
		return result, err
	}

	// Determine status of each table
	for _, model := range models {
		tableName := getTableName(db, model)
		tableResult := TableResult{Name: tableName}

		existingCols, existed := existingTables[tableName]
		if !existed {
			tableResult.Status = TableCreated
			result.Created++
		} else {
			newCols := getColumnInfo(db, tableName)
			hasNewColumns := false
			for col := range newCols {
				if !existingCols[col] {
					hasNewColumns = true
					break
				}
			}
			if hasNewColumns {
				tableResult.Status = TableUpdated
				result.Updated++
			} else {
				tableResult.Status = TableUnchanged
				result.NoChange++
			}
		}
		result.Tables = append(result.Tables, tableResult)
	}

	return result, nil
}

// logGroupResult logs the result of a group migration (only in verbose mode)
func logGroupResult(group GroupResult) {
	if !Verbose {
		return
	}
	fmt.Printf("• %s:\n", group.Name)
	for _, t := range group.Tables {
		switch t.Status {
		case TableCreated:
			fmt.Printf("    + %s (created)\n", t.Name)
		case TableUpdated:
			fmt.Printf("    ~ %s (updated)\n", t.Name)
		case TableUnchanged:
			fmt.Printf("    · %s\n", t.Name)
		}
	}
}

// EnsureSchemaExists creates the csd_pilote schema if it doesn't exist
func EnsureSchemaExists() error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}
	return DB.Exec("CREATE SCHEMA IF NOT EXISTS " + SchemaName).Error
}

// AutoMigrateWithResult runs GORM auto-migrations and returns detailed results
func AutoMigrateWithResult() (*MigrationResult, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not connected")
	}

	result := NewMigrationResult()

	if Verbose {
		fmt.Println()
		fmt.Println("═══════════════════════════════════════════════════════════")
		fmt.Println("  Database Migrations")
		fmt.Println("═══════════════════════════════════════════════════════════")
		fmt.Println()
		fmt.Println("• Ensuring database schema exists...")
	}

	if err := EnsureSchemaExists(); err != nil {
		return nil, fmt.Errorf("failed to ensure schema exists: %w", err)
	}

	// Kubernetes Clusters
	clusterModels := []interface{}{
		&clusters.Cluster{},
	}
	group, err := migrateGroup(DB, "Kubernetes Clusters", clusterModels)
	if err != nil {
		return nil, err
	}
	result.AddGroup(group)
	logGroupResult(group)

	// Libvirt Hypervisors
	hypervisorModels := []interface{}{
		&hypervisors.Hypervisor{},
	}
	group, err = migrateGroup(DB, "Libvirt Hypervisors", hypervisorModels)
	if err != nil {
		return nil, err
	}
	result.AddGroup(group)
	logGroupResult(group)

	// Container Engines
	containerModels := []interface{}{
		&containers.ContainerEngine{},
	}
	group, err = migrateGroup(DB, "Container Engines", containerModels)
	if err != nil {
		return nil, err
	}
	result.AddGroup(group)
	logGroupResult(group)

	// Create indexes
	if Verbose {
		fmt.Println("• Creating performance indexes...")
	}
	indexCount, err := createIndexes(DB)
	if err != nil {
		fmt.Printf("Warning: Failed to create some indexes: %v\n", err)
	}
	result.IndexesCreated = indexCount

	return result, nil
}

// indexExists checks if an index already exists in the database
func indexExists(db *gorm.DB, tableName, indexName string) bool {
	schemaName := SchemaName
	table := tableName
	if strings.Contains(tableName, ".") {
		parts := strings.SplitN(tableName, ".", 2)
		schemaName = parts[0]
		table = parts[1]
	}

	var count int64
	db.Raw(`
		SELECT COUNT(*) FROM pg_indexes
		WHERE schemaname = ? AND tablename = ? AND indexname = ?
	`, schemaName, table, indexName).Scan(&count)
	return count > 0
}

// createIndexes adds performance indexes
func createIndexes(db *gorm.DB) (int, error) {
	created := 0
	indexes := []struct {
		name  string
		table string
		expr  string
	}{
		// Clusters
		{"idx_clusters_name", SchemaName + ".clusters", "name"},
		{"idx_clusters_status", SchemaName + ".clusters", "status"},
		{"idx_clusters_tenant", SchemaName + ".clusters", "tenant_id"},

		// Hypervisors
		{"idx_hypervisors_name", SchemaName + ".hypervisors", "name"},
		{"idx_hypervisors_status", SchemaName + ".hypervisors", "status"},
		{"idx_hypervisors_tenant", SchemaName + ".hypervisors", "tenant_id"},

		// Container Engines
		{"idx_container_engines_name", SchemaName + ".container_engines", "name"},
		{"idx_container_engines_status", SchemaName + ".container_engines", "status"},
		{"idx_container_engines_tenant", SchemaName + ".container_engines", "tenant_id"},
		{"idx_container_engines_type", SchemaName + ".container_engines", "engine_type"},
	}

	for _, idx := range indexes {
		tableName := idx.table
		if strings.Contains(tableName, ".") {
			parts := strings.SplitN(tableName, ".", 2)
			tableName = parts[1]
		}

		if indexExists(db, tableName, idx.name) {
			continue
		}

		stmt := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", idx.name, idx.table, idx.expr)
		if err := db.Exec(stmt).Error; err != nil {
			if Verbose {
				fmt.Printf("    ! Failed to create index %s: %v\n", idx.name, err)
			}
		} else {
			created++
			if Verbose {
				fmt.Printf("    + %s\n", idx.name)
			}
		}
	}

	return created, nil
}
