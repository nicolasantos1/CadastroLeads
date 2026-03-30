package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"

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

	app := fiber.New()

	app.Use(recoverer.New())
	app.Use(logger.New(logger.Config{
		Format: "${time} | ${status} | ${latency} | ${method} ${path}\n",
	}))

	leadRepo := repository.NewLeadRepository(db)
	leadService := service.NewLeadService(leadRepo)
	leadHandler := handler.NewLeadHandler(leadService)

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"data": "API rodando com banco",
		})
	})

	// rota temporária para testar o recover
	app.Get("/panic-test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"error": "SDF",
		})
	})
	
	leadHandler.RegisterRoutes(app)

	log.Fatal(app.Listen(":" + port))
}