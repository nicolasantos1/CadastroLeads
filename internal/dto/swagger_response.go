// internal/dto/swagger_response.go
package dto

import "github.com/nicolasantos1/CadastroLeads/internal/model"

type ErrorBody struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type LeadResponse struct {
	Data model.Lead `json:"data"`
}

type LeadListResponse struct {
	Data []model.Lead `json:"data"`
}

type MessageBody struct {
	Message string `json:"message"`
}

type DeleteLeadResponse struct {
	Data MessageBody `json:"data"`
}

type HealthResponse struct {
	Data string `json:"data"`
}