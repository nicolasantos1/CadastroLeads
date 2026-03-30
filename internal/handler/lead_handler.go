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