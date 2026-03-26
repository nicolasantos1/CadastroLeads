package model

import "time"

const (
	StatusNew       = "new"
	StatusContacted = "contacted"
	StatusQualified = "qualified"
	StatusLost      = "lost"
)

type Lead struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Source    string    `json:"source"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func IsValidStatus(status string) bool {
	switch status {
	case StatusNew, StatusContacted, StatusQualified, StatusLost:
		return true
	default:
		return false
	}
}