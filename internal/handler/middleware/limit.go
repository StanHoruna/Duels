package middleware

import (
	"duels-api/pkg/middleware/limiter"
	"github.com/gofiber/fiber/v3"
	"time"
)

func NewLimiter() fiber.Handler {
	cfg := limiter.Config{
		Max:        1,
		Expiration: 1 * time.Second,
		Skip: func(c fiber.Ctx) bool {
			return c.Response().StatusCode() == fiber.StatusUnauthorized
		},
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP() + ";" + c.Path()
		},
	}

	return limiter.NewSlidingWindow(cfg)
}
