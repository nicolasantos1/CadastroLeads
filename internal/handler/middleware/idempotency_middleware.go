package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/nicolasantos1/CadastroLeads/internal/model"
	"github.com/nicolasantos1/CadastroLeads/internal/repository"
)

func RequireIdempotency(repo repository.IdempotencyRepository) fiber.Handler {
	return func(c fiber.Ctx) error {
		key := strings.TrimSpace(c.Get("Idempotency-Key"))
		if key == "" {
			return c.Next()
		}

		method := strings.TrimSpace(c.Method())
		path := strings.TrimSpace(c.Path())
		requestHash := sha256Hex(c.Body())

		record, err := repo.Get(key, method, path)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "erro ao verificar idempotência",
				},
			})
		}

		if record != nil {
			if record.RequestHash != requestHash {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": fiber.Map{
						"message": "Idempotency-Key já foi usada com outro payload",
					},
				})
			}

			if record.Status == model.IdempotencyStatusCompleted {
				c.Set("X-Idempotency-Replayed", "true")
				c.Set("Content-Type", "application/json")
				return c.Status(record.ResponseStatusCode).SendString(record.ResponseBody)
			}

			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "requisição idempotente já está em processamento",
				},
			})
		}

		reserved, err := repo.Reserve(key, method, path, requestHash)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "erro ao reservar chave idempotente",
				},
			})
		}

		if !reserved {
			record, err = repo.Get(key, method, path)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fiber.Map{
						"message": "erro ao verificar idempotência",
					},
				})
			}

			if record != nil {
				if record.RequestHash != requestHash {
					return c.Status(fiber.StatusConflict).JSON(fiber.Map{
						"error": fiber.Map{
							"message": "Idempotency-Key já foi usada com outro payload",
						},
					})
				}

				if record.Status == model.IdempotencyStatusCompleted {
					c.Set("X-Idempotency-Replayed", "true")
					c.Set("Content-Type", "application/json")
					return c.Status(record.ResponseStatusCode).SendString(record.ResponseBody)
				}

				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": fiber.Map{
						"message": "requisição idempotente já está em processamento",
					},
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "erro ao processar idempotência",
				},
			})
		}

		if err := c.Next(); err != nil {
			return err
		}

		if err := repo.Complete(key, method, path, c.Response().StatusCode(), string(c.Response().Body())); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "erro ao persistir resultado idempotente",
				},
			})
		}

		return nil
	}
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}