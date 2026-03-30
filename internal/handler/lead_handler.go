package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/nicolasantos1/CadastroLeads/internal/dto"
	"github.com/nicolasantos1/CadastroLeads/internal/repository"
	"github.com/nicolasantos1/CadastroLeads/internal/service"
)

type LeadHandler struct {
	service service.LeadService
}

func NewLeadHandler(service service.LeadService) *LeadHandler {
	return &LeadHandler{service: service}
}

func (h *LeadHandler) RegisterRoutes(app *fiber.App) {
	app.Post("/leads", h.CreateLead)
	app.Get("/leads", h.ListLeads)
	app.Get("/leads/:id", h.GetLeadByID)
	app.Put("/leads/:id", h.UpdateLead)
	app.Patch("/leads/:id/status", h.UpdateLeadStatus)
	app.Delete("/leads/:id", h.DeleteLead)
}

type successResponse struct {
	Data any `json:"data"`
}

type errorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func newErrorResponse(message string) errorResponse {
	var resp errorResponse
	resp.Error.Message = message
	return resp
}

type leadIDParam struct {
	ID int `uri:"id"`
}

type listLeadsQuery struct {
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
	Status string `query:"status"`
	Source string `query:"source"`
}

// CreateLead godoc
// @Summary Cria um lead
// @Description Cadastra um novo lead comercial
// @Tags leads
// @Accept json
// @Produce json
// @Param input body dto.CreateLeadRequest true "Dados do lead"
// @Success 201 {object} dto.LeadResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /leads [post]
func (h *LeadHandler) CreateLead(c fiber.Ctx) error {
	var req dto.CreateLeadRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(newErrorResponse(jsonErrorMessage(err)))
	}

	lead, err := h.service.Create(req)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(successResponse{
		Data: lead,
	})
}

// ListLeads godoc
// @Summary Lista leads
// @Description Lista leads com paginação e filtros
// @Tags leads
// @Produce json
// @Param page query int false "Página"
// @Param limit query int false "Limite por página"
// @Param status query string false "Filtro por status"
// @Param source query string false "Filtro por source"
// @Success 200 {object} dto.LeadListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /leads [get]
func (h *LeadHandler) ListLeads(c fiber.Ctx) error {
	var query listLeadsQuery
	if err := c.Bind().Query(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(newErrorResponse("query params inválidos"))
	}

	leads, err := h.service.List(query.Page, query.Limit, query.Status, query.Source)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(successResponse{
		Data: leads,
	})
}

// GetLeadByID godoc
// @Summary Busca lead por ID
// @Description Retorna um lead pelo ID
// @Tags leads
// @Produce json
// @Param id path int true "ID do lead"
// @Success 200 {object} dto.LeadResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /leads/{id} [get]
func (h *LeadHandler) GetLeadByID(c fiber.Ctx) error {
	var params leadIDParam
	if err := c.Bind().URI(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(newErrorResponse("id inválido"))
	}

	lead, err := h.service.GetByID(params.ID)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(successResponse{
		Data: lead,
	})
}


func decodeStrictJSON[T any](c fiber.Ctx, dst *T) error {
	dec := json.NewDecoder(bytes.NewReader(c.Body()))
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return err
	}

	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return errors.New("json inválido")
	}

	return nil
}

func jsonErrorMessage(err error) string {
	msg := err.Error()

	if field, found := strings.CutPrefix(msg, "json: unknown field "); found {
		return fmt.Sprintf("campo não permitido: %s", field)
	}

	return "corpo da requisição inválido"
}

// UpdateLead godoc
// @Summary Atualiza um lead
// @Description Atualiza os dados de um lead existente
// @Tags leads
// @Accept json
// @Produce json
// @Param id path int true "ID do lead"
// @Param input body dto.UpdateLeadRequest true "Dados atualizados do lead"
// @Success 200 {object} dto.LeadResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /leads/{id} [put]
func (h *LeadHandler) UpdateLead(c fiber.Ctx) error {
	var params leadIDParam
	if err := c.Bind().URI(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(newErrorResponse("id inválido"))
	}

	var req dto.UpdateLeadRequest
	if err := decodeStrictJSON(c, &req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(newErrorResponse(jsonErrorMessage(err)))
	}

	lead, err := h.service.Update(params.ID, req)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(successResponse{
		Data: lead,
	})
}

// UpdateLeadStatus godoc
// @Summary Atualiza status do lead
// @Description Atualiza apenas o status de um lead existente
// @Tags leads
// @Accept json
// @Produce json
// @Param id path int true "ID do lead"
// @Param input body dto.UpdateStatusRequest true "Novo status"
// @Success 200 {object} dto.LeadResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /leads/{id}/status [patch]
func (h *LeadHandler) UpdateLeadStatus(c fiber.Ctx) error {
	var params leadIDParam
	if err := c.Bind().URI(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(newErrorResponse("id inválido"))
	}

	var req dto.UpdateStatusRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(newErrorResponse(jsonErrorMessage(err)))
	}

	lead, err := h.service.UpdateStatus(params.ID, req)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(successResponse{
		Data: lead,
	})
}

// DeleteLead godoc
// @Summary Remove um lead
// @Description Remove um lead pelo ID
// @Tags leads
// @Produce json
// @Param id path int true "ID do lead"
// @Success 200 {object} dto.DeleteLeadResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /leads/{id} [delete]
func (h *LeadHandler) DeleteLead(c fiber.Ctx) error {
	var params leadIDParam
	if err := c.Bind().URI(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(newErrorResponse("id inválido"))
	}

	if err := h.service.Delete(params.ID); err != nil {
		return h.handleServiceError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(successResponse{
		Data: fiber.Map{
			"message": "lead removido com sucesso",
		},
	})
}

func (h *LeadHandler) handleServiceError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidName),
		errors.Is(err, service.ErrInvalidEmail),
		errors.Is(err, service.ErrInvalidSource),
		errors.Is(err, service.ErrInvalidStatus):
		return c.Status(fiber.StatusBadRequest).JSON(newErrorResponse(err.Error()))

	case errors.Is(err, service.ErrDuplicateEmail):
		return c.Status(fiber.StatusConflict).JSON(newErrorResponse(err.Error()))

	case errors.Is(err, repository.ErrLeadNotFound):
		return c.Status(fiber.StatusNotFound).JSON(newErrorResponse(err.Error()))

	default:
		return c.Status(fiber.StatusInternalServerError).JSON(newErrorResponse("erro interno do servidor"))
	}
}