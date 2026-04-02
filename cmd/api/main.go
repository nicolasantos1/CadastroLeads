package main

// @title CadastroLeads API
// @version 1.0
// @description API REST para cadastro e gestão de leads.
// @host localhost:3000
// @BasePath /

import (
	"log"
	"os"

	swaggo "github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"

	_ "github.com/nicolasantos1/CadastroLeads/docs"

	"github.com/nicolasantos1/CadastroLeads/internal/database"
	"github.com/nicolasantos1/CadastroLeads/internal/handler"
	"github.com/nicolasantos1/CadastroLeads/internal/repository"
	"github.com/nicolasantos1/CadastroLeads/internal/service"
)

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func main() {
	
	app := fiber.New()

	app.Use(logger.New(logger.Config{
		Format: "${time} | ${status} | ${latency} | ${method} ${path}\n",
	}))
	app.Use(recoverer.New())

	
	app.Get("/swagger/*", swaggo.HandlerDefault)

	port := getEnv("PORT", "3000")
	dbPath := getEnv("DB_PATH", "leads.db")

	db, err := database.ConnectSQLite(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("erro ao fechar banco: %v", err)
		}
	}()
	
	if err := database.SeedFromFiles(db, true); err != nil {
   		log.Fatalf("erro ao rodar seeds: %v", err)
	}

	leadRepo := repository.NewLeadRepository(db)
	leadService := service.NewLeadService(leadRepo)
	leadHandler := handler.NewLeadHandler(leadService)

	// Health godoc
	// @Summary Verifica saúde da API
	// @Description Retorna status de saúde da API
	// @Tags health
	// @Produce json
	// @Success 200 {object} dto.HealthResponse
	// @Router /health [get]
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"data": "API rodando com banco",
		})
	})
	
	leadHandler.RegisterRoutes(app)

	log.Fatal(app.Listen(":" + port))
}

