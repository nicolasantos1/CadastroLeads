package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	
	"github.com/gofiber/fiber/v3"
	_ "modernc.org/sqlite"

	"github.com/nicolasantos1/CadastroLeads/internal/handler"
	"github.com/nicolasantos1/CadastroLeads/internal/repository"
	"github.com/nicolasantos1/CadastroLeads/internal/service"
)

const testSchema = `
CREATE TABLE IF NOT EXISTS leads (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	phone TEXT,
	source TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'new',
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`
const testToken = "dev-token-123"
func setupTestApp(t *testing.T) *fiber.App {
	
	t.Helper()
	t.Setenv("API_TOKEN", testToken)	
	
	dbPath := filepath.Join(t.TempDir(), "test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("erro ao abrir banco de teste: %v", err)
	}

	if _, err := db.Exec(testSchema); err != nil {
		t.Fatalf("erro ao criar schema de teste: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	app := fiber.New()

	leadRepo := repository.NewLeadRepository(db)
	leadService := service.NewLeadService(leadRepo)
	leadHandler := handler.NewLeadHandler(leadService)

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"data": "API rodando com banco",
		})
	})

	leadHandler.RegisterRoutes(app)

	return app
}

func performRequest(t *testing.T, app *fiber.App, method, url string, body []byte) *http.Response {
	t.Helper()

	req := httptest.NewRequest(method, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro ao executar request de teste: %v", err)
	}

	return resp
}

func decodeJSONResponse(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()

	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("erro ao decodificar resposta JSON: %v", err)
	}

	return data
}

func TestCreateLeadSuccess(t *testing.T) {
	app := setupTestApp(t)

	body := []byte(`{
		"name":"Nicolas",
		"email":"nicolas@email.com",
		"phone":"11999999999",
		"source":"landing_page"
	}`)

	resp := performRequest(t, app, http.MethodPost, "/leads", body)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status esperado %d, recebido %d", http.StatusCreated, resp.StatusCode)
	}

	payload := decodeJSONResponse(t, resp)

	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("campo data ausente ou inválido")
	}

	if data["name"] != "Nicolas" {
		t.Fatalf("name esperado Nicolas, recebido %v", data["name"])
	}

	if data["email"] != "nicolas@email.com" {
		t.Fatalf("email esperado nicolas@email.com, recebido %v", data["email"])
	}

	if data["status"] != "new" {
		t.Fatalf("status esperado new, recebido %v", data["status"])
	}
}

func TestCreateLeadDuplicateEmail(t *testing.T) {
	app := setupTestApp(t)

	firstBody := []byte(`{
		"name":"Nicolas",
		"email":"nicolas@email.com",
		"phone":"11999999999",
		"source":"landing_page"
	}`)

	secondBody := []byte(`{
		"name":"Outro Nome",
		"email":"nicolas@email.com",
		"phone":"11888888888",
		"source":"google_ads"
	}`)

	firstResp := performRequest(t, app, http.MethodPost, "/leads", firstBody)
	if firstResp.StatusCode != http.StatusCreated {
		t.Fatalf("primeira criação falhou com status %d", firstResp.StatusCode)
	}

	secondResp := performRequest(t, app, http.MethodPost, "/leads", secondBody)

	if secondResp.StatusCode != http.StatusConflict {
		t.Fatalf("status esperado %d, recebido %d", http.StatusConflict, secondResp.StatusCode)
	}

	payload := decodeJSONResponse(t, secondResp)
	errorField, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("campo error ausente ou inválido")
	}

	if errorField["message"] != "já existe um lead com este email" {
		t.Fatalf("mensagem inesperada: %v", errorField["message"])
	}
}

func TestUpdateLeadStatusInvalid(t *testing.T) {
	app := setupTestApp(t)

	createBody := []byte(`{
		"name":"Nicolas",
		"email":"nicolas@email.com",
		"phone":"11999999999",
		"source":"landing_page"
	}`)

	createResp := performRequest(t, app, http.MethodPost, "/leads", createBody)
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("criação falhou com status %d", createResp.StatusCode)
	}

	updateBody := []byte(`{
		"status":"status_invalido"
	}`)

	resp := performRequest(t, app, http.MethodPatch, "/leads/1/status", updateBody)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status esperado %d, recebido %d", http.StatusBadRequest, resp.StatusCode)
	}

	payload := decodeJSONResponse(t, resp)
	errorField, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("campo error ausente ou inválido")
	}

	if errorField["message"] != "status inválido" {
		t.Fatalf("mensagem inesperada: %v", errorField["message"])
	}
}

func TestListLeadsWithStatusFilter(t *testing.T) {
	app := setupTestApp(t)

	lead1 := []byte(`{
		"name":"Lead 1",
		"email":"lead1@email.com",
		"phone":"11111111111",
		"source":"landing_page"
	}`)
	lead2 := []byte(`{
		"name":"Lead 2",
		"email":"lead2@email.com",
		"phone":"22222222222",
		"source":"google_ads"
	}`)

	resp1 := performRequest(t, app, http.MethodPost, "/leads", lead1)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("criação do lead1 falhou com status %d", resp1.StatusCode)
	}

	resp2 := performRequest(t, app, http.MethodPost, "/leads", lead2)
	if resp2.StatusCode != http.StatusCreated {
		t.Fatalf("criação do lead2 falhou com status %d", resp2.StatusCode)
	}

	updateBody := []byte(`{"status":"qualified"}`)
	updateResp := performRequest(t, app, http.MethodPatch, "/leads/1/status", updateBody)
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("update de status falhou com status %d", updateResp.StatusCode)
	}

	resp := performRequest(t, app, http.MethodGet, "/leads?status=qualified", nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status esperado %d, recebido %d", http.StatusOK, resp.StatusCode)
	}

	payload := decodeJSONResponse(t, resp)

	items, ok := payload["data"].([]any)
	if !ok {
		t.Fatalf("campo data não é uma lista válida")
	}

	if len(items) != 1 {
		t.Fatalf("esperado 1 lead filtrado, recebido %d", len(items))
	}

	lead, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("item da lista inválido")
	}

	if lead["status"] != "qualified" {
		t.Fatalf("status esperado qualified, recebido %v", lead["status"])
	}
}

