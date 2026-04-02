# CadastroLeads

API REST para cadastro e gestão de leads, desenvolvida em **Go 1.26**, utilizando **Fiber v3** e **SQLite**.

## Objetivo

Este projeto foi desenvolvido para gerenciar leads com operações de:

- criar
- listar
- buscar por ID
- atualizar
- atualizar apenas o status
- remover

A aplicação segue uma estrutura em camadas, separando responsabilidades entre:

- handler
- service
- repository

## Tecnologias utilizadas

- Go 1.26
- Fiber v3
- SQLite
- modernc.org/sqlite
- Swagger/OpenAPI
- Docker
- Docker Compose

## Funcionalidades implementadas

- CRUD completo de leads
- validação de campos obrigatórios
- validação de e-mail
- bloqueio de e-mail duplicado
- paginação na listagem
- filtro por `status`
- filtro por `source`
- status inicial automático como `new`
- autenticação por token Bearer nas rotas administrativas
- soft delete
- middleware de logging
- documentação Swagger
- suporte a variáveis de ambiente
- suporte a Docker e Docker Compose
- testes automatizados

## Estrutura do projeto

```text
CadastroLeads/
├── cmd/
│   └── api/
│       └── main.go
├── docs/
├── internal/
│   ├── database/
│   │   ├── migrations/
│   │   └── seeds/
│   ├── dto/
│   ├── handler/
│   │   └── middleware/
│   ├── model/
│   ├── repository/
│   └── service/
├── tests/
├── .env.example
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

## Regras de negócio implementadas

- `name`, `email` e `source` são obrigatórios na criação
- o e-mail deve ter formato válido
- não é permitido cadastrar dois leads com o mesmo e-mail
- o status inicial de um lead criado é `new`
- os status aceitos são:
  - `new`
  - `contacted`
  - `qualified`
  - `lost`
- a listagem suporta paginação
- a listagem suporta filtro por `status`
- a listagem suporta filtro por `source`
- não permite atualização de campos inexistentes
- as respostas seguem padrão JSON
- o delete é lógico, usando soft delete

## Variáveis de ambiente

Exemplo no arquivo `.env.example`:

```env
PORT=3000
DB_PATH=leads.db
API_TOKEN=dev-token-123
```

### Descrição

- `PORT`: porta da aplicação
- `DB_PATH`: caminho do arquivo SQLite
- `API_TOKEN`: token fixo usado na autenticação Bearer

## Como rodar o projeto localmente

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

## Como rodar com Docker

### Build da imagem

```bash
docker build -t cadastroleads .
```

### Rodar com Docker

```bash
docker run -p 3000:3000 cadastroleads
```

## Como rodar com Docker Compose

```bash
docker compose up --build
```

A API ficará disponível em:

```text
http://localhost:3000
```

## Autenticação

As rotas de leads exigem o header:

```text
Authorization: Bearer <token>
```

Exemplo usando o token padrão:

```text
Authorization: Bearer dev-token-123
```

## Swagger

A documentação Swagger fica disponível em:

```text
http://localhost:3000/swagger/index.html
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

### Headers

```text
Authorization: Bearer dev-token-123
Content-Type: application/json
```

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

- `page`
- `limit`
- `status`
- `source`

### Exemplo

```text
GET /leads?page=1&limit=10&status=qualified&source=google_ads
```

---

## 3. Buscar lead por ID

### GET /leads/:id

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

---

## 5. Atualizar apenas o status

### PATCH /leads/:id/status

### Body

```json
{
  "status": "qualified"
}
```

---

## 6. Remover lead

### DELETE /leads/:id

A remoção é lógica, ou seja, o registro não é apagado fisicamente do banco. O campo `deleted_at` é preenchido e o lead deixa de aparecer nas consultas normais.

---

## Seeds de dados

O projeto possui arquivo SQL de seed em:

```text
internal/database/seeds/000001_seed_leads.sql
```

Esse arquivo contém leads de exemplo para facilitar testes locais.

> Se a seed estiver conectada no startup da aplicação, os dados serão inseridos automaticamente.
> Caso não esteja, o arquivo ainda pode ser usado manualmente no banco SQLite.

---

## Como rodar os testes

```bash
go test ./...
```

Para ver mais detalhes:

```bash
go test ./... -v
```

## Testes implementados

Atualmente o projeto possui testes cobrindo cenários como:

- criação com sucesso
- e-mail duplicado
- status inválido
- filtro por status
- filtro por source
- paginação
- exclusão e busca posterior por ID inexistente
- rejeição de campo desconhecido
- autenticação nas rotas protegidas

## Exemplos de uso no PowerShell

### Criar lead

```powershell
Invoke-RestMethod -Method POST -Uri "http://localhost:3000/leads" `
  -Headers @{ Authorization = "Bearer dev-token-123" } `
  -ContentType "application/json" `
  -Body '{"name":"Nicolas","email":"nicolas@email.com","phone":"11999999999","source":"landing_page"}'
```

### Listar leads

```powershell
Invoke-RestMethod -Method GET -Uri "http://localhost:3000/leads" `
  -Headers @{ Authorization = "Bearer dev-token-123" }
```

### Buscar por ID

```powershell
Invoke-RestMethod -Method GET -Uri "http://localhost:3000/leads/1" `
  -Headers @{ Authorization = "Bearer dev-token-123" }
```

### Atualizar lead

```powershell
Invoke-RestMethod -Method PUT -Uri "http://localhost:3000/leads/1" `
  -Headers @{ Authorization = "Bearer dev-token-123" } `
  -ContentType "application/json" `
  -Body '{"name":"Nicolas Santos","phone":"11988887777","source":"google_ads","status":"contacted"}'
```

### Atualizar status

```powershell
Invoke-RestMethod -Method PATCH -Uri "http://localhost:3000/leads/1/status" `
  -Headers @{ Authorization = "Bearer dev-token-123" } `
  -ContentType "application/json" `
  -Body '{"status":"qualified"}'
```

### Deletar lead

```powershell
Invoke-RestMethod -Method DELETE -Uri "http://localhost:3000/leads/1" `
  -Headers @{ Authorization = "Bearer dev-token-123" }
```

## Status HTTP utilizados

- `200 OK`
- `201 Created`
- `400 Bad Request`
- `401 Unauthorized`
- `404 Not Found`
- `409 Conflict`
- `500 Internal Server Error`

## Autor

Nicolas Santos do Nascimento