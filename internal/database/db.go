package database

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func ConnectSQLite(dbPath string) (*sql.DB, error) {
	dbPath = strings.TrimSpace(dbPath)
	if dbPath == "" {
		dbPath = "leads.db"
	}

	dir := filepath.Dir(dbPath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("erro ao preparar diretório do banco: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir banco: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erro ao conectar no banco: %w", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("erro ao rodar migrations: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	upSQL, err := migrationFiles.ReadFile("migrations/000001_create_leads.up.sql")
	if err != nil {
		return fmt.Errorf("erro ao ler migration: %w", err)
	}

	if _, err := db.Exec(string(upSQL)); err != nil {
		return fmt.Errorf("erro ao aplicar migration: %w", err)
	}

	return nil
}