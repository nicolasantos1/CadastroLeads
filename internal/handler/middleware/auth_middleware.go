package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func RequireAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := strings.TrimSpace(c.Get("Authorization"))

		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "header Authorization é obrigatório",
				},
			})
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "formato inválido. Use: Bearer <token>",
				},
			})
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "token não informado",
				},
			})
		}

		expectedToken := os.Getenv("API_TOKEN")
		if expectedToken == "" {
			expectedToken = "dev-token-123"
		}

		if token != expectedToken {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "token inválido",
				},
			})
		}

		return c.Next()
	}
}