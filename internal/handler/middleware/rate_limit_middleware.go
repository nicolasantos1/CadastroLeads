package middleware

import (
    "fmt"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/gofiber/fiber/v3"
)

type rateLimitEntry struct {
    Count     int
    ExpiresAt time.Time
}

func RequireRateLimit() fiber.Handler {
    maxRequests := getEnvInt("RATE_LIMIT_MAX", 100)
    windowSeconds := getEnvInt("RATE_LIMIT_WINDOW_SECONDS", 60)

    if maxRequests <= 0 {
        maxRequests = 100
    }
    if windowSeconds <= 0 {
        windowSeconds = 60
    }

    window := time.Duration(windowSeconds) * time.Second

    var (
        mu      sync.Mutex
        clients = make(map[string]rateLimitEntry)
    )

    return func(c fiber.Ctx) error {
        key := clientIP(c)
        now := time.Now().UTC()

        mu.Lock()
        entry, found := clients[key]
        if !found || now.After(entry.ExpiresAt) {
            entry = rateLimitEntry{
                Count:     0,
                ExpiresAt: now.Add(window),
            }
        }

        entry.Count++
        clients[key] = entry

        remaining := maxRequests - entry.Count
        if remaining < 0 {
            remaining = 0
        }

        retryAfter := int(time.Until(entry.ExpiresAt).Seconds())
        if retryAfter < 0 {
            retryAfter = 0
        }
        resetAt := entry.ExpiresAt
        exceeded := entry.Count > maxRequests
        if exceeded {
            mu.Unlock()
            c.Set("X-RateLimit-Limit", strconv.Itoa(maxRequests))
            c.Set("X-RateLimit-Remaining", "0")
            c.Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))
            c.Set("Retry-After", strconv.Itoa(retryAfter))
            return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
                "error": fiber.Map{
                    "message": fmt.Sprintf("limite de requisições excedido. Tente novamente em %d segundos", retryAfter),
                },
            })
        }
        mu.Unlock()

        c.Set("X-RateLimit-Limit", strconv.Itoa(maxRequests))
        c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
        c.Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

        return c.Next()
    }
}

func getEnvInt(key string, fallback int) int {
    value := strings.TrimSpace(os.Getenv(key))
    if value == "" {
        return fallback
    }

    parsed, err := strconv.Atoi(value)
    if err != nil {
        return fallback
    }

    return parsed
}

func clientIP(c fiber.Ctx) string {
    forwardedFor := strings.TrimSpace(c.Get("X-Forwarded-For"))
    if forwardedFor != "" {
        parts := strings.Split(forwardedFor, ",")
        if len(parts) > 0 {
            return strings.TrimSpace(parts[0])
        }
    }

    realIP := strings.TrimSpace(c.Get("X-Real-IP"))
    if realIP != "" {
        return realIP
    }

    ip := strings.TrimSpace(c.IP())
    if ip == "" {
        return "unknown"
    }

    return ip
}
