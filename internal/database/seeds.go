// internal/database/seed.go
package database

import (
    "database/sql"
    "embed"
    "fmt"
    "strings"
)

//go:embed seeds/*.sql
var seedFiles embed.FS

// SeedFromFiles executa todos os arquivos .sql presentes em internal/database/seeds
// quando runSeed for true. Usa Exec para aplicar cada seed no banco.
func SeedFromFiles(db *sql.DB, runSeed bool) error {
    if !runSeed {
        return nil // não faz nada se o caller não quiser rodar os seeds
    }

    entries, err := seedFiles.ReadDir("seeds")
    if err != nil {
        return fmt.Errorf("erro ao listar pasta seeds: %w", err)
    }

    for _, entry := range entries {
        // ignorar subpastas ou arquivos não‑SQL
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