package dto

type CreateLeadRequest struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	Source string `json:"source"`
}