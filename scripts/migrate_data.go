package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// Configuration for database connections
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// FileMigration represents a file record to migrate
type FileMigration struct {
	ID              int64
	CategoryID      int64
	CategoryName    string
	UploadUsername  string
	CatalogUsername string
	PutoutUsername  string
	Type            int
	Title           string
	Name            string
	Ext             string
	Size            int64
	Path            string
	Status          int
	Level           int
	Groups          string
	CatalogInfo     string
	UploadAt        sql.NullTime
	CatalogAt       sql.NullTime
	PutoutAt        sql.NullTime
	IsDownload      int
}

func main() {
	// Legacy database configuration (from legacy PHP system)
	legacyDB := DBConfig{
		Host:     os.Getenv("LEGACY_DB_HOST"),
		Port:     3306,
		User:     os.Getenv("LEGACY_DB_USER"),
		Password: os.Getenv("LEGACY_DB_PASSWORD"),
		Database: os.Getenv("LEGACY_DB_NAME"),
	}

	// New database configuration (Go application)
	newDB := DBConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     3306,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_NAME"),
	}

	if legacyDB.Host == "" || newDB.Host == "" {
		log.Fatal("Database configuration not set. Please set environment variables.")
	}

	// Connect to legacy database
	legacyDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		legacyDB.User, legacyDB.Password, legacyDB.Host, legacyDB.Port, legacyDB.Database)
	legacyConn, err := sql.Open("mysql", legacyDSN)
	if err != nil {
		log.Fatalf("Failed to connect to legacy database: %v", err)
	}
	defer legacyConn.Close()

	// Connect to new database
	newDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		newDB.User, newDB.Password, newDB.Host, newDB.Port, newDB.Database)
	newConn, err := sql.Open("mysql", newDSN)
	if err != nil {
		log.Fatalf("Failed to connect to new database: %v", err)
	}
	defer newConn.Close()

	log.Println("Starting data migration...")

	// Migrate users
	if err := migrateUsers(legacyConn, newConn); err != nil {
		log.Fatalf("Failed to migrate users: %v", err)
	}

	// Migrate groups
	if err := migrateGroups(legacyConn, newConn); err != nil {
		log.Fatalf("Failed to migrate groups: %v", err)
	}

	// Migrate roles
	if err := migrateRoles(legacyConn, newConn); err != nil {
		log.Fatalf("Failed to migrate roles: %v", err)
	}

	// Migrate permissions
	if err := migratePermissions(legacyConn, newConn); err != nil {
		log.Fatalf("Failed to migrate permissions: %v", err)
	}

	// Migrate levels
	if err := migrateLevels(legacyConn, newConn); err != nil {
		log.Fatalf("Failed to migrate levels: %v", err)
	}

	// Migrate categories
	if err := migrateCategories(legacyConn, newConn); err != nil {
		log.Fatalf("Failed to migrate categories: %v", err)
	}

	// Migrate catalog configuration
	if err := migrateCatalog(legacyConn, newConn); err != nil {
		log.Fatalf("Failed to migrate catalog: %v", err)
	}

	// Migrate files
	if err := migrateFiles(legacyConn, newConn); err != nil {
		log.Fatalf("Failed to migrate files: %v", err)
	}

	// Migrate relationships
	if err := migrateRelationships(legacyConn, newConn); err != nil {
		log.Fatalf("Failed to migrate relationships: %v", err)
	}

	log.Println("Data migration completed successfully!")
}

