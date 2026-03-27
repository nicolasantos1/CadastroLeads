package service

import (
	"errors"
	"net/mail"
	"strings"

	"github.com/nicolasantos1/CadastroLeads/internal/dto"
	"github.com/nicolasantos1/CadastroLeads/internal/model"
	"github.com/nicolasantos1/CadastroLeads/internal/repository"
)

var (
	ErrInvalidName   = errors.New("name é obrigatório")
	ErrInvalidEmail  = errors.New("email inválido")
	ErrInvalidSource = errors.New("source é obrigatório")
	ErrInvalidStatus = errors.New("status inválido")
	ErrDuplicateEmail = errors.New("já existe um lead com este email")
)

type LeadService interface {
	Create(req dto.CreateLeadRequest) (*model.Lead, error)
	List(page, limit int, status, source string) ([]model.Lead, error)
	GetByID(id int) (*model.Lead, error)
	Update(id int, req dto.UpdateLeadRequest) (*model.Lead, error)
	UpdateStatus(id int, req dto.UpdateStatusRequest) (*model.Lead, error)
	Delete(id int) error
}

type leadService struct {
	repo repository.LeadRepository
}

func NewLeadService(repo repository.LeadRepository) LeadService {
	return &leadService{repo: repo}
}

func (s *leadService) Create(req dto.CreateLeadRequest) (*model.Lead, error) {
	name := strings.TrimSpace(req.Name)
	email := normalizeEmail(req.Email)
	phone := strings.TrimSpace(req.Phone)
	source := strings.TrimSpace(req.Source)

	if name == "" {
		return nil, ErrInvalidName
	}

	if source == "" {
		return nil, ErrInvalidSource
	}

	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	existingLead, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if existingLead != nil {
		return nil, ErrDuplicateEmail
	}

	lead := &model.Lead{
		Name:   name,
		Email:  email,
		Phone:  phone,
		Source: source,
		Status: model.StatusNew,
	}

	if err := s.repo.Create(lead); err != nil {
		return nil, err
	}

	return lead, nil
}

func (s *leadService) List(page, limit int, status, source string) ([]model.Lead, error) {
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	status = strings.TrimSpace(status)
	source = strings.TrimSpace(source)

	if status != "" && !model.IsValidStatus(status) {
		return nil, ErrInvalidStatus
	}

	return s.repo.List(page, limit, status, source)
}

func (s *leadService) GetByID(id int) (*model.Lead, error) {
	if id <= 0 {
		return nil, repository.ErrLeadNotFound
	}

	return s.repo.GetByID(id)
}

func (s *leadService) Update(id int, req dto.UpdateLeadRequest) (*model.Lead, error) {
	if id <= 0 {
		return nil, repository.ErrLeadNotFound
	}

	name := strings.TrimSpace(req.Name)
	phone := strings.TrimSpace(req.Phone)
	source := strings.TrimSpace(req.Source)
	status := strings.TrimSpace(req.Status)

	if name == "" {
		return nil, ErrInvalidName
	}

	if source == "" {
		return nil, ErrInvalidSource
	}

	if !model.IsValidStatus(status) {
		return nil, ErrInvalidStatus
	}

	existingLead, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	existingLead.Name = name
	existingLead.Phone = phone
	existingLead.Source = source
	existingLead.Status = status

	if err := s.repo.Update(existingLead); err != nil {
		return nil, err
	}

	updatedLead, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return updatedLead, nil
}

func (s *leadService) UpdateStatus(id int, req dto.UpdateStatusRequest) (*model.Lead, error) {
	if id <= 0 {
		return nil, repository.ErrLeadNotFound
	}

	status := strings.TrimSpace(req.Status)
	if !model.IsValidStatus(status) {
		return nil, ErrInvalidStatus
	}

	if _, err := s.repo.GetByID(id); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateStatus(id, status); err != nil {
		return nil, err
	}

	updatedLead, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return updatedLead, nil
}

func (s *leadService) Delete(id int) error {
	if id <= 0 {
		return repository.ErrLeadNotFound
	}

	return s.repo.Delete(id)
}

func isValidEmail(email string) bool {
	if email == "" {
		return false
	}

	_, err := mail.ParseAddress(email)
	return err == nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}