package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/nicolasantos1/CadastroLeads/internal/model"
)

type IdempotencyRepository interface {
	Get(key, method, path string) (*model.IdempotencyKey, error)
	Reserve(key, method, path, requestHash string) (bool, error)
	Complete(key, method, path string, statusCode int, responseBody string) error
}

type SQLiteIdempotencyRepository struct {
	db *sql.DB
}

func NewIdempotencyRepository(db *sql.DB) IdempotencyRepository {
	return &SQLiteIdempotencyRepository{db: db}
}

func (r *SQLiteIdempotencyRepository) Get(key, method, path string) (*model.IdempotencyKey, error) {
	row := r.db.QueryRow(`
		SELECT id, idempotency_key, method, path, request_hash, status,
		       COALESCE(response_status_code, 0),
		       COALESCE(response_body, '')
		FROM idempotency_keys
		WHERE idempotency_key = ? AND method = ? AND path = ?
	`, key, method, path)

	var record model.IdempotencyKey
	err := row.Scan(
		&record.ID,
		&record.Key,
		&record.Method,
		&record.Path,
		&record.RequestHash,
		&record.Status,
		&record.ResponseStatusCode,
		&record.ResponseBody,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar chave idempotente: %w", err)
	}

	return &record, nil
}

func (r *SQLiteIdempotencyRepository) Reserve(key, method, path, requestHash string) (bool, error) {
	now := time.Now().UTC()

	result, err := r.db.Exec(`
		INSERT OR IGNORE INTO idempotency_keys (
			idempotency_key, method, path, request_hash, status, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		key,
		method,
		path,
		requestHash,
		model.IdempotencyStatusProcessing,
		now,
		now,
	)
	if err != nil {
		return false, fmt.Errorf("erro ao reservar chave idempotente: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("erro ao verificar reserva da chave idempotente: %w", err)
	}

	return rowsAffected == 1, nil
}

func (r *SQLiteIdempotencyRepository) Complete(key, method, path string, statusCode int, responseBody string) error {
	result, err := r.db.Exec(`
		UPDATE idempotency_keys
		SET status = ?, response_status_code = ?, response_body = ?, updated_at = ?
		WHERE idempotency_key = ? AND method = ? AND path = ?
	`,
		model.IdempotencyStatusCompleted,
		statusCode,
		responseBody,
		time.Now().UTC(),
		key,
		method,
		path,
	)
	if err != nil {
		return fmt.Errorf("erro ao finalizar chave idempotente: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao verificar finalização da chave idempotente: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("chave idempotente não encontrada para finalização")
	}

	return nil
}