func TestDeleteLeadThenGetByID(t *testing.T) {
	app := setupTestApp(t)

	createBody := []byte(`{
		"name":"Nicolas",
		"email":"nicolas@email.com",
		"phone":"11999999999",
		"source":"landing_page"
	}`)

	createResp := performRequest(t, app, http.MethodPost, "/leads", createBody)
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("criação falhou com status %d", createResp.StatusCode)
	}

	deleteResp := performRequest(t, app, http.MethodDelete, "/leads/1", nil)
	if deleteResp.StatusCode != http.StatusOK {
		t.Fatalf("delete falhou com status %d", deleteResp.StatusCode)
	}

	getResp := performRequest(t, app, http.MethodGet, "/leads/1", nil)
	if getResp.StatusCode != http.StatusNotFound {
		t.Fatalf("status esperado %d, recebido %d", http.StatusNotFound, getResp.StatusCode)
	}
}

func TestCreateLeadInvalidEmail(t *testing.T) {
	app := setupTestApp(t)

	body := []byte(`{
		"name":"Nicolas",
		"email":"email_invalido",
		"phone":"11999999999",
		"source":"landing_page"
	}`)

	resp := performRequest(t, app, http.MethodPost, "/leads", body)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status esperado %d, recebido %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestGetLeadByIDNotFound(t *testing.T) {
	app := setupTestApp(t)

	resp := performRequest(t, app, http.MethodGet, "/leads/999", nil)

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status esperado %d, recebido %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestListLeadsWithSourceFilter(t *testing.T) {
	app := setupTestApp(t)

	lead1 := []byte(`{
		"name":"Lead 1",
		"email":"lead1@email.com",
		"phone":"11111111111",
		"source":"landing_page"
	}`)
	lead2 := []byte(`{
		"name":"Lead 2",
		"email":"lead2@email.com",
		"phone":"22222222222",
		"source":"google_ads"
	}`)

	resp1 := performRequest(t, app, http.MethodPost, "/leads", lead1)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("criação do lead1 falhou com status %d", resp1.StatusCode)
	}

	resp2 := performRequest(t, app, http.MethodPost, "/leads", lead2)
	if resp2.StatusCode != http.StatusCreated {
		t.Fatalf("criação do lead2 falhou com status %d", resp2.StatusCode)
	}

	resp := performRequest(t, app, http.MethodGet, "/leads?source=google_ads", nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status esperado %d, recebido %d", http.StatusOK, resp.StatusCode)
	}

	payload := decodeJSONResponse(t, resp)

	items, ok := payload["data"].([]any)
	if !ok {
		t.Fatalf("campo data não é uma lista válida")
	}

	if len(items) != 1 {
		t.Fatalf("esperado 1 lead filtrado, recebido %d", len(items))
	}
}

func TestUpdateLeadRejectsUnknownField(t *testing.T) {
	app := setupTestApp(t)

	createBody := []byte(`{
		"name":"Nicolas",
		"email":"nicolas@email.com",
		"phone":"11999999999",
		"source":"landing_page"
	}`)

	createResp := performRequest(t, app, http.MethodPost, "/leads", createBody)
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("criação falhou com status %d", createResp.StatusCode)
	}

	updateBody := []byte(`{
		"name":"Nicolas Santos",
		"phone":"11988887777",
		"source":"google_ads",
		"status":"contacted",
		"teste":"valor-invalido"
	}`)

	resp := performRequest(t, app, http.MethodPut, "/leads/1", updateBody)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status esperado %d, recebido %d", http.StatusBadRequest, resp.StatusCode)
	}

	payload := decodeJSONResponse(t, resp)
	errorField, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("campo error ausente ou inválido")
	}

	if errorField["message"] != `campo não permitido: "teste"` {
		t.Fatalf("mensagem inesperada: %v", errorField["message"])
	}
}

func TestListLeadsPagination(t *testing.T) {
	app := setupTestApp(t)

	for i := 1; i <= 3; i++ {
		body := []byte(fmt.Sprintf(`{
			"name":"Lead %d",
			"email":"lead%d@email.com",
			"phone":"11999999999",
			"source":"landing_page"
		}`, i, i))

		resp := performRequest(t, app, http.MethodPost, "/leads", body)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("criação falhou com status %d", resp.StatusCode)
		}
	}

	resp := performRequest(t, app, http.MethodGet, "/leads?page=1&limit=2", nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status esperado %d, recebido %d", http.StatusOK, resp.StatusCode)
	}

	payload := decodeJSONResponse(t, resp)

	items, ok := payload["data"].([]any)
	if !ok {
		t.Fatalf("campo data não é uma lista válida")
	}

	if len(items) != 2 {
		t.Fatalf("esperado 2 itens na página, recebido %d", len(items))
	}
}