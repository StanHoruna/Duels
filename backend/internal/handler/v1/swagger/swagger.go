package swagger

import (
	_ "embed"
	"net/url"
	"strings"

	"duels-api/config"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Handler struct {
	ValidatorBase string
}

func NewSwaggerHandler(c *config.Config) *Handler {
	return &Handler{
		ValidatorBase: strings.TrimSpace(c.HTTP.SwaggerValidatorURL),
	}
}

//go:embed swagger.json
var swaggerJSON []byte

func (h *Handler) RegisterRoutes(app *fiber.App) {

	app.Get("/swagger.json", func(c fiber.Ctx) error {
		c.Set("Content-Type", "application/json; charset=utf-8")
		return c.Send(swaggerJSON)
	})

	localHTTP := httpSwagger.Handler(
		httpSwagger.URL("/swagger.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
	)
	localUI := adaptor.HTTPHandler(localHTTP)

	isHTTPS := func(c fiber.Ctx) bool {
		if strings.EqualFold(c.Protocol(), "https") {
			return true
		}
		if strings.EqualFold(c.Get("X-Forwarded-Proto"), "https") {
			return true
		}
		return false
	}

	serveDocs := func(c fiber.Ctx) error {
		if isHTTPS(c) && h.ValidatorBase != "" {
			jsonURL := url.URL{
				Scheme: "https",
				Host:   c.Hostname(),
				Path:   "/swagger.json",
			}
			return c.Redirect().To(h.ValidatorBase + url.QueryEscape(jsonURL.String()))
		}
		return localUI(c)
	}

	app.Get("/docs", serveDocs)
	app.Get("/docs/*", serveDocs)
}
