package database

import (
	"database/sql"
	"embed"
	"fmt"
	"strings"
)

//go:embed seeds/*.sql
var seedFiles embed.FS

func SeedFromFiles(db *sql.DB, runSeed bool) error {
	if !runSeed {
		return nil
	}

	var total int
	err := db.QueryRow(`SELECT COUNT(*) FROM leads`).Scan(&total)
	if err != nil {
		return fmt.Errorf("erro ao verificar se a tabela leads está vazia: %w", err)
	}

	if total > 0 {
		return nil
	}

	entries, err := seedFiles.ReadDir("seeds")
	if err != nil {
		return fmt.Errorf("erro ao listar pasta seeds: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		content, err := seedFiles.ReadFile("seeds/" + entry.Name())
		if err != nil {
			return fmt.Errorf("erro ao ler arquivo %s: %w", entry.Name(), err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("erro ao aplicar seed %s: %w", entry.Name(), err)
		}
	}

	return nil
}