func migrateUsers(legacy, new *sql.DB) error {
	log.Println("Migrating users...")

	rows, err := legacy.Query("SELECT id, username, password, email, group_id, level_id, created_at, updated_at FROM ow_users")
	if err != nil {
		return fmt.Errorf("query legacy users: %w", err)
	}
	defer rows.Close()

	stmt, err := new.Prepare(`
		INSERT INTO ow_users (id, username, password, email, group_id, level_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			username = VALUES(username),
			password = VALUES(password),
			email = VALUES(email),
			group_id = VALUES(group_id),
			level_id = VALUES(level_id),
			updated_at = VALUES(updated_at)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for rows.Next() {
		var id, groupID, levelID int64
		var username, password, email string
		var createdAt, updatedAt sql.NullTime

		if err := rows.Scan(&id, &username, &password, &email, &groupID, &levelID, &createdAt, &updatedAt); err != nil {
			log.Printf("Warning: Failed to scan user row: %v", err)
			continue
		}

		if _, err := stmt.Exec(id, username, password, email, groupID, levelID, createdAt, updatedAt); err != nil {
			log.Printf("Warning: Failed to insert user %s: %v", username, err)
			continue
		}

		count++
	}

	log.Printf("Migrated %d users", count)
	return nil
}

func migrateGroups(legacy, new *sql.DB) error {
	log.Println("Migrating groups...")

	rows, err := legacy.Query("SELECT id, name, description, created_at, updated_at FROM ow_groups")
	if err != nil {
		return fmt.Errorf("query legacy groups: %w", err)
	}
	defer rows.Close()

	stmt, err := new.Prepare(`
		INSERT INTO ow_groups (id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			description = VALUES(description),
			updated_at = VALUES(updated_at)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for rows.Next() {
		var id int64
		var name, description string
		var createdAt, updatedAt sql.NullTime

		if err := rows.Scan(&id, &name, &description, &createdAt, &updatedAt); err != nil {
			log.Printf("Warning: Failed to scan group row: %v", err)
			continue
		}

		if _, err := stmt.Exec(id, name, description, createdAt, updatedAt); err != nil {
			log.Printf("Warning: Failed to insert group %s: %v", name, err)
			continue
		}

		count++
	}

	log.Printf("Migrated %d groups", count)
	return nil
}

func migrateRoles(legacy, new *sql.DB) error {
	log.Println("Migrating roles...")

	rows, err := legacy.Query("SELECT id, name, description, created_at, updated_at FROM ow_roles")
	if err != nil {
		return fmt.Errorf("query legacy roles: %w", err)
	}
	defer rows.Close()

	stmt, err := new.Prepare(`
		INSERT INTO ow_roles (id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			description = VALUES(description),
			updated_at = VALUES(updated_at)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for rows.Next() {
		var id int64
		var name, description string
		var createdAt, updatedAt sql.NullTime

		if err := rows.Scan(&id, &name, &description, &createdAt, &updatedAt); err != nil {
			log.Printf("Warning: Failed to scan role row: %v", err)
			continue
		}

		if _, err := stmt.Exec(id, name, description, createdAt, updatedAt); err != nil {
			log.Printf("Warning: Failed to insert role %s: %v", name, err)
			continue
		}

		count++
	}

	log.Printf("Migrated %d roles", count)
	return nil
}

func migratePermissions(legacy, new *sql.DB) error {
	log.Println("Migrating permissions...")

	rows, err := legacy.Query("SELECT id, name, description, resource, action, created_at, updated_at FROM ow_permissions")
	if err != nil {
		return fmt.Errorf("query legacy permissions: %w", err)
	}
	defer rows.Close()

	stmt, err := new.Prepare(`
		INSERT INTO ow_permissions (id, name, description, resource, action, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			description = VALUES(description),
			resource = VALUES(resource),
			action = VALUES(action),
			updated_at = VALUES(updated_at)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for rows.Next() {
		var id int64
		var name, description, resource, action string
		var createdAt, updatedAt sql.NullTime

		if err := rows.Scan(&id, &name, &description, &resource, &action, &createdAt, &updatedAt); err != nil {
			log.Printf("Warning: Failed to scan permission row: %v", err)
			continue
		}

		if _, err := stmt.Exec(id, name, description, resource, action, createdAt, updatedAt); err != nil {
			log.Printf("Warning: Failed to insert permission %s: %v", name, err)
			continue
		}

		count++
	}

	log.Printf("Migrated %d permissions", count)
	return nil
}

func migrateLevels(legacy, new *sql.DB) error {
	log.Println("Migrating levels...")

	rows, err := legacy.Query("SELECT id, name, level, description, created_at, updated_at FROM ow_levels")
	if err != nil {
		return fmt.Errorf("query legacy levels: %w", err)
	}
	defer rows.Close()

	stmt, err := new.Prepare(`
		INSERT INTO ow_levels (id, name, level, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			level = VALUES(level),
			description = VALUES(description),
			updated_at = VALUES(updated_at)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for rows.Next() {
		var id, level int64
		var name, description string
		var createdAt, updatedAt sql.NullTime

		if err := rows.Scan(&id, &name, &level, &description, &createdAt, &updatedAt); err != nil {
			log.Printf("Warning: Failed to scan level row: %v", err)
			continue
		}

		if _, err := stmt.Exec(id, name, level, description, createdAt, updatedAt); err != nil {
			log.Printf("Warning: Failed to insert level %s: %v", name, err)
			continue
		}

		count++
	}

	log.Printf("Migrated %d levels", count)
	return nil
}

func migrateCategories(legacy, new *sql.DB) error {
	log.Println("Migrating categories...")

	rows, err := legacy.Query(`
		SELECT id, parent_id, name, description, path, weight, status, 
		       level, groups, created_at, updated_at 
		FROM ow_categories
	`)
	if err != nil {
		return fmt.Errorf("query legacy categories: %w", err)
	}
	defer rows.Close()

	stmt, err := new.Prepare(`
		INSERT INTO ow_categories (id, parent_id, name, description, path, weight, status, level, groups, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			parent_id = VALUES(parent_id),
			name = VALUES(name),
			description = VALUES(description),
			path = VALUES(path),
			weight = VALUES(weight),
			status = VALUES(status),
			level = VALUES(level),
			groups = VALUES(groups),
			updated_at = VALUES(updated_at)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for rows.Next() {
		var id, parentID, weight, status, level int64
		var name, description, path, groups string
		var createdAt, updatedAt sql.NullTime

		if err := rows.Scan(&id, &parentID, &name, &description, &path, &weight, &status, &level, &groups, &createdAt, &updatedAt); err != nil {
			log.Printf("Warning: Failed to scan category row: %v", err)
			continue
		}

		if _, err := stmt.Exec(id, parentID, name, description, path, weight, status, level, groups, createdAt, updatedAt); err != nil {
			log.Printf("Warning: Failed to insert category %s: %v", name, err)
			continue
		}

		count++
	}

	log.Printf("Migrated %d categories", count)
	return nil
}

func migrateCatalog(legacy, new *sql.DB) error {
	log.Println("Migrating catalog configuration...")

	rows, err := legacy.Query(`
		SELECT id, parent_id, name, label, type, path, weight, enabled, created_at, updated_at 
		FROM ow_catalog
	`)
	if err != nil {
		return fmt.Errorf("query legacy catalog: %w", err)
	}
	defer rows.Close()

	stmt, err := new.Prepare(`
		INSERT INTO ow_catalog (id, parent_id, name, label, type, path, weight, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			parent_id = VALUES(parent_id),
			name = VALUES(name),
			label = VALUES(label),
			type = VALUES(type),
			path = VALUES(path),
			weight = VALUES(weight),
			enabled = VALUES(enabled),
			updated_at = VALUES(updated_at)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for rows.Next() {
		var id, parentID, fieldType, weight, enabled int64
		var name, label, path string
		var createdAt, updatedAt sql.NullTime

		if err := rows.Scan(&id, &parentID, &name, &label, &fieldType, &path, &weight, &enabled, &createdAt, &updatedAt); err != nil {
			log.Printf("Warning: Failed to scan catalog row: %v", err)
			continue
		}

		if _, err := stmt.Exec(id, parentID, name, label, fieldType, path, weight, enabled, createdAt, updatedAt); err != nil {
			log.Printf("Warning: Failed to insert catalog %s: %v", name, err)
			continue
		}

		count++
	}

	log.Printf("Migrated %d catalog fields", count)
	return nil
}

func migrateFiles(legacy, new *sql.DB) error {
	log.Println("Migrating files...")

	rows, err := legacy.Query(`
		SELECT id, category_id, category_name, upload_username, catalog_username, putout_username,
		       type, title, name, ext, size, path, status, level, groups, catalog_info,
		       upload_at, catalog_at, putout_at, is_download
		FROM ow_files
	`)
	if err != nil {
		return fmt.Errorf("query legacy files: %w", err)
	}
	defer rows.Close()

	stmt, err := new.Prepare(`
		INSERT INTO ow_files (id, category_id, category_name, upload_username, catalog_username, putout_username,
		                      type, title, name, ext, size, path, status, level, groups, catalog_info,
		                      upload_at, catalog_at, putout_at, is_download, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			category_id = VALUES(category_id),
			category_name = VALUES(category_name),
			catalog_username = VALUES(catalog_username),
			putout_username = VALUES(putout_username),
			title = VALUES(title),
			status = VALUES(status),
			catalog_info = VALUES(catalog_info),
			catalog_at = VALUES(catalog_at),
			putout_at = VALUES(putout_at),
			updated_at = NOW()
	`)
	if err != nil {
		return fmt.Errorf("prepare insert statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for rows.Next() {
		var file FileMigration

		if err := rows.Scan(
			&file.ID, &file.CategoryID, &file.CategoryName,
			&file.UploadUsername, &file.CatalogUsername, &file.PutoutUsername,
			&file.Type, &file.Title, &file.Name, &file.Ext, &file.Size, &file.Path,
			&file.Status, &file.Level, &file.Groups, &file.CatalogInfo,
			&file.UploadAt, &file.CatalogAt, &file.PutoutAt, &file.IsDownload,
		); err != nil {
			log.Printf("Warning: Failed to scan file row: %v", err)
			continue
		}

		// Validate catalog_info is valid JSON
		if file.CatalogInfo != "" {
			var jsonData interface{}
			if err := json.Unmarshal([]byte(file.CatalogInfo), &jsonData); err != nil {
				log.Printf("Warning: Invalid JSON in catalog_info for file %d: %v", file.ID, err)
				file.CatalogInfo = "{}"
			}
		} else {
			file.CatalogInfo = "{}"
		}

		if _, err := stmt.Exec(
			file.ID, file.CategoryID, file.CategoryName,
			file.UploadUsername, file.CatalogUsername, file.PutoutUsername,
			file.Type, file.Title, file.Name, file.Ext, file.Size, file.Path,
			file.Status, file.Level, file.Groups, file.CatalogInfo,
			file.UploadAt, file.CatalogAt, file.PutoutAt, file.IsDownload,
		); err != nil {
			log.Printf("Warning: Failed to insert file %s: %v", file.Title, err)
			continue
		}

		count++
		if count%100 == 0 {
			log.Printf("Migrated %d files...", count)
		}
	}

	log.Printf("Migrated %d files total", count)
	return nil
}

func migrateRelationships(legacy, new *sql.DB) error {
	log.Println("Migrating relationships...")

	// Migrate groups_has_roles
	if err := migrateTable(legacy, new, "ow_groups_has_roles", "group_id, role_id"); err != nil {
		return fmt.Errorf("migrate groups_has_roles: %w", err)
	}

	// Migrate roles_has_permissions
	if err := migrateTable(legacy, new, "ow_roles_has_permissions", "role_id, permission_id"); err != nil {
		return fmt.Errorf("migrate roles_has_permissions: %w", err)
	}

	// Migrate groups_has_category
	if err := migrateTable(legacy, new, "ow_groups_has_category", "group_id, category_id"); err != nil {
		return fmt.Errorf("migrate groups_has_category: %w", err)
	}

	return nil
}

func migrateTable(legacy, new *sql.DB, tableName, columns string) error {
	log.Printf("Migrating %s...", tableName)

	rows, err := legacy.Query(fmt.Sprintf("SELECT %s FROM %s", columns, tableName))
	if err != nil {
		return fmt.Errorf("query legacy %s: %w", tableName, err)
	}
	defer rows.Close()

	columnList := strings.Split(columns, ", ")
	placeholders := strings.Repeat("?,", len(columnList))
	placeholders = placeholders[:len(placeholders)-1]

	stmt, err := new.Prepare(fmt.Sprintf(
		"INSERT IGNORE INTO %s (%s) VALUES (%s)",
		tableName, columns, placeholders,
	))
	if err != nil {
		return fmt.Errorf("prepare insert statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for rows.Next() {
		values := make([]interface{}, len(columnList))
		valuePtrs := make([]interface{}, len(columnList))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			log.Printf("Warning: Failed to scan row: %v", err)
			continue
		}

		if _, err := stmt.Exec(values...); err != nil {
			log.Printf("Warning: Failed to insert into %s: %v", tableName, err)
			continue
		}

		count++
	}

	log.Printf("Migrated %d rows in %s", count, tableName)
	return nil
}
