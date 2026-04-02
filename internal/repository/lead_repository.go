package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nicolasantos1/CadastroLeads/internal/model"
)

var ErrLeadNotFound = errors.New("lead não encontrado")

type LeadRepository interface {
	Create(lead *model.Lead) error
	List(page, limit int, status, source string) ([]model.Lead, error)
	GetByID(id int) (*model.Lead, error)
	GetByEmail(email string) (*model.Lead, error)
	Update(lead *model.Lead) error
	UpdateStatus(id int, status string) error
	Delete(id int) error
}

type SQLiteLeadRepository struct {
	db *sql.DB
}

func NewLeadRepository(db *sql.DB) LeadRepository {
	return &SQLiteLeadRepository{db: db}
}

func (r *SQLiteLeadRepository) Create(lead *model.Lead) error {
	now := time.Now().UTC()

	if lead.Status == "" {
		lead.Status = model.StatusNew
	}

	result, err := r.db.Exec(`
		INSERT INTO leads (name, email, phone, source, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		lead.Name,
		lead.Email,
		lead.Phone,
		lead.Source,
		lead.Status,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("erro ao criar lead: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("erro ao obter id do lead criado: %w", err)
	}

	lead.ID = int(id)
	lead.CreatedAt = now
	lead.UpdatedAt = now

	return nil
}

func (r *SQLiteLeadRepository) List(page, limit int, status, source string) ([]model.Lead, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit

	query := `
		SELECT id, name, email, phone, source, status, created_at, updated_at
		FROM leads
		WHERE deleted_at IS NULL
	`
	var conditions []string
	var args []any

	if status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, status)
	}
	
	if source != "" {
		conditions = append(conditions, "source = ?")
		args = append(args, source)
	}
	
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar leads: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("erro ao fechar rows: %v", err)
		}
	}()

	var leads []model.Lead

	for rows.Next() {
		lead, err := scanLead(rows)
		if err != nil {
			return nil, err
		}
		leads = append(leads, *lead)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar leads: %w", err)
	}

	return leads, nil
}

func (r *SQLiteLeadRepository) GetByID(id int) (*model.Lead, error) {
	row := r.db.QueryRow(`
		SELECT id, name, email, phone, source, status, created_at, updated_at
		FROM leads
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	
	lead, err := scanLead(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLeadNotFound
		}
		return nil, err
	}

	return lead, nil
}

func (r *SQLiteLeadRepository) GetByEmail(email string) (*model.Lead, error) {
	row := r.db.QueryRow(`
		SELECT id, name, email, phone, source, status, created_at, updated_at
		FROM leads
		WHERE email = ? AND deleted_at IS NULL
	`, email)

	lead, err := scanLead(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return lead, nil
}

func (r *SQLiteLeadRepository) Update(lead *model.Lead) error {
	now := time.Now().UTC()

	result, err := r.db.Exec(`
		UPDATE leads
		SET name = ?, phone = ?, source = ?, status = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`,
		lead.Name,
		lead.Phone,
		lead.Source,
		lead.Status,
		now,
		lead.ID,
	)
	
	if err != nil {
		return fmt.Errorf("erro ao atualizar lead: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao verificar atualização do lead: %w", err)
	}

	if rowsAffected == 0 {
		return ErrLeadNotFound
	}

	lead.UpdatedAt = now
	return nil
}

func (r *SQLiteLeadRepository) UpdateStatus(id int, status string) error {
	result, err := r.db.Exec(`
		UPDATE leads
		SET status = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("erro ao atualizar status do lead: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao verificar atualização de status: %w", err)
	}

	if rowsAffected == 0 {
		return ErrLeadNotFound
	}

	return nil
}

func (r *SQLiteLeadRepository) Delete(id int) error {
	now := time.Now().UTC()

	result, err := r.db.Exec(`
		UPDATE leads
		SET deleted_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, now, now, id)
	if err != nil {
		return fmt.Errorf("erro ao deletar lead: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao verificar exclusão do lead: %w", err)
	}

	if rowsAffected == 0 {
		return ErrLeadNotFound
	}

	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanLead(s scanner) (*model.Lead, error) {
	var lead model.Lead
	var phone sql.NullString
	var createdAtRaw any
	var updatedAtRaw any

	err := s.Scan(
		&lead.ID,
		&lead.Name,
		&lead.Email,
		&phone,
		&lead.Source,
		&lead.Status,
		&createdAtRaw,
		&updatedAtRaw,
	)
	if err != nil {
		return nil, err
	}

	if phone.Valid {
		lead.Phone = phone.String
	}

	createdAt, err := parseSQLiteTime(createdAtRaw)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter created_at: %w", err)
	}

	updatedAt, err := parseSQLiteTime(updatedAtRaw)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter updated_at: %w", err)
	}

	lead.CreatedAt = createdAt
	lead.UpdatedAt = updatedAt

	return &lead, nil
}

func parseSQLiteTime(value any) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		return parseTimeString(v)
	case []byte:
		return parseTimeString(string(v))
	default:
		return time.Time{}, fmt.Errorf("tipo de data não suportado: %T", value)
	}
}

func parseTimeString(value string) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("formato de data inválido: %s", value)
}