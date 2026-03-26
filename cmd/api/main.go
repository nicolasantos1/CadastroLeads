package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/nicolasantos1/CadastroLeads/internal/database"
)

func main() {
	db, err := database.ConnectSQLite()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := fiber.New()

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"data": "API rodando com banco",
		})
	})

	log.Fatal(app.Listen(":3000"))
}