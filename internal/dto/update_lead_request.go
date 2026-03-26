package dto

type UpdateLeadRequest struct {
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Source string `json:"source"`
	Status string `json:"status"`
}