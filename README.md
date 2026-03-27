# CadastroLeads

API REST para cadastro e gestão de leads, desenvolvida em **Go 1.26**, utilizando **Fiber v3** e **SQLite**.

## Objetivo

Este projeto foi desenvolvido para gerenciar leads com operações de:

* criar
* listar
* buscar por ID
* atualizar
* atualizar apenas o status
* remover

A aplicação segue uma estrutura em camadas, separando responsabilidades entre:

* handler
* service
* repository

## Tecnologias utilizadas

* Go 1.26
* Fiber v3
* SQLite
* modernc.org/sqlite

## Estrutura do projeto

```text
CadastroLeads/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── database/
│   │   └── migrations/
│   ├── dto/
│   ├── handler/
│   ├── model/
│   ├── repository/
│   └── service/
├── tests/
├── go.mod
├── go.sum
└── README.md
```

## Regras de negócio implementadas

* `name`, `email` e `source` são obrigatórios na criação
* o e-mail deve ter formato válido
* não é permitido cadastrar dois leads com o mesmo e-mail
* o status inicial de um lead criado é `new`
* os status aceitos são:

  * `new`
  * `contacted`
  * `qualified`
  * `lost`
* a listagem suporta paginação
* a listagem suporta filtro por `status`
* a listagem suporta filtro por `source`

## Como rodar o projeto

### 1. Clonar o repositório

```bash
git clone https://github.com/nicolasantos1/CadastroLeads.git
cd CadastroLeads
```

### 2. Instalar as dependências

```bash
go mod tidy
```

### 3. Rodar a aplicação

```bash
go run ./cmd/api
```

A API ficará disponível em:

```text
http://localhost:3000
```

## Como rodar os testes

```bash
go test ./...
```

## Endpoint de verificação

### GET /health

Resposta esperada:

```json
{
  "data": "API rodando com banco"
}
```

---

# Endpoints

## 1. Criar lead

### POST /leads

### Body

```json
{
  "name": "Nicolas",
  "email": "nicolas@email.com",
  "phone": "11999999999",
  "source": "landing_page"
}
```

### Resposta de sucesso

```json
{
  "data": {
    "id": 1,
    "name": "Nicolas",
    "email": "nicolas@email.com",
    "phone": "11999999999",
    "source": "landing_page",
    "status": "new",
    "created_at": "2026-03-27T14:47:25Z",
    "updated_at": "2026-03-27T14:47:25Z"
  }
}
```

### Possíveis erros

```json
{
  "error": {
    "message": "name é obrigatório"
  }
}
```

```json
{
  "error": {
    "message": "email inválido"
  }
}
```

```json
{
  "error": {
    "message": "source é obrigatório"
  }
}
```

```json
{
  "error": {
    "message": "já existe um lead com este email"
  }
}
```

---

## 2. Listar leads

### GET /leads

### Query params opcionais

* `page`
* `limit`
* `status`
* `source`

### Exemplo

```text
GET /leads?page=1&limit=10&status=qualified&source=google_ads
```

### Resposta de sucesso

```json
{
  "data": [
    {
      "id": 1,
      "name": "Nicolas",
      "email": "nicolas@email.com",
      "phone": "11999999999",
      "source": "google_ads",
      "status": "qualified",
      "created_at": "2026-03-27T14:47:25Z",
      "updated_at": "2026-03-27T14:50:10Z"
    }
  ]
}
```

---

## 3. Buscar lead por ID

### GET /leads/:id

### Exemplo

```text
GET /leads/1
```

### Resposta de sucesso

```json
{
  "data": {
    "id": 1,
    "name": "Nicolas",
    "email": "nicolas@email.com",
    "phone": "11999999999",
    "source": "landing_page",
    "status": "new",
    "created_at": "2026-03-27T14:47:25Z",
    "updated_at": "2026-03-27T14:47:25Z"
  }
}
```

### Erro se não encontrar

```json
{
  "error": {
    "message": "lead not found"
  }
}
```

---

## 4. Atualizar lead completo

### PUT /leads/:id

### Body

```json
{
  "name": "Nicolas Santos",
  "phone": "11988887777",
  "source": "google_ads",
  "status": "contacted"
}
```

### Resposta de sucesso

```json
{
  "data": {
    "id": 1,
    "name": "Nicolas Santos",
    "email": "nicolas@email.com",
    "phone": "11988887777",
    "source": "google_ads",
    "status": "contacted",
    "created_at": "2026-03-27T14:47:25Z",
    "updated_at": "2026-03-27T14:55:00Z"
  }
}
```

---

## 5. Atualizar apenas o status

### PATCH /leads/:id/status

### Body

```json
{
  "status": "qualified"
}
```

### Resposta de sucesso

```json
{
  "data": {
    "id": 1,
    "name": "Nicolas Santos",
    "email": "nicolas@email.com",
    "phone": "11988887777",
    "source": "google_ads",
    "status": "qualified",
    "created_at": "2026-03-27T14:47:25Z",
    "updated_at": "2026-03-27T14:57:00Z"
  }
}
```

### Erro para status inválido

```json
{
  "error": {
    "message": "status inválido"
  }
}
```

---

## 6. Remover lead

### DELETE /leads/:id

### Resposta de sucesso

```json
{
  "data": {
    "message": "lead removido com sucesso"
  }
}
```

---

# Exemplos de uso no PowerShell

## Criar lead

```powershell
Invoke-RestMethod -Method POST -Uri "http://localhost:3000/leads" `
  -ContentType "application/json" `
  -Body '{"name":"Nicolas","email":"nicolas@email.com","phone":"11999999999","source":"landing_page"}'
```

## Listar leads

```powershell
Invoke-RestMethod -Method GET -Uri "http://localhost:3000/leads"
```

## Buscar por ID

```powershell
Invoke-RestMethod -Method GET -Uri "http://localhost:3000/leads/1"
```

## Atualizar lead

```powershell
Invoke-RestMethod -Method PUT -Uri "http://localhost:3000/leads/1" `
  -ContentType "application/json" `
  -Body '{"name":"Nicolas Santos","phone":"11988887777","source":"google_ads","status":"contacted"}'
```

## Atualizar status

```powershell
Invoke-RestMethod -Method PATCH -Uri "http://localhost:3000/leads/1/status" `
  -ContentType "application/json" `
  -Body '{"status":"qualified"}'
```

## Deletar lead

```powershell
Invoke-RestMethod -Method DELETE -Uri "http://localhost:3000/leads/1"
```

---

## Status HTTP utilizados

* `200 OK`
* `201 Created`
* `400 Bad Request`
* `404 Not Found`
* `409 Conflict`
* `500 Internal Server Error`

## Testes implementados

Atualmente o projeto possui testes cobrindo cenários como:

* criação com sucesso
* e-mail duplicado
* status inválido
* filtro por status
* exclusão e busca posterior por ID inexistente

## Melhorias futuras

* Docker
* Swagger/OpenAPI
* logging
* variáveis de ambiente
* autenticação para rotas administrativas

## Autor

Nicolas Santos do Nascimento

