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

	if err := ensureDeletedAtColumn(db); err != nil {
		return fmt.Errorf("erro ao garantir coluna deleted_at: %w", err)
	}

	return nil
}

func ensureDeletedAtColumn(db *sql.DB) error {
	rows, err := db.Query(`PRAGMA table_info(leads)`)
	if err != nil {
		return fmt.Errorf("erro ao inspecionar tabela leads: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("erro ao fechar rows: %v\n", closeErr)
		}
	}()

	var (
		cid        int 	
		name       string
		columnType string
		notNull    int
		defaultVal any
		pk         int
	)

	for rows.Next() {
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultVal, &pk); err != nil {
			return fmt.Errorf("erro ao ler schema da tabela leads: %w", err)
		}

		if name == "deleted_at" {
			return nil
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("erro ao iterar schema da tabela leads: %w", err)
	}

	if _, err := db.Exec(`ALTER TABLE leads ADD COLUMN deleted_at DATETIME NULL`); err != nil {
		return fmt.Errorf("erro ao adicionar coluna deleted_at: %w", err)
	}

	return nil
}