package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/nicolasantos1/CadastroLeads/internal/database"
	"github.com/nicolasantos1/CadastroLeads/internal/handler"
	"github.com/nicolasantos1/CadastroLeads/internal/repository"
	"github.com/nicolasantos1/CadastroLeads/internal/service"
)

func main() {
	db, err := database.ConnectSQLite()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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

	log.Fatal(app.Listen(":3000"))
}