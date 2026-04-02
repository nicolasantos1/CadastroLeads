package model

type IdempotencyKey struct {
	ID                 int
	Key                string
	Method             string
	Path               string
	RequestHash        string
	Status             string
	ResponseStatusCode int
	ResponseBody       string
}

const (
	IdempotencyStatusProcessing = "processing"
	IdempotencyStatusCompleted  = "completed"
)