package server

import (
	"duels-api/config"
	"duels-api/internal/handler/middleware"
	"encoding/json"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
	"go.uber.org/zap"
	"strings"
	"time"
)

func NewServer(c *config.Config, l *zap.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		ProxyHeader: "X-Forwarded-For",
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	corsConfig := cors.Config{
		AllowOrigins: strings.Split(c.HTTP.AllowOrigins, ","),
		AllowMethods: []string{"GET", "POST", "HEAD", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-Forwarded-For", "X-CSRF-Token",
			"Authorization", "User-Env", "Access-Control-Request-Headers", "Access-Control-Allow-Headers",
			"Access-Control-Request-Method", "Content-Unique-Identifier", "Content-Index",
			"Access-Control-Allow-BaseError", "Access-Control-Request-Headers", "Access-Control-Request-Method",
		},
		AllowCredentials: c.HTTP.AllowCredentials,
		MaxAge:           int(12 * time.Hour),
	}

	app.Use(cors.New(corsConfig))

	logMiddleware := middleware.NewLoggingMiddleware(l)
	logMiddleware.RegisterLogger(c, app)
	app.Use(recoverer.New())

	app.Get("/", healthcheck.New(healthcheck.Config{
		Probe: func(c fiber.Ctx) bool {
			_, err := c.WriteString("This is new service")
			if err != nil {
				return false
			}

			return true
		},
	}))

	app.Get("/static/*", static.New("./resources/static"))
	app.Get("/media/*", static.New("./storage"))

	return app
}
