package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type Migration struct {
	Version   string
	UpFile    string
	DownFile  string
	UpSQL     string
	DownSQL   string
}

func main() {
	// Parse command line flags
	action := flag.String("action", "up", "Migration action: up, down, status")
	host := flag.String("host", "localhost", "Database host")
	port := flag.Int("port", 3306, "Database port")
	database := flag.String("database", "openwan_db", "Database name")
	username := flag.String("username", "root", "Database username")
	password := flag.String("password", "", "Database password")
	migrationsDir := flag.String("migrations", "../migrations", "Migrations directory")
	flag.Parse()

	// Connect to database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
		*username, *password, *host, *port, *database)
	
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Printf("Connected to database: %s@%s:%d/%s", *username, *host, *port, *database)

	// Load migrations
	migrations, err := loadMigrations(*migrationsDir)
	if err != nil {
		log.Fatalf("Failed to load migrations: %v", err)
	}

	// Execute migration action
	switch *action {
	case "up":
		if err := runMigrationsUp(db, migrations); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		log.Println("Migration up completed successfully")
	case "down":
		if err := runMigrationsDown(db, migrations); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Println("Migration down completed successfully")
	case "status":
		showMigrationStatus(migrations)
	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}

func loadMigrations(dir string) ([]Migration, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	migrationsMap := make(map[string]*Migration)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		// Parse filename: 000001_init_schema.up.sql or 000001_init_schema.down.sql
		parts := strings.Split(name, ".")
		if len(parts) < 3 {
			continue
		}

		version := strings.Split(parts[0], "_")[0]
		direction := parts[len(parts)-2]

		if _, exists := migrationsMap[version]; !exists {
			migrationsMap[version] = &Migration{Version: version}
		}

		fullPath := filepath.Join(dir, name)
		content, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return nil, err
		}

		if direction == "up" {
			migrationsMap[version].UpFile = name
			migrationsMap[version].UpSQL = string(content)
		} else if direction == "down" {
			migrationsMap[version].DownFile = name
			migrationsMap[version].DownSQL = string(content)
		}
	}

	// Convert map to sorted slice
	var migrations []Migration
	for _, m := range migrationsMap {
		migrations = append(migrations, *m)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func runMigrationsUp(db *sql.DB, migrations []Migration) error {
	for _, m := range migrations {
		log.Printf("Running migration %s: %s", m.Version, m.UpFile)
		
		if _, err := db.Exec(m.UpSQL); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", m.Version, err)
		}
		
		log.Printf("✓ Migration %s completed", m.Version)
	}
	return nil
}

func runMigrationsDown(db *sql.DB, migrations []Migration) error {
	// Run migrations in reverse order
	for i := len(migrations) - 1; i >= 0; i-- {
		m := migrations[i]
		log.Printf("Rolling back migration %s: %s", m.Version, m.DownFile)
		
		if _, err := db.Exec(m.DownSQL); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", m.Version, err)
		}
		
		log.Printf("✓ Migration %s rolled back", m.Version)
	}
	return nil
}

func showMigrationStatus(migrations []Migration) {
	fmt.Println("\nMigration Status:")
	fmt.Println("================")
	for _, m := range migrations {
		fmt.Printf("Version: %s\n", m.Version)
		fmt.Printf("  Up:   %s\n", m.UpFile)
		fmt.Printf("  Down: %s\n", m.DownFile)
	}
	fmt.Println()
}
