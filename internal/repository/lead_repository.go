package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nicolasantos1/CadastroLeads/internal/dto"
	"github.com/nicolasantos1/CadastroLeads/internal/model"
)

type LeadRepository interface {
	Create(ctx context.Context, req dto.CreateLeadRequest) (*model.Lead, error)
	GetByID(ctx context.Context, id int) (*model.Lead, error)
	GetAll(ctx context.Context) ([]model.Lead, error)
	Update(ctx context.Context, id int, req dto.UpdateLeadRequest) (*model.Lead, error)
	UpdateStatus(ctx context.Context, id int, req dto.UpdateStatusRequest) (*model.Lead, error)
	Delete(ctx context.Context, id int) error
	GetByEmail(ctx context.Context, email string) (*model.Lead, error)
}

type leadRepository struct {
	db *sql.DB
}

func NewLeadRepository(db *sql.DB) LeadRepository {
	return &leadRepository{db: db}
}

func (r *leadRepository) Create(ctx context.Context, req dto.CreateLeadRequest) (*model.Lead, error) {
	query := `INSERT INTO leads (name, email, phone, source, status) VALUES (?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query, req.Name, req.Email, req.Phone, req.Source, model.StatusNew)
	if err != nil {
		return nil, fmt.Errorf("failed to create lead: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return r.GetByID(ctx, int(id))
}

func (r *leadRepository) GetByID(ctx context.Context, id int) (*model.Lead, error) {
	query := `SELECT id, name, email, phone, source, status, created_at, updated_at FROM leads WHERE id = ?`

	var lead model.Lead
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&lead.ID, &lead.Name, &lead.Email, &lead.Phone,
		&lead.Source, &lead.Status, &lead.CreatedAt, &lead.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lead with id %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get lead: %w", err)
	}

	return &lead, nil
}

func (r *leadRepository) GetAll(ctx context.Context) ([]model.Lead, error) {
	query := `SELECT id, name, email, phone, source, status, created_at, updated_at FROM leads ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list leads: %w", err)
	}
	defer rows.Close()

	var leads []model.Lead
	for rows.Next() {
		var lead model.Lead
		err := rows.Scan(
			&lead.ID, &lead.Name, &lead.Email, &lead.Phone,
			&lead.Source, &lead.Status, &lead.CreatedAt, &lead.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lead: %w", err)
		}
		leads = append(leads, lead)
	}

	return leads, rows.Err()
}

func (r *leadRepository) Update(ctx context.Context, id int, req dto.UpdateLeadRequest) (*model.Lead, error) {
	query := `UPDATE leads SET name = ?, phone = ?, source = ?, status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, req.Name, req.Phone, req.Source, req.Status, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update lead: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("lead with id %d not found", id)
	}

	return r.GetByID(ctx, id)
}

func (r *leadRepository) UpdateStatus(ctx context.Context, id int, req dto.UpdateStatusRequest) (*model.Lead, error) {
	query := `UPDATE leads SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, req.Status, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update lead status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("lead with id %d not found", id)
	}

	return r.GetByID(ctx, id)
}

func (r *leadRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM leads WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete lead: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("lead with id %d not found", id)
	}

	return nil
}

func (r *leadRepository) GetByEmail(ctx context.Context, email string) (*model.Lead, error) {
	query := `SELECT id, name, email, phone, source, status, created_at, updated_at FROM leads WHERE email = ?`

	var lead model.Lead
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&lead.ID, &lead.Name, &lead.Email, &lead.Phone,
		&lead.Source, &lead.Status, &lead.CreatedAt, &lead.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lead with email %s not found", email)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get lead by email: %w", err)
	}

	return &lead, nil
